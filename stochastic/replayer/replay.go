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

package replayer

import (
	"errors"
	"fmt"
	"math/rand"
	"time"

	"github.com/0xsoniclabs/aida/logger"
	"github.com/0xsoniclabs/aida/state"
	"github.com/0xsoniclabs/aida/stochastic"
	"github.com/0xsoniclabs/aida/stochastic/operations"
	"github.com/0xsoniclabs/aida/stochastic/recorder"
	"github.com/0xsoniclabs/aida/stochastic/replayer/arguments"
	"github.com/0xsoniclabs/aida/stochastic/statistics/markov"
	"github.com/0xsoniclabs/aida/utils"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/tracing"
	"github.com/holiman/uint256"
)

var (
	progressLogIntervalSec = 15
	randReadBytes          = func(rg *rand.Rand, buf []byte) (int, error) { return rg.Read(buf) }
	mcLabel                = func(mc *markov.Chain, state int) (string, error) { return mc.Label(state) }
	mcSample               = func(mc *markov.Chain, i int, u float64) (int, error) { return mc.Sample(i, u) }
)

// Parameterisable simulation constants
// Simulation constants
const (
	MaxCodeSize = 24576 // fixed upper limit by EIP-170
)

// replayContext data structure as a context for simulating StateDB operations
type replayContext struct {
	db              state.StateDB         // StateDB database
	traceDebug      bool                  // trace-debug flag
	log             logger.Logger         // logger for output
	rg              *rand.Rand            // random arguments for sampling
	contracts       arguments.Set         // random argument arguments for contracts
	selfDestructed  map[int64]struct{}    // set of self destructed accounts in a block
	keys            arguments.Set         // random argument arguments for keys
	values          arguments.Set         // random argument arguments for values
	snapshots       arguments.SnapshotSet // random arguments for snapshot ids
	activeSnapshots []int                 // stack of active snapshots
	totalTx         uint64                // total number of transactions
	txNum           uint32                // current transaction number
	blockNum        uint64                // current block number
	syncPeriodNum   uint64                // current sync-period number
	balanceRange    int64                 // balance range for randomized values
	nonceRange      int                   // nonce range for randomized nonces
	balanceSampler  *arguments.ScalarSampler
	nonceSampler    *arguments.ScalarSampler
	codeSampler     *arguments.ScalarSampler
}

// newReplayContext creates a new replay context for execution StateDB operations stochastically.
func newReplayContext(
	rg *rand.Rand,
	db state.StateDB,
	contracts arguments.Set,
	keys arguments.Set,
	values arguments.Set,
	snapshots arguments.SnapshotSet,
	log logger.Logger,
	balanceRange int64,
	nonceRange int,
) replayContext {
	// return stochastic state
	return replayContext{
		db:             db,
		contracts:      contracts,
		keys:           keys,
		values:         values,
		snapshots:      snapshots,
		traceDebug:     false,
		selfDestructed: map[int64]struct{}{},
		blockNum:       1,
		syncPeriodNum:  1,
		rg:             rg,
		log:            log,
		balanceRange:   balanceRange,
		nonceRange:     nonceRange,
		balanceSampler: arguments.NewScalarSampler(rg, nil),
		nonceSampler:   arguments.NewScalarSampler(rg, nil),
		codeSampler:    arguments.NewScalarSampler(rg, nil),
	}
}

// populateReplayContext creates a stochastic state and primes the StateDB
func populateReplayContext(
	e *recorder.StatsJSON,
	db state.StateDB,
	rg *rand.Rand,
	log logger.Logger,
	balanceRange int64,
	nonceRange int,
) (*replayContext, error) {
	// produce random variables for contract addresses,
	// storage-keys, storage addresses, and snapshot ids.
	contractRandomizer, err := arguments.NewRandomizer(
		rg,
		e.Contracts.Queuing.Distribution,
		e.Contracts.Counting.ECDF)
	if err != nil {
		return nil, fmt.Errorf("populateReplayContext: construct contract randomizer: %w", err)
	}
	contracts := arguments.NewSingleUse(
		arguments.NewReusable(e.Contracts.Counting.N, contractRandomizer),
	)

	keyRandomizer, err := arguments.NewRandomizer(
		rg,
		e.Keys.Queuing.Distribution,
		e.Keys.Counting.ECDF)
	if err != nil {
		return nil, fmt.Errorf("populateReplayContext: construct key randomizer: %w", err)
	}
	keys := arguments.NewReusable(e.Keys.Counting.N, keyRandomizer)

	valueRandomizer, err := arguments.NewRandomizer(
		rg,
		e.Values.Queuing.Distribution,
		e.Values.Counting.ECDF)
	if err != nil {
		return nil, fmt.Errorf("populateReplayContext: construct value randomizer: %w", err)
	}
	values := arguments.NewReusable(e.Values.Counting.N, valueRandomizer)

	snapshots := arguments.NewEmpiricalSnapshotRandomizer(rg, e.SnapshotECDF)

	if recordedRange := e.Balance.Max + 1; recordedRange > balanceRange {
		balanceRange = recordedRange
	}
	if balanceRange <= 0 {
		balanceRange = 1
	}
	if recordedNonce := int(e.Nonce.Max + 1); recordedNonce > nonceRange {
		nonceRange = recordedNonce
	}
	if nonceRange <= 0 {
		nonceRange = 1
	}

	// setup state
	ss := newReplayContext(rg, db, contracts, keys, values, snapshots, log, balanceRange, nonceRange)
	ss.balanceSampler = arguments.NewScalarSampler(rg, e.Balance.ECDF)
	ss.nonceSampler = arguments.NewScalarSampler(rg, e.Nonce.ECDF)
	ss.codeSampler = arguments.NewScalarSampler(rg, e.CodeSize.ECDF)

	// create accounts in StateDB before starting the simulation
	err = ss.prime()
	if err != nil {
		return nil, err
	}

	return &ss, nil
}

// getStochasticMatrix returns the stochastic matrix with its operations and the initial state
func getStochasticMatrix(e *recorder.StatsJSON) (*markov.Chain, int, error) {
	ops := e.Operations
	A := e.StochasticMatrix
	mc, err := markov.New(A, ops)
	if err != nil {
		return nil, 0, fmt.Errorf("getStochasticMatrix: cannot retrieve markov chain from estimation model: %w", err)
	}
	opM, err := operations.OpMnemo(operations.BeginSyncPeriodID)
	if err != nil {
		return nil, 0, fmt.Errorf("getStochasticMatrix: cannot retrieve OpMnemo from estimation model: %w", err)
	}
	state, _ := mc.Find(opM)
	if state < 0 {
		return nil, 0, fmt.Errorf("getStochasticMatrix: cannot retrieve initial state. Error: not found")
	}
	return mc, state, nil
}

// retrieve operations and stochastic matrix from simulation object

// RunStochasticReplay runs the stochastic simulation for StateDB operations.
// It requires the simulation model and simulation length. The trace-debug flag
// enables/disables the printing of StateDB operations and their arguments on
// the screen.
func RunStochasticReplay(db state.StateDB, e *recorder.StatsJSON, nBlocks int, cfg *utils.Config, log logger.Logger) error {
	var (
		opFrequency [operations.NumOps]uint64 // operation frequency
		numOps      uint64                    // total number of operations
		errCount    int
		errList     []error
	)

	if db.GetShadowDB() == nil {
		log.Notice("No validation with a shadow DB.")
	}
	balanceRange := cfg.BalanceRange
	if balanceRange <= 0 {
		log.Warning("balance range <= 0, defaulting to 1")
		balanceRange = 1
	}

	nonceRange := cfg.NonceRange
	if nonceRange <= 0 {
		log.Warning("nonce range <= 0, defaulting to 1")
		nonceRange = 1
	}

	// random arguments
	rg := rand.New(rand.NewSource(cfg.RandomSeed))
	log.Noticef("using random seed %d", cfg.RandomSeed)

	// create a stochastic state
	ss, err := populateReplayContext(e, db, rg, log, balanceRange, nonceRange)
	if err != nil {
		return err
	}
	log.Noticef("balance range %d", ss.balanceRange)
	log.Noticef("nonce range %d", ss.nonceRange)

	// get stochastic matrix
	mc, state, mcErr := getStochasticMatrix(e)
	if mcErr != nil {
		return fmt.Errorf("RunStochasticReplay: expected a markov chain: %w", mcErr)
	}

	// progress message setup
	start := time.Now()
	lastLog := start
	interval := time.Duration(progressLogIntervalSec) * time.Second
	// if block after priming is greater or equal to debug block, enable debug.
	if cfg.Debug && ss.blockNum >= cfg.DebugFrom {
		ss.enableDebug()
	}

	block := 0
	// inclusive range
	log.Noticef("Simulation block range: first %v, last %v", ss.blockNum, ss.blockNum+uint64(nBlocks-1))
	for {
		label, err := mcLabel(mc, state)
		if err != nil {
			return fmt.Errorf("RunStochasticReplay: cannot retrieve state label: %w", err)
		}

		// decode opcode
		op, addrCl, keyCl, valueCl, err := operations.DecodeOpcode(label)
		if err != nil {
			return fmt.Errorf("RunStochasticReplay: cannot decode opcode: %w", err)
		}

		// keep track of stats
		numOps++
		opFrequency[op]++

		// execute operation with its argument classes
		if err := ss.execute(op, addrCl, keyCl, valueCl); err != nil {
			return fmt.Errorf("RunStochasticReplay: %w", err)
		}

		// check for end of simulation
		if op == operations.EndBlockID {
			block++
			if block >= nBlocks {
				break
			}
			// if current block is greater or equal to debug block, enable debug.
			if cfg.Debug && !ss.traceDebug && ss.blockNum >= cfg.DebugFrom {
				ss.enableDebug()
			}
		}

		// report progress
		elapsed := time.Since(start)
		if interval <= 0 || time.Since(lastLog) >= interval {
			log.Debugf("Elapsed time: %.0f s, at block %v", elapsed.Seconds(), block)
			lastLog = time.Now()
		}

		// check for errors
		if err := ss.db.Error(); err != nil {
			errCount++
			errList = append(errList, fmt.Errorf("block %v tx %v: %w", ss.blockNum, ss.txNum, err))
			if !cfg.ContinueOnFailure {
				break
			}
		}

		// transit to next state in Markovian process
		u := rg.Float64()
		state, err = mcSample(mc, state, u)
		if err != nil {
			return fmt.Errorf("RunStochasticReplay: failed sampling the next state: %w", err)
		}
	}

	// print progress summary
	elapsed := time.Since(start)
	log.Noticef("Total elapsed time: %.3f s, processed %v blocks", elapsed.Seconds(), block)
	if errCount > 0 {
		log.Warningf("%v errors were found", errCount)
	}

	// print statistics
	log.Noticef("SyncPeriods: %v", ss.syncPeriodNum)
	log.Noticef("Blocks: %v", ss.blockNum)
	log.Noticef("Transactions: %v", ss.totalTx)
	log.Noticef("Operations: %v", numOps)
	log.Noticef("Operation Frequencies:")
	for op := range operations.NumOps {
		log.Noticef("\t%v: %v", operations.OpText[op], opFrequency[op])
	}
	if len(errList) == 0 {
		return nil
	}
	joined := errors.Join(errList...)
	return fmt.Errorf("stochastic replay failed: %w", joined)
}

// prime creates initial accounts in the StateDB before starting the simulation.
func (ss *replayContext) prime() error {
	numInitialAccounts := ss.contracts.Size() + 1
	ss.log.Notice("Start priming...")
	ss.log.Noticef("\tinitializing %v accounts\n", numInitialAccounts)
	pt := utils.NewProgressTracker(int(numInitialAccounts), ss.log)
	db := ss.db
	db.BeginSyncPeriod(0)
	err := db.BeginBlock(0)
	if err != nil {
		return err
	}
	err = db.BeginTransaction(0)
	if err != nil {
		return err
	}
	for i := range int64(numInitialAccounts) {
		var addr common.Address
		if addr, err = operations.ToAddress(i); err != nil {
			return err
		}
		db.CreateAccount(addr)
		value := ss.balanceSampler.Sample(ss.balanceRange)
		db.AddBalance(addr, uint256.NewInt(uint64(value)), 0)
		pt.PrintProgress()
	}
	ss.log.Notice("Finalizing...")
	if err = db.EndTransaction(); err != nil {
		return err
	}
	if err = db.EndBlock(); err != nil {
		return err
	}
	db.EndSyncPeriod()
	ss.log.Notice("End priming...")
	return nil
}

// EnableDebug set traceDebug flag to true, and enable debug message when executing an operation
func (ss *replayContext) enableDebug() {
	ss.traceDebug = true
}

// execute StateDB operations on a stochastic state.
func (ss *replayContext) execute(op int, addrCl int, keyCl int, valueCl int) error {
	var (
		addr  common.Address
		key   common.Hash
		value common.Hash
		db    = ss.db
		rg    = ss.rg
		msg   string
	)

	// fetch indexes from index access argumentss only when an argument is required
	var addrIdx int64
	var keyIdx int64
	var valueIdx int64
	var err error

	if addrCl != stochastic.NoArgID {
		addrIdx, err = ss.contracts.Choose(addrCl)
		if err != nil {
			return fmt.Errorf("execute: failed to fetch contract address. Error: %v", err)
		}
	}
	if keyCl != stochastic.NoArgID {
		keyIdx, err = ss.keys.Choose(keyCl)
		if err != nil {
			return fmt.Errorf("execute: failed to fetch storage key. Error: %v", err)
		}
	}
	if valueCl != stochastic.NoArgID {
		valueIdx, err = ss.values.Choose(valueCl)
		if err != nil {
			return fmt.Errorf("execute: failed to fetch storage value. Error: %v", err)
		}
	}

	// convert index to address/hashes
	if addrCl != stochastic.NoArgID {
		addr, err = operations.ToAddress(addrIdx)
		if err != nil {
			return fmt.Errorf("execute: failed to convert index to address. Error: %v", err)
		}
	}
	if keyCl != stochastic.NoArgID {
		key, err = operations.ToHash(keyIdx)
		if err != nil {
			return fmt.Errorf("execute: failed to convert index to hash. Error: %v", err)
		}
	}
	if valueCl != stochastic.NoArgID {
		value, err = operations.ToHash(valueIdx)
		if err != nil {
			return err
		}
	}

	// print opcode and its arguments
	if ss.traceDebug {
		// print operation
		opc, err := operations.EncodeOpcode(op, addrCl, keyCl, valueCl)
		if err != nil {
			return fmt.Errorf("execute: failed encoding opcode. Error: %v", err)
		}

		// print operation
		msg = fmt.Sprintf("opcode:%v (%v) ", operations.OpText[op], opc)

		// print indexes of contract address, storage key, and storage value.
		if addrCl != stochastic.NoArgID {
			msg = fmt.Sprintf("%v addr-idx: %v", msg, addrIdx)
			msg = fmt.Sprintf("%v addr: %v", msg, addr)
		}
		if keyCl != stochastic.NoArgID {
			msg = fmt.Sprintf("%v key-idx: %v", msg, keyIdx)
			msg = fmt.Sprintf("%v key: %v", msg, key)
		}
		if valueCl != stochastic.NoArgID {
			msg = fmt.Sprintf("%v value-idx: %v", msg, valueIdx)
			msg = fmt.Sprintf("%v value: %v", msg, value)
		}
	}

	switch op {
	case operations.AddBalanceID:
		value := ss.balanceSampler.Sample(ss.balanceRange)
		if ss.traceDebug {
			msg = fmt.Sprintf("%v value: %v", msg, value)
		}
		db.AddBalance(addr, uint256.NewInt(uint64(value)), 0)

	case operations.BeginBlockID:
		if ss.traceDebug {
			msg = fmt.Sprintf("%v id: %v", msg, ss.blockNum)
		}
		err := db.BeginBlock(ss.blockNum)
		if err != nil {
			return fmt.Errorf("execute: BeginBlock failed: %w", err)
		}
		ss.txNum = 0
		ss.selfDestructed = map[int64]struct{}{} // reset selfDestructed accounts set

	case operations.BeginSyncPeriodID:
		if ss.traceDebug {
			msg = fmt.Sprintf("%v id: %v", msg, ss.syncPeriodNum)
		}
		db.BeginSyncPeriod(ss.syncPeriodNum)

	case operations.BeginTransactionID:
		if ss.traceDebug {
			msg = fmt.Sprintf("%v id: %v", msg, ss.txNum)
		}
		err := db.BeginTransaction(ss.txNum)
		if err != nil {
			return fmt.Errorf("execute: BeginTransaction failed: %w", err)
		}
		ss.activeSnapshots = []int{}
		ss.selfDestructed = map[int64]struct{}{}

	case operations.CreateAccountID:
		db.CreateAccount(addr)

	case operations.CreateContractID:
		db.CreateContract(addr)

	case operations.EmptyID:
		db.Empty(addr)

	case operations.EndBlockID:
		err := db.EndBlock()
		if err != nil {
			return fmt.Errorf("execute: EndBlock failed: %w", err)
		}
		ss.blockNum++
		for addrIdx := range ss.selfDestructed {
			if err := ss.contracts.Remove(addrIdx); err != nil {
				return fmt.Errorf("deleteAccounts: failed deleting index (%v)", addrIdx)
			}
		}
		ss.selfDestructed = map[int64]struct{}{}

	case operations.EndSyncPeriodID:
		db.EndSyncPeriod()
		ss.syncPeriodNum++

	case operations.EndTransactionID:
		err := db.EndTransaction()
		if err != nil {
			return fmt.Errorf("execute: EndTransaction failed: %w", err)
		}
		ss.txNum++
		ss.totalTx++

	case operations.ExistID:
		db.Exist(addr)

	case operations.GetBalanceID:
		db.GetBalance(addr)

	case operations.GetCodeHashID:
		db.GetCodeHash(addr)

	case operations.GetCodeID:
		db.GetCode(addr)

	case operations.GetCodeSizeID:
		db.GetCodeSize(addr)

	case operations.GetCommittedStateID:
		db.GetCommittedState(addr, key)

	case operations.GetNonceID:
		db.GetNonce(addr)

	case operations.GetStateID:
		db.GetState(addr, key)

	case operations.GetStateAndCommittedStateID:
		db.GetStateAndCommittedState(addr, key)

	case operations.GetStorageRootID:
		db.GetStorageRoot(addr)

	case operations.GetTransientStateID:
		db.GetTransientState(addr, key)

	case operations.HasSelfDestructedID:
		db.HasSelfDestructed(addr)

	case operations.RevertToSnapshotID:
		snapshotNum := len(ss.activeSnapshots)
		if snapshotNum > 0 {
			snapshotIdx := snapshotNum - ss.snapshots.SampleSnapshot(snapshotNum) - 1
			if snapshotIdx < 0 {
				snapshotIdx = 0
			} else if snapshotIdx >= snapshotNum {
				snapshotIdx = snapshotNum - 1
			}
			snapshot := ss.activeSnapshots[snapshotIdx]
			if ss.traceDebug {
				msg = fmt.Sprintf("%v id: %v", msg, snapshot)
			}
			ss.activeSnapshots = ss.activeSnapshots[:snapshotIdx]

			// update active snapshots and perform a rollback in balance log
			db.RevertToSnapshot(snapshot)
		}

	case operations.SelfDestructID:
		db.SelfDestruct(addr)
		ss.selfDestructed[addrIdx] = struct{}{}

	case operations.SelfDestruct6780ID:
		db.SelfDestruct6780(addr)
		ss.selfDestructed[addrIdx] = struct{}{}

	case operations.SetCodeID:
		sz := int(ss.codeSampler.Sample(int64(MaxCodeSize)))
		if sz <= 0 {
			sz = 1
		}
		if ss.traceDebug {
			msg = fmt.Sprintf("%v code-size: %v", msg, sz)
		}
		code := make([]byte, sz)
		_, err := randReadBytes(rg, code)
		if err != nil {
			return fmt.Errorf("execute: error producing a random byte slice for code: %w", err)
		}
		db.SetCode(addr, code, tracing.CodeChangeUnspecified)

	case operations.SetNonceID:
		value := uint64(ss.nonceSampler.Sample(int64(ss.nonceRange)))
		db.SetNonce(addr, value, tracing.NonceChangeUnspecified)

	case operations.SetStateID:
		db.SetState(addr, key, value)

	case operations.SetTransientStateID:
		db.SetTransientState(addr, key, value)

	case operations.SnapshotID:
		id := db.Snapshot()
		if ss.traceDebug {
			msg = fmt.Sprintf("%v id: %v", msg, id)
		}
		ss.activeSnapshots = append(ss.activeSnapshots, id)

	case operations.SubBalanceID:
		var balance uint64
		balance = db.GetBalance(addr).Uint64()
		if balance > 0 {
			// get a delta that does not exceed current balance
			// in the current snapshot
			value := uint64(ss.balanceSampler.Sample(int64(balance)))
			if ss.traceDebug {
				msg = fmt.Sprintf("%v value: %v", msg, value)
			}
			db.SubBalance(addr, uint256.NewInt(value), 0)
		}
	default:
		return fmt.Errorf("execute: invalid operation %v; opcode %v", operations.OpText[op], op)
	}
	if ss.traceDebug && msg != "" {
		ss.log.Infof("%s", msg)
	}
	return nil
}
