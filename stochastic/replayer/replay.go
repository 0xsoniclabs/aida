// Copyright 2024 Fantom Foundation
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
	"fmt"
	"math/rand"
	"time"

	"github.com/0xsoniclabs/aida/logger"
	"github.com/0xsoniclabs/aida/state"
	"github.com/0xsoniclabs/aida/stochastic"
	"github.com/0xsoniclabs/aida/stochastic/operations"
	"github.com/0xsoniclabs/aida/stochastic/recorder"
	"github.com/0xsoniclabs/aida/stochastic/statistics/generator"
	"github.com/0xsoniclabs/aida/stochastic/statistics/markov_chain"
	"github.com/0xsoniclabs/aida/utils"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/tracing"
	"github.com/holiman/uint256"
)

// Parameterisable simulation constants
var (
	BalanceRange int64 = 100000  // balance range for generating randomized values
	NonceRange   int   = 1000000 // nonce range for generating randomized nonces
)

// Simulation constants
const (
	MaxCodeSize  = 24576 // fixed upper limit by EIP-170
	FinaliseFlag = true  // flag for Finalise() StateDB operation
)

// replayContext data structure as a context for simulating StateDB operations
type replayContext struct {
	db              state.StateDB         // StateDB database
	traceDebug      bool                  // trace-debug flag
	log             logger.Logger         // logger for output
	rg              *rand.Rand            // random generator for sampling
	contracts       generator.ArgumentSet // random argument generator for contracts
	selfDestructed  map[int64]struct{}    // set of self destructed accounts in a block
	keys            generator.ArgumentSet // random argument generator for keys
	values          generator.ArgumentSet // random argument generator for values
	snapshots       generator.SnapshotSet // random generator for snapshot ids
	activeSnapshots []int                 // stack of active snapshots
	totalTx         uint64                // total number of transactions
	txNum           uint32                // current transaction number
	blockNum        uint64                // current block number
	syncPeriodNum   uint64                // current sync-period number
}

// newReplayContext creates a new replay context for execution StateDB operations stochastically.
func newReplayContext(
	rg *rand.Rand,
	db state.StateDB,
	contracts generator.ArgumentSet,
	keys generator.ArgumentSet,
	values generator.ArgumentSet,
	snapshots generator.SnapshotSet,
	log logger.Logger,
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
	}
}

// populateReplayContext creates a stochastic state and primes the StateDB
func populateReplayContext(cfg *utils.Config, e *recorder.StateJSON, db state.StateDB, rg *rand.Rand, log logger.Logger) (*replayContext, error) {
	// produce random variables for contract addresses,
	// storage-keys, storage addresses, and snapshot ids.
	contracts := generator.NewSingleUseArgumentSet(
		generator.NewReusableArgumentSet(
			e.Contracts.Counting.N,
			generator.NewEmpiricalArgSetRandomizer(
				rg,
				e.Contracts.Queuing.Distribution,
				e.Contracts.Counting.ECDF),
		))
	keys := generator.NewReusableArgumentSet(
		e.Keys.Counting.N,
		generator.NewEmpiricalArgSetRandomizer(
			rg,
			e.Keys.Queuing.Distribution,
			e.Keys.Counting.ECDF),
	)
	values := generator.NewReusableArgumentSet(
		e.Values.Counting.N,
		generator.NewEmpiricalArgSetRandomizer(
			rg,
			e.Values.Queuing.Distribution,
			e.Values.Counting.ECDF),
	)
	snapshots := generator.NewEmpiricalSnapshotRandomizer(rg, e.SnapshotECDF)

	// setup state
	ss := newReplayContext(rg, db, contracts, keys, values, snapshots, log)

	// create accounts in StateDB before starting the simulation
	err := ss.prime()
	if err != nil {
		return nil, err
	}

	return &ss, nil
}

// getStochasticMatrix returns the stochastic matrix with its operations and the initial state
func getStochasticMatrix(e *recorder.StateJSON) (*markov_chain.MarkovChain, int, error) {
	ops := e.Operations
	A := e.StochasticMatrix
	mc, err := markov_chain.New(A, ops)
	if err != nil {
		return nil, 0, fmt.Errorf("getStochasticMatrix: cannot retrieve markov chain from estimation model. Error: %v", err)
	}
	state, f_err := mc.FindState(operations.OpMnemo(operations.BeginSyncPeriodID))
	if f_err != nil {
		return nil, 0, fmt.Errorf("getStochasticMatrix: cannot retrieve initial state. Error: %v", f_err)
	}
	return mc, state, nil
}

// retrieve operations and stochastic matrix from simulation object

// RunStochasticReplay runs the stochastic simulation for StateDB operations.
// It requires the simulation model and simulation length. The trace-debug flag
// enables/disables the printing of StateDB operations and their arguments on
// the screen.
func RunStochasticReplay(db state.StateDB, e *recorder.StateJSON, nBlocks int, cfg *utils.Config, log logger.Logger) error {
	var (
		opFrequency [operations.NumOps]uint64 // operation frequency
		numOps      uint64                    // total number of operations
	)

	if db.GetShadowDB() == nil {
		log.Notice("No validation with a shadow DB.")
	}
	log.Noticef("balance range %d", cfg.BalanceRange)
	BalanceRange = cfg.BalanceRange

	log.Noticef("nonce range %d", cfg.NonceRange)
	NonceRange = cfg.NonceRange

	// random generator
	rg := rand.New(rand.NewSource(cfg.RandomSeed))
	log.Noticef("using random seed %d", cfg.RandomSeed)

	// create a stochastic state
	ss, err := populateReplayContext(cfg, e, db, rg, log)
	if err != nil {
		return err
	}

	// get stochastic matrix
	mc, state, mc_err := getStochasticMatrix(e)
	if mc_err != nil {
		return fmt.Errorf("RunStochasticReplay: expected a markov chain. Error: %v", mc_err)
	}

	// progress message setup
	var (
		start    time.Time
		sec      float64
		lastSec  float64
		runErr   error
		errCount int
	)

	start = time.Now()
	sec = time.Since(start).Seconds()
	lastSec = time.Since(start).Seconds()
	// if block after priming is greater or equal to debug block, enable debug.
	if cfg.Debug && ss.blockNum >= cfg.DebugFrom {
		ss.enableDebug()
	}

	block := 0
	// inclusive range
	log.Noticef("Simulation block range: first %v, last %v", ss.blockNum, ss.blockNum+uint64(nBlocks-1))
	for {
		label, err := mc.Label(state)
		if err != nil {
			return fmt.Errorf("RunStochasticReplay: cannot retrieve state label. Error: %v", err)
		}

		// decode opcode
		op, addrCl, keyCl, valueCl, err := operations.DecodeOpcode(label)
		if err != nil {
			return fmt.Errorf("RunStochasticReplay: cannot decode opcode. Error: %v", err)
		}

		// keep track of stats
		numOps++
		opFrequency[op]++

		// execute operation with its argument classes
		ss.execute(op, addrCl, keyCl, valueCl)

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
		sec = time.Since(start).Seconds()
		if sec-lastSec >= 15 {
			log.Debugf("Elapsed time: %.0f s, at block %v", sec, block)
			lastSec = sec
		}

		// check for errors
		if err := ss.db.Error(); err != nil {
			errCount++
			if runErr == nil {
				runErr = fmt.Errorf("error: stochastic replay failed")
			}

			runErr = fmt.Errorf("%v\n\tBlock %v Tx %v: %v", runErr, ss.blockNum, ss.txNum, err)
			if !cfg.ContinueOnFailure {
				break
			}
		}

		// transit to next state in Markovian process
		u := rg.Float64()
		state, err = mc.Sample(state, u)
		if err != nil {
			return fmt.Errorf("RunStochasticReplay: Failed sampling the next state. Error: %v", err)
		}
	}

	// print progress summary
	log.Noticef("Total elapsed time: %.3f s, processed %v blocks", sec, block)
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
	return runErr
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
		db.AddBalance(addr, uint256.NewInt(uint64(ss.rg.Int63n(BalanceRange))), 0)
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
	)

	// fetch indexes from index access generators only when an argument is required
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
		ss.log.Infof("opcode:%v (%v)", operations.OpText[op], opc)

		// print indexes of contract address, storage key, and storage value.
		if addrCl != stochastic.NoArgID {
			ss.log.Infof(" addr-idx: %v", addrIdx)
		}
		if keyCl != stochastic.NoArgID {
			ss.log.Infof(" key-idx: %v", keyIdx)
		}
		if valueCl != stochastic.NoArgID {
			ss.log.Infof(" value-idx: %v", valueIdx)
		}
	}

	switch op {
	case operations.AddBalanceID:
		value := rg.Int63n(BalanceRange)
		if ss.traceDebug {
			ss.log.Infof("value: %v", value)
		}
		db.AddBalance(addr, uint256.NewInt(uint64(value)), 0)

	case operations.BeginBlockID:
		if ss.traceDebug {
			ss.log.Infof(" id: %v", ss.blockNum)
		}
		err := db.BeginBlock(ss.blockNum)
		if err != nil {
			ss.log.Fatal(err)
		}
		ss.txNum = 0
		ss.selfDestructed = map[int64]struct{}{} // reset selfDestructed accounts set

	case operations.BeginSyncPeriodID:
		if ss.traceDebug {
			ss.log.Infof(" id: %v", ss.syncPeriodNum)
		}
		db.BeginSyncPeriod(ss.syncPeriodNum)

	case operations.BeginTransactionID:
		if ss.traceDebug {
			ss.log.Infof(" id: %v", ss.txNum)
		}
		err := db.BeginTransaction(ss.txNum)
		if err != nil {
			ss.log.Fatal(err)
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
			ss.log.Fatal(err)
		}
		ss.blockNum++
		for addrIdx, _ := range ss.selfDestructed {
			if err := ss.contracts.Remove(addrIdx); err != nil {
				return fmt.Errorf("deleteAccounts: failed deleting index (%v).", addrIdx)
			}
		}
		ss.selfDestructed = map[int64]struct{}{}

	case operations.EndSyncPeriodID:
		db.EndSyncPeriod()
		ss.syncPeriodNum++

	case operations.EndTransactionID:
		err := db.EndTransaction()
		if err != nil {
			ss.log.Fatal(err)
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
				ss.log.Infof(" id: %v", snapshot)
			}
			ss.activeSnapshots = ss.activeSnapshots[:snapshotIdx+1]

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
		sz := rg.Intn(MaxCodeSize-1) + 1
		if ss.traceDebug {
			ss.log.Infof(" code-size: %v", sz)
		}
		code := make([]byte, sz)
		_, err := rg.Read(code)
		if err != nil {
			ss.log.Fatalf("execute: error producing a random byte slice for code. Error: %v", err)
		}
		db.SetCode(addr, code)

	case operations.SetNonceID:
		value := uint64(rg.Intn(NonceRange))
		db.SetNonce(addr, value, tracing.NonceChangeUnspecified)

	case operations.SetStateID:
		db.SetState(addr, key, value)

	case operations.SetTransientStateID:
		db.SetTransientState(addr, key, value)

	case operations.SnapshotID:
		id := db.Snapshot()
		if ss.traceDebug {
			ss.log.Infof(" id: %v", id)
		}
		ss.activeSnapshots = append(ss.activeSnapshots, id)

	case operations.SubBalanceID:
		var balance uint64
		balance = db.GetBalance(addr).Uint64()
		if balance > 0 {
			// get a delta that does not exceed current balance
			// in the current snapshot
			value := uint64(rg.Int63n(int64(balance)))
			if ss.traceDebug {
				ss.log.Infof(" value: %v", value)
			}
			db.SubBalance(addr, uint256.NewInt(value), 0)
		}
	default:
		return fmt.Errorf("execute: invalid operation %v; opcode %v", operations.OpText[op], op)
	}
	return nil
}
