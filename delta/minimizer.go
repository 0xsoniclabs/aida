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

// outcome describes the observed result when executing a candidate trace.
type outcome int

const (
	outcomePass outcome = iota
	outcomeFail
	outcomeUnresolved
)

func (o outcome) String() string {
	switch o {
	case outcomePass:
		return "pass"
	case outcomeFail:
		return "fail"
	case outcomeUnresolved:
		return "unresolved"
	default:
		return fmt.Sprintf("unknown(%d)", int(o))
	}
}

// testFunc evaluates a candidate trace and reports the observed outcome.
type testFunc func(ctx context.Context, ops []TraceOp) (outcome, error)

// operationMeta captures properties needed for minimisation.
type operationMeta struct {
	Op          TraceOp
	Kind        string
	Mandatory   bool
	HasContract bool
	Contract    common.Address
	Index       int
}

// MinimizerConfig customizes the minimisation process.
type MinimizerConfig struct {
	AddressSampleRuns int   // number of attempts when sampling addresses for elimination
	RandSeed          int64 // RNG seed (<=0 uses time-based seed)
	MaxFactor         int   // optional upper bound for sampled address-set size
	MandatoryKinds    map[string]struct{}
	Logger            func(format string, args ...any)
}

// Minimizer orchestrates multi-strategy trace minimisation.
type Minimizer struct {
	cfg  MinimizerConfig
	rand *rand.Rand
}

type scopeNode struct {
	kind     string
	start    int
	end      int
	children []*scopeNode
	leaves   []int
}

var scopeBeginToEnd = map[string]string{
	"BeginSyncPeriod":  "EndSyncPeriod",
	"BeginBlock":       "EndBlock",
	"BeginTransaction": "EndTransaction",
}

var scopeEndToBegin = map[string]string{
	"EndSyncPeriod":  "BeginSyncPeriod",
	"EndBlock":       "BeginBlock",
	"EndTransaction": "BeginTransaction",
}

// Minimize reduces the trace while maintaining the failure outcome.
func (m *Minimizer) Minimize(ctx context.Context, ops []TraceOp, test testFunc) ([]TraceOp, error) {
	if test == nil {
		return nil, fmt.Errorf("delta: test function must be provided")
	}
	if len(ops) == 0 {
		return nil, fmt.Errorf("delta: trace is empty")
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	guards := newGuardVector(len(ops))
	fails, err := m.reproducesFailure(ctx, ops, guards, test)
	if err != nil {
		return nil, err
	}
	if !fails {
		return nil, ErrInputDoesNotFail
	}

	scopeForest := buildScopeForest(ops)

	for {
		if err := ctx.Err(); err != nil {
			return nil, err
		}

		startCount := countOnes(guards)

		guards, err = m.structuralHalvening(ctx, ops, guards, test)
		if err != nil {
			return nil, err
		}

		guards, err = m.addressElimination(ctx, ops, guards, test)
		if err != nil {
			return nil, err
		}

		guards, err = m.emptyStructureElimination(ctx, ops, scopeForest, guards, test)
		if err != nil {
			return nil, err
		}

		if countOnes(guards) == startCount {
			break
		}
	}

	return operationsForGuards(ops, guards), nil
}

func (m *Minimizer) structuralHalvening(
	ctx context.Context,
	ops []TraceOp,
	guards []bool,
	test testFunc,
) ([]bool, error) {
	current := copyGuards(guards)

	for {
		if err := ctx.Err(); err != nil {
			return nil, err
		}

		activeNonStructural := enabledNonStructuralIndices(ops, current)
		if len(activeNonStructural) == 0 {
			break
		}

		lo := 0
		hi := len(activeNonStructural) + 1
		best := 0

		for hi-lo > 1 {
			if err := ctx.Err(); err != nil {
				return nil, err
			}

			mid := (lo + hi) / 2
			candidate := removeNonStructuralPrefix(current, activeNonStructural, mid)
			if !isSubset(candidate, current) {
				return nil, fmt.Errorf("delta: structural halvening produced a non-subset candidate")
			}

			fails, err := m.reproducesFailure(ctx, ops, candidate, test)
			if err != nil {
				return nil, err
			}

			if fails {
				best = mid
				lo = mid
			} else {
				hi = mid
			}
		}

		if best == 0 {
			break
		}

		next := removeNonStructuralPrefix(current, activeNonStructural, best)
		removed := countOnes(current) - countOnes(next)
		if removed <= 0 {
			break
		}
		if !isSubset(next, current) {
			return nil, fmt.Errorf("delta: structural halvening produced a non-subset candidate")
		}

		current = next
		m.log("structural halvening accepted: removed=%d", removed)
	}

	return current, nil
}

func (m *Minimizer) addressElimination(
	ctx context.Context,
	ops []TraceOp,
	guards []bool,
	test testFunc,
) ([]bool, error) {
	current := copyGuards(guards)
	sampleSize := 0

	for {
		if err := ctx.Err(); err != nil {
			return nil, err
		}

		meta := collectActiveMetadata(ops, current, m.cfg.MandatoryKinds)
		addresses := activeContracts(meta, current)
		if len(addresses) <= 1 {
			break
		}

		if sampleSize <= 0 {
			sampleSize = m.initialAddressSampleSize(len(addresses))
		}
		if sampleSize >= len(addresses) {
			sampleSize = len(addresses) - 1
		}
		if sampleSize <= 0 {
			break
		}

		startCount := countOnes(current)
		reduced := false

		for attempt := 0; attempt < m.cfg.AddressSampleRuns; attempt++ {
			if err := ctx.Err(); err != nil {
				return nil, err
			}

			excluded := m.sampleAddresses(addresses, sampleSize)
			if len(excluded) == 0 {
				continue
			}

			excludedSet := make(map[common.Address]struct{}, len(excluded))
			for _, addr := range excluded {
				excludedSet[addr] = struct{}{}
			}

			candidate := disableContracts(current, meta, excludedSet)
			if !isSubset(candidate, current) {
				return nil, fmt.Errorf("delta: address elimination produced a non-subset candidate")
			}

			fails, err := m.reproducesFailure(ctx, ops, candidate, test)
			if err != nil {
				return nil, err
			}
			if !fails {
				continue
			}

			current = candidate
			reduced = true
			m.log(
				"address elimination accepted: sampled=%d removed=%d",
				sampleSize,
				startCount-countOnes(current),
			)
			break
		}

		if reduced {
			continue
		}

		if sampleSize == 1 {
			break
		}
		sampleSize = max(1, sampleSize/2)
		m.log("address elimination reducing sample size to %d", sampleSize)
	}

	return current, nil
}

func (m *Minimizer) emptyStructureElimination(
	ctx context.Context,
	ops []TraceOp,
	scopeForest []*scopeNode,
	guards []bool,
	test testFunc,
) ([]bool, error) {
	current := copyGuards(guards)

	for {
		if err := ctx.Err(); err != nil {
			return nil, err
		}

		startCount := countOnes(current)
		candidate := removeEmptyStructureGuards(current, scopeForest)
		if !isSubset(candidate, current) {
			return nil, fmt.Errorf("delta: empty structure elimination produced a non-subset candidate")
		}
		if countOnes(candidate) == startCount {
			break
		}

		fails, err := m.reproducesFailure(ctx, ops, candidate, test)
		if err != nil {
			return nil, err
		}
		if !fails {
			break
		}

		current = candidate
		m.log("empty structure elimination accepted: removed=%d", startCount-countOnes(current))
	}

	return current, nil
}

func (m *Minimizer) initialAddressSampleSize(addressCount int) int {
	if addressCount <= 1 {
		return 0
	}
	if m.cfg.MaxFactor > 0 {
		return min(addressCount-1, m.cfg.MaxFactor)
	}
	return max(1, addressCount/2)
}

func (m *Minimizer) reproducesFailure(
	ctx context.Context,
	ops []TraceOp,
	guards []bool,
	test testFunc,
) (bool, error) {
	if err := ctx.Err(); err != nil {
		return false, err
	}
	candidate := operationsForGuards(ops, guards)
	out, err := test(ctx, candidate)
	if err != nil {
		return false, err
	}
	return out == outcomeFail, nil
}

func buildScopeForest(ops []TraceOp) []*scopeNode {
	roots := make([]*scopeNode, 0)
	stack := make([]*scopeNode, 0)

	for idx, op := range ops {
		if _, ok := scopeBeginToEnd[op.Kind]; ok {
			node := &scopeNode{
				kind:  op.Kind,
				start: idx,
				end:   -1,
			}
			if len(stack) == 0 {
				roots = append(roots, node)
			} else {
				parent := stack[len(stack)-1]
				parent.children = append(parent.children, node)
			}
			stack = append(stack, node)
			continue
		}

		beginKind, isEnd := scopeEndToBegin[op.Kind]
		if isEnd {
			for len(stack) > 0 {
				top := stack[len(stack)-1]
				stack = stack[:len(stack)-1]
				if top.kind == beginKind {
					top.end = idx
					break
				}
			}
			continue
		}

		if len(stack) > 0 {
			parent := stack[len(stack)-1]
			parent.leaves = append(parent.leaves, idx)
		}
	}

	return filterValidScopeRoots(roots)
}

func filterValidScopeRoots(roots []*scopeNode) []*scopeNode {
	result := make([]*scopeNode, 0, len(roots))
	for _, root := range roots {
		if filterValidScopeNode(root) {
			result = append(result, root)
		}
	}
	return result
}

func filterValidScopeNode(node *scopeNode) bool {
	filtered := make([]*scopeNode, 0, len(node.children))
	for _, child := range node.children {
		if filterValidScopeNode(child) {
			filtered = append(filtered, child)
		}
	}
	node.children = filtered
	return node.end >= node.start
}

func removeEmptyStructureGuards(current []bool, scopeForest []*scopeNode) []bool {
	candidate := copyGuards(current)
	for _, root := range scopeForest {
		markEmptyScopes(root, current, candidate)
	}
	return candidate
}

func markEmptyScopes(node *scopeNode, baseline []bool, candidate []bool) int {
	active := 0
	for _, idx := range node.leaves {
		if idx >= 0 && idx < len(baseline) && baseline[idx] {
			active++
		}
	}
	for _, child := range node.children {
		active += markEmptyScopes(child, baseline, candidate)
	}
	if active == 0 {
		if node.start >= 0 && node.start < len(candidate) {
			candidate[node.start] = false
		}
		if node.end >= 0 && node.end < len(candidate) {
			candidate[node.end] = false
		}
	}
	return active
}

func enabledNonStructuralIndices(ops []TraceOp, guards []bool) []int {
	indices := make([]int, 0)
	for idx, enabled := range guards {
		if !enabled {
			continue
		}
		if isStructuralKind(ops[idx].Kind) {
			continue
		}
		indices = append(indices, idx)
	}
	return indices
}

func removeNonStructuralPrefix(guards []bool, sparse []int, remove int) []bool {
	candidate := copyGuards(guards)
	if remove > len(sparse) {
		remove = len(sparse)
	}
	for i := 0; i < remove; i++ {
		candidate[sparse[i]] = false
	}
	return candidate
}

func collectActiveMetadata(ops []TraceOp, guards []bool, mandatoryKinds map[string]struct{}) []operationMeta {
	collector := metaCollector{
		prevContract: common.Address{},
		mandatory:    mandatoryKinds,
	}

	meta := make([]operationMeta, len(ops))
	for idx, op := range ops {
		if !guards[idx] {
			continue
		}
		entry := collector.collect(op)
		entry.Index = idx
		meta[idx] = entry
	}
	return meta
}

func activeContracts(meta []operationMeta, guards []bool) []common.Address {
	set := make(map[common.Address]struct{})
	for idx, entry := range meta {
		if !guards[idx] {
			continue
		}
		if !entry.HasContract {
			continue
		}
		set[entry.Contract] = struct{}{}
	}
	addrs := make([]common.Address, 0, len(set))
	for addr := range set {
		addrs = append(addrs, addr)
	}
	sort.Slice(addrs, func(i, j int) bool { return addrs[i].Hex() < addrs[j].Hex() })
	return addrs
}

func disableContracts(
	guards []bool,
	meta []operationMeta,
	excluded map[common.Address]struct{},
) []bool {
	candidate := copyGuards(guards)
	for idx, enabled := range guards {
		if !enabled {
			continue
		}
		entry := meta[idx]
		if !entry.HasContract {
			continue
		}
		if _, ok := excluded[entry.Contract]; ok {
			candidate[idx] = false
		}
	}
	return candidate
}

func operationsForGuards(ops []TraceOp, guards []bool) []TraceOp {
	result := make([]TraceOp, 0, countOnes(guards))
	for idx, enabled := range guards {
		if enabled {
			result = append(result, ops[idx])
		}
	}
	return result
}

func newGuardVector(size int) []bool {
	guards := make([]bool, size)
	for idx := range guards {
		guards[idx] = true
	}
	return guards
}

func copyGuards(guards []bool) []bool {
	out := make([]bool, len(guards))
	copy(out, guards)
	return out
}

func countOnes(guards []bool) int {
	count := 0
	for _, enabled := range guards {
		if enabled {
			count++
		}
	}
	return count
}

func isSubset(candidate []bool, current []bool) bool {
	if len(candidate) != len(current) {
		return false
	}
	for idx := range candidate {
		if candidate[idx] && !current[idx] {
			return false
		}
	}
	return true
}

func isStructuralKind(kind string) bool {
	if _, ok := scopeBeginToEnd[kind]; ok {
		return true
	}
	if _, ok := scopeEndToBegin[kind]; ok {
		return true
	}
	return false
}

// reducePrefix removes leading operations from the trace using binary search.
// It is preserved as an internal compatibility helper for tests.
func (m *Minimizer) reducePrefix(ctx context.Context, meta []operationMeta, test testFunc) ([]TraceOp, []operationMeta, error) {
	if len(meta) == 0 {
		return nil, nil, fmt.Errorf("delta: empty metadata")
	}

	full := buildOperations(meta, func(operationMeta) bool { return true })
	out, err := test(ctx, full)
	if err != nil {
		return nil, nil, err
	}
	if out != outcomeFail {
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
		candidate := buildOperations(meta, func(m operationMeta) bool {
			return m.Index >= mid
		})

		out, err := test(ctx, candidate)
		if err != nil {
			return nil, nil, err
		}

		if out == outcomeFail {
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

// reduceAddresses removes operations by sampled contract groups.
// It is preserved as an internal compatibility helper for tests.
func (m *Minimizer) reduceAddresses(
	ctx context.Context,
	ops []TraceOp,
	meta []operationMeta,
	test testFunc,
) ([]TraceOp, []operationMeta, error) {
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

			candidate := buildOperations(currentMeta, func(meta operationMeta) bool {
				if !meta.HasContract {
					return true
				}
				_, skip := excludeSet[meta.Contract]
				return !skip
			})

			out, err := test(ctx, candidate)
			if err != nil {
				return nil, nil, err
			}

			if out == outcomeFail {
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

func collectMetadata(ops []TraceOp, mandatoryKinds map[string]struct{}) []operationMeta {
	collector := metaCollector{
		prevContract: common.Address{},
		mandatory:    mandatoryKinds,
	}
	meta := make([]operationMeta, 0, len(ops))
	for idx, op := range ops {
		entry := collector.collect(op)
		entry.Index = idx
		meta = append(meta, entry)
	}
	return meta
}

func (c *metaCollector) collect(op TraceOp) operationMeta {
	mandatory := false
	if c.mandatory != nil {
		_, mandatory = c.mandatory[op.Kind]
	}

	contract, has := c.contractFor(op)
	if has {
		c.prevContract = contract
	}

	return operationMeta{
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

func buildOperations(meta []operationMeta, include func(operationMeta) bool) []TraceOp {
	result := make([]TraceOp, 0, len(meta))
	for _, m := range meta {
		if include(m) || m.Mandatory {
			result = append(result, m.Op)
		}
	}
	return result
}

func uniqueContracts(meta []operationMeta) []common.Address {
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
