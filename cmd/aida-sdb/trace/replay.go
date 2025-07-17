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

package trace

import (
	"fmt"
	"github.com/0xsoniclabs/aida/logger"
	"github.com/0xsoniclabs/aida/state"
	"github.com/0xsoniclabs/aida/stochastic/exponential"
	"github.com/0xsoniclabs/aida/tracer"
	"github.com/0xsoniclabs/aida/utils"
	"github.com/ethereum/go-ethereum/common"
	"github.com/urfave/cli/v2"
	"time"
)

// todo: will be handled in upcoming PR
func ReplayTrace(ctx *cli.Context) error {
	//	cfg, err := utils.NewConfig(ctx, utils.BlockRangeArgs)
	//	if err != nil {
	//		return err
	//	}
	//
	//	operationProvider, err := executor.OpenOperations(cfg)
	//	if err != nil {
	//
	//	}
	//
	//	defer operationProvider.Close()
	//
	//	rCtx := context.NewReplay()
	//
	//	processor := operationProcessor{cfg, rCtx}
	//
	//	var extra = []executor.Extension[[]operation.Operation]{
	//		profiler.MakeReplayProfiler[[]operation.Operation](cfg, rCtx),
	//	}
	//
	//	var aidaDb db.BaseDB
	//	// we need to open substate if we are priming
	//	if cfg.First > 0 && !cfg.SkipPriming {
	//		aidaDb, err = db.NewReadOnlyBaseDB(cfg.AidaDb)
	//		if err != nil {
	//			return fmt.Errorf("cannot open aida-db; %w", err)
	//		}
	//		defer aidaDb.Close()
	//	}
	//
	//	return replay(cfg, operationProvider, processor, extra, aidaDb)
	return nil
}



func Replay(db state.StateDB, file tracer.FileReader, cfg *utils.Config, log logger.Logger) error {
	// progress message setup
	var (
		start    time.Time
		sec      float64
		lastSec  float64
		runErr   error
		errCount int
		contract tracer.Queue[common.Address]
		keys     tracer.Queue[common.Hash]
		values   tracer.Queue[common.Hash]
	)

	if db.GetShadowDB() == nil {
		log.Notice("No validation with a shadow DB.")
	}

	start = time.Now()
	sec = time.Since(start).Seconds()
	lastSec = time.Since(start).Seconds()
	// if block after priming is greater or equal to debug block, enable debug.
	block := 0
	// inclusive range
	log.Noticef("Run trace for block range: first %v, last %v", ss.blockNum, ss.blockNum+uint64(nBlocks-1))
	for {
		// read 16-bit number from file
		state, err := file.ReadUint16()
		if (err != nil ) {
			// error
		}

		// decode opcode
		op, addrCl, keyCl, valueCl := DecodeOpcode(operations)


		switch(addrCl) {
		NoArgID:
		NewValueId:
			addr, err = file.ReadAddr()
			contract.Place(addr)
		PreviousValueID:
			addr = contract.Find(0)
			RecentValueID
			idx, err := file.ReadUInt8()
			if err != nil {
				// do error
			}
			addr = contract.Find(idx)
			contract.Place(addr)
		default:
			panic("Wrong address class")
		}

		switch(keyCl) {
		NoArgID:
		NewValueId:
			addr, err = file.ReadHash()
			keys.Place(key)
		PreviousValueID:
			addr = keys.Find(0)
			RecentValueID
			idx, err := file.ReadUInt8()
			if err != nil {
				// do error
			}
			addr = keys.Find(idx)
		default:
			panic("Wrong key class")
		}

		switch(valueCl) {
		NoArgID:
		NewValueId:
			addr, err = file.ReadHash()
			values.Place(key)
		PreviousValueID:
			addr = values.Find(0)
			RecentValueID
			idx, err := file.ReadUInt8()
			if err != nil {
				// do error
			}
			addr = values.Find(idx)
		default:
			panic("Wrong key class")
		}

		// execute operation with its argument classes
		Execute(file, op, addr, key, value)

		// check for end of simulation
		if op == EndBlockID {
			block++
			if block >= nBlocks {
				break
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
				runErr = fmt.Errorf("error: stochastic replay failed.")
			}

			runErr = fmt.Errorf("%v\n\tBlock %v Tx %v: %v", runErr, ss.blockNum, ss.txNum, err)
			if !cfg.ContinueOnFailure {
				break
			}
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
	for op := 0; op < NumOps; op++ {
		log.Noticef("\t%v: %v", opText[op], opFrequency[op])
	}
	return runErr
}

// execute StateDB operations on a stochastic state.
func  execute(file FileHandler, op int, addr *common.address, keyCl *common.hash, value *common.hash) {

	switch op {
	case AddBalanceID:
		value := file.readValue()
		reason := file.readReason()
		db.AddBalance(addr, value, reason)

	case BeginBlockID:
		if ss.traceDebug {
			ss.log.Infof(" id: %v", ss.blockNum)
		}
		db.BeginBlock(ss.blockNum)
		ss.txNum = 0
		ss.selfDestructed = []int64{}

	case BeginSyncPeriodID:
		if ss.traceDebug {
			ss.log.Infof(" id: %v", ss.syncPeriodNum)
		}
		db.BeginSyncPeriod(ss.syncPeriodNum)

	case BeginTransactionID:
		if ss.traceDebug {
			ss.log.Infof(" id: %v", ss.txNum)
		}
		db.BeginTransaction(ss.txNum)
		ss.snapshot = []int{}
		ss.selfDestructed = []int64{}

	case CreateAccountID:
		db.CreateAccount(addr)

	case CreateContractID:
		db.CreateContract(addr)

	case EmptyID:
		db.Empty(addr)

	case EndBlockID:
		db.EndBlock()
		ss.blockNum++
		ss.deleteAccounts()

	case EndSyncPeriodID:
		db.EndSyncPeriod()
		ss.syncPeriodNum++

	case EndTransactionID:
		db.EndTransaction()
		ss.txNum++
		ss.totalTx++

	case ExistID:
		db.Exist(addr)

	case GetBalanceID:
		db.GetBalance(addr)

	case GetCodeHashID:
		db.GetCodeHash(addr)

	case GetCodeID:
		db.GetCode(addr)

	case GetCodeSizeID:
		db.GetCodeSize(addr)

	case GetCommittedStateID:
		db.GetCommittedState(addr, key)

	case GetNonceID:
		db.GetNonce(addr)

	case GetStateID:
		db.GetState(addr, key)

	case GetStorageRootID:
		db.GetStorageRoot(addr)

	case HasSelfDestructedID:
		db.HasSelfDestructed(addr)

	case RevertToSnapshotID:
		snapshotNum := len(ss.snapshot)
		if snapshotNum > 0 {
			// TODO: consider a more realistic distribution
			// rather than the uniform distribution.
			snapshotIdx := snapshotNum - int(exponential.DiscreteSample(rg, ss.snapshotLambda, int64(snapshotNum))) - 1
			if snapshotIdx < 0 {
				snapshotIdx = 0
			} else if snapshotIdx >= snapshotNum {
				snapshotIdx = snapshotNum - 1
			}
			snapshot := ss.snapshot[snapshotIdx]
			if ss.traceDebug {
				ss.log.Infof(" id: %v", snapshot)
			}
			ss.snapshot = ss.snapshot[:snapshotIdx+1]

			// update active snapshots and perform a rollback in balance log
			db.RevertToSnapshot(snapshot)
		}

	case SetCodeID:
		sz := rg.Intn(MaxCodeSize-1) + 1
		if ss.traceDebug {
			ss.log.Infof(" code-size: %v", sz)
		}
		code := make([]byte, sz)
		_, err := rg.Read(code)
		if err != nil {
			ss.log.Fatalf("error producing a random byte slice. Error: %v", err)
		}
		db.SetCode(addr, code)

	case SetNonceID:
		value := uint64(rg.Intn(NonceRange))
		db.SetNonce(addr, value, tracing.NonceChangeUnspecified)

	case SetStateID:
		db.SetState(addr, key, value)

	case SnapshotID:
		id := db.Snapshot()
		if ss.traceDebug {
			ss.log.Infof(" id: %v", id)
		}
		ss.snapshot = append(ss.snapshot, id)

	case SubBalanceID:
		shadowDB := db.GetShadowDB()
		var balance uint64
		if shadowDB == nil {
			balance = db.GetBalance(addr).Uint64()
		} else {
			balance = shadowDB.GetBalance(addr).Uint64()
		}
		if balance > 0 {
			// get a delta that does not exceed current balance
			// in the current snapshot
			value := uint64(rg.Int63n(int64(balance)))
			if ss.traceDebug {
				ss.log.Infof(" value: %v", value)
			}
			db.SubBalance(addr, uint256.NewInt(value), 0)
		}

	case SelfDestructID:
		db.SelfDestruct(addr)
		if idx := find(ss.selfDestructed, addrIdx); idx == -1 {
			ss.selfDestructed = append(ss.selfDestructed, addrIdx)
		}

	case SelfDestruct6780ID:
		db.SelfDestruct6780(addr)
		if idx := find(ss.selfDestructed, addrIdx); idx == -1 {
			ss.selfDestructed = append(ss.selfDestructed, addrIdx)
		}

	default:
		ss.log.Fatal("invalid operation")
	}
}

// nextState produces the next state in the Markovian process.
func nextState(rg *rand.Rand, A [][]float64, i int) int {
	// Retrieve a random number in [0,1.0).
	r := rg.Float64()

	// Use Kahan's sum for summing values
	// in case we have a combination of very small
	// and very large values.
	sum := float64(0.0)
	c := float64(0.0)
	k := -1
	for j := 0; j < len(A); j++ {
		y := A[i][j] - c
		t := sum + y
		c = (t - sum) - y
		sum = t
		if r <= sum {
			return j
		}
		// If we have a numerical unstable cumulative
		// distribution (large and small numbers that cancel
		// each other out when summing up), we can take the last
		// non-zero entry as a solution. It also detects
		// stochastic matrices with a row whose row
		// sum is not zero (return value is -1 for such a case).
		if A[i][j] > 0.0 {
			k = j
		}
	}
	return k
}


//
//type operationProcessor struct {
//	cfg  *utils.Config
//	rCtx *context.Replay
//}
//
//func (p operationProcessor) Process(state executor.State[[]operation.Operation], ctx *executor.Context) error {
//	p.runTransaction(uint64(state.Block), state.Data, ctx.State)
//	return nil
//}
//
//func (p operationProcessor) runTransaction(block uint64, operations []operation.Operation, stateDb state.StateDB) {
//	for _, op := range operations {
//		operation.Execute(op, stateDb, p.rCtx)
//		if p.cfg.Debug && block >= p.cfg.DebugFrom {
//			operation.Debug(&p.rCtx.Context, op)
//		}
//	}
//}
//
//func replay(cfg *utils.Config, provider executor.Provider[[]operation.Operation], processor executor.Processor[[]operation.Operation], extra []executor.Extension[[]operation.Operation], aidaDb db.BaseDB) error {
//	var extensionList = []executor.Extension[[]operation.Operation]{
//		profiler.MakeCpuProfiler[[]operation.Operation](cfg),
//		statedb.MakeStateDbManager[[]operation.Operation](cfg, ""),
//		profiler.MakeMemoryUsagePrinter[[]operation.Operation](cfg),
//		profiler.MakeMemoryProfiler[[]operation.Operation](cfg),
//		logger.MakeProgressLogger[[]operation.Operation](cfg, 0),
//		primer.MakeStateDbPrimer[[]operation.Operation](cfg),
//	}
//
//	extensionList = append(extensionList, extra...)
//
//	return executor.NewExecutor(provider, cfg.LogLevel).Run(
//		executor.Params{
//			From: int(cfg.First),
//			To:   int(cfg.Last) + 1,
//		},
//		processor,
//		extensionList,
//		aidaDb,
//	)
//}
