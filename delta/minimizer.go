// Copyright 2025 Sonic Labs
// This file is part of Aida Testing Infrastructure for Sonic
//
// Aida is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// Aida is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with Aida. If not, see <http://www.gnu.org/licenses/>.

package delta

import (
	"context"
	"fmt"
	"math/rand"
	"sort"
	"time"

	"github.com/ethereum/go-ethereum/common"
)

// Outcome describes the observed result when executing a candidate trace.
type Outcome int

const (
	OutcomePass Outcome = iota
	OutcomeFail
	OutcomeUnresolved
)

func (o Outcome) String() string {
	switch o {
	case OutcomePass:
		return "pass"
	case OutcomeFail:
		return "fail"
	case OutcomeUnresolved:
		return "unresolved"
	default:
		return fmt.Sprintf("unknown(%d)", int(o))
	}
}

// TestFunc evaluates a candidate trace and reports the observed outcome.
type TestFunc func(ctx context.Context, ops []TraceOp) (Outcome, error)

// OperationMeta captures properties needed for minimisation.
type OperationMeta struct {
	Op          TraceOp
	Kind        string
	Mandatory   bool
	HasContract bool
	Contract    common.Address
	Index       int
}

// MinimizerConfig customizes the minimisation process.
type MinimizerConfig struct {
	AddressSampleRuns int   // number of attempts per factor when sampling addresses
	RandSeed          int64 // RNG seed (<=0 uses time-based seed)
	MaxFactor         int   // upper bound for sampling factor, 0 defaults to len(addresses)
	MandatoryKinds    map[string]struct{}
	Logger            func(format string, args ...any)
}

// Minimizer orchestrates range and address reduction for traces.
type Minimizer struct {
	cfg  MinimizerConfig
	rand *rand.Rand
}

// Minimize reduces the trace while maintaining the failure outcome.
func (m *Minimizer) Minimize(ctx context.Context, ops []TraceOp, test TestFunc) ([]TraceOp, error) {
	if test == nil {
		return nil, fmt.Errorf("delta: test function must be provided")
	}
	if len(ops) == 0 {
		return nil, fmt.Errorf("delta: trace is empty")
	}

	meta := collectMetadata(ops, m.cfg.MandatoryKinds)

	// strip leading operations using binary search.
	reducedOps, reducedMeta, err := m.reducePrefix(ctx, meta, test)
	if err != nil {
		return nil, err
	}

	// probabilistic address-based reduction.
	addressReduced, _, err := m.reduceAddresses(ctx, reducedOps, reducedMeta, test)
	if err != nil {
		return nil, err
	}

	return addressReduced, nil
}

func (m *Minimizer) reducePrefix(ctx context.Context, meta []OperationMeta, test TestFunc) ([]TraceOp, []OperationMeta, error) {
	if len(meta) == 0 {
		return nil, nil, fmt.Errorf("delta: empty metadata")
	}

	full := buildOperations(meta, func(OperationMeta) bool { return true })
	outcome, err := test(ctx, full)
	if err != nil {
		return nil, nil, err
	}
	if outcome != OutcomeFail {
		return nil, nil, ErrInputDoesNotFail
	}

	bestOps := full
	bestMeta := collectMetadata(bestOps, m.cfg.MandatoryKinds)
	lo := 0
	hi := len(meta)

	for hi-lo > 1 {
		if err := ctx.Err(); err != nil {
			return nil, nil, err
		}

		mid := (lo + hi) / 2
		candidate := buildOperations(meta, func(m OperationMeta) bool {
			return m.Index >= mid
		})

		outcome, err := test(ctx, candidate)
		if err != nil {
			return nil, nil, err
		}

		if outcome == OutcomeFail {
			lo = mid
			bestOps = candidate
			bestMeta = collectMetadata(bestOps, m.cfg.MandatoryKinds)
			m.log("prefix reduction accepted: start=%d", mid)
		} else {
			hi = mid
			m.log("prefix reduction rejected: start=%d", mid)
		}
	}

	return bestOps, bestMeta, nil
}

func (m *Minimizer) reduceAddresses(
	ctx context.Context,
	ops []TraceOp,
	meta []OperationMeta,
	test TestFunc,
) ([]TraceOp, []OperationMeta, error) {
	addresses := uniqueContracts(meta)
	if len(addresses) <= 1 {
		return ops, meta, nil
	}

	maxFactor := m.cfg.MaxFactor
	if maxFactor <= 0 || maxFactor > len(addresses) {
		maxFactor = len(addresses)
	}

	currentOps := ops
	currentMeta := meta

	for factor := 2; factor <= maxFactor && len(addresses) > 1; {
		if err := ctx.Err(); err != nil {
			return nil, nil, err
		}

		sampleSize := len(addresses) / factor
		if sampleSize == 0 {
			factor++
			continue
		}
		if sampleSize >= len(addresses) {
			factor++
			continue
		}

		m.log("address reduction: factor=1/%d, sample=%d, addresses=%d", factor, sampleSize, len(addresses))

		reduced := false
		for attempt := 0; attempt < m.cfg.AddressSampleRuns && len(addresses) > 1; attempt++ {
			if err := ctx.Err(); err != nil {
				return nil, nil, err
			}

			exclude := m.sampleAddresses(addresses, sampleSize)
			excludeSet := make(map[common.Address]struct{}, len(exclude))
			for _, addr := range exclude {
				excludeSet[addr] = struct{}{}
			}

			candidate := buildOperations(currentMeta, func(meta OperationMeta) bool {
				if !meta.HasContract {
					return true
				}
				_, skip := excludeSet[meta.Contract]
				return !skip
			})

			outcome, err := test(ctx, candidate)
			if err != nil {
				return nil, nil, err
			}

			if outcome == OutcomeFail {
				currentOps = candidate
				currentMeta = collectMetadata(currentOps, m.cfg.MandatoryKinds)
				addresses = uniqueContracts(currentMeta)
				reduced = true
				m.log("address reduction accepted: removed=%d", len(exclude))
				break
			}
			m.log("address reduction rejected: attempt=%d", attempt+1)
		}

		if !reduced {
			factor++
		}
	}

	return currentOps, currentMeta, nil
}

func (m *Minimizer) sampleAddresses(addresses []common.Address, sampleSize int) []common.Address {
	if sampleSize <= 0 || len(addresses) == 0 {
		return nil
	}
	if sampleSize > len(addresses) {
		sampleSize = len(addresses)
	}

	idx := m.rand.Perm(len(addresses))
	chosen := make([]common.Address, 0, sampleSize)
	for i := 0; i < sampleSize; i++ {
		chosen = append(chosen, addresses[idx[i]])
	}
	return chosen
}

func (m *Minimizer) log(format string, args ...any) {
	if m.cfg.Logger != nil {
		m.cfg.Logger(format, args...)
	}
}

type metaCollector struct {
	prevContract common.Address
	mandatory    map[string]struct{}
}

func collectMetadata(ops []TraceOp, mandatoryKinds map[string]struct{}) []OperationMeta {
	collector := metaCollector{
		prevContract: common.Address{},
		mandatory:    mandatoryKinds,
	}
	meta := make([]OperationMeta, 0, len(ops))
	for idx, op := range ops {
		entry := collector.collect(op)
		entry.Index = idx
		meta = append(meta, entry)
	}
	return meta
}

func (c *metaCollector) collect(op TraceOp) OperationMeta {
	mandatory := false
	if c.mandatory != nil {
		_, mandatory = c.mandatory[op.Kind]
	}

	contract, has := c.contractFor(op)
	if has {
		c.prevContract = contract
	}

	return OperationMeta{
		Op:          op,
		Kind:        op.Kind,
		Mandatory:   mandatory,
		HasContract: has,
		Contract:    contract,
	}
}

func (c *metaCollector) contractFor(op TraceOp) (common.Address, bool) {
	if op.HasContract {
		return op.Contract, true
	}
	if _, ok := inheritContractKinds[op.Kind]; ok {
		return c.prevContract, true
	}
	return common.Address{}, false
}

func buildOperations(meta []OperationMeta, include func(OperationMeta) bool) []TraceOp {
	result := make([]TraceOp, 0, len(meta))
	for _, m := range meta {
		if include(m) || m.Mandatory {
			result = append(result, m.Op)
		}
	}
	return result
}

func uniqueContracts(meta []OperationMeta) []common.Address {
	set := make(map[common.Address]struct{})
	for _, m := range meta {
		if !m.HasContract {
			continue
		}
		set[m.Contract] = struct{}{}
	}
	addrs := make([]common.Address, 0, len(set))
	for addr := range set {
		addrs = append(addrs, addr)
	}
	sort.Slice(addrs, func(i, j int) bool { return addrs[i].Hex() < addrs[j].Hex() })
	return addrs
}

// NewMinimizer creates a minimizer with the provided configuration.
func NewMinimizer(cfg MinimizerConfig) *Minimizer {
	if cfg.AddressSampleRuns <= 0 {
		cfg.AddressSampleRuns = 5
	}
	if cfg.MandatoryKinds == nil {
		cfg.MandatoryKinds = defaultMandatoryKinds()
	}
	seed := cfg.RandSeed
	if seed == 0 {
		seed = time.Now().UnixNano()
	}
	return &Minimizer{
		cfg:  cfg,
		rand: rand.New(rand.NewSource(seed)),
	}
}

func defaultMandatoryKinds() map[string]struct{} {
	return map[string]struct{}{
		"BeginBlock":       {},
		"EndBlock":         {},
		"BeginTransaction": {},
		"EndTransaction":   {},
		"Snapshot":         {},
		"RevertToSnapshot": {},
	}
}

var inheritContractKinds = map[string]struct{}{
	"SetStateLcls":                  {},
	"SetTransientStateLcls":         {},
	"GetTransientStateLccs":         {},
	"GetTransientStateLc":           {},
	"GetStateAndCommittedStateLcls": {},
	"GetStateLcls":                  {},
	"GetStateLccs":                  {},
	"GetStateLc":                    {},
	"GetCommittedStateLcls":         {},
	"GetCodeHashLc":                 {},
	"GetTransientStateLcls":         {},
}

// ErrInputDoesNotFail indicates the original trace did not reproduce the failure.
var ErrInputDoesNotFail = fmt.Errorf("delta: original trace does not reproduce the failure")

// UniqueContracts returns the sorted set of contract addresses referenced by operations.
func UniqueContracts(ops []TraceOp) []common.Address {
	meta := collectMetadata(ops, defaultMandatoryKinds())
	return uniqueContracts(meta)
}
