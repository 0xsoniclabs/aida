package main

import (
	"errors"
	"github.com/0xsoniclabs/aida/executor"
	"github.com/0xsoniclabs/aida/tracer"
	"github.com/0xsoniclabs/aida/txcontext"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/tracing"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/params"
	"github.com/holiman/uint256"
)

type traceProcessor struct {
	counter uint32
}

func (p *traceProcessor) Process(state executor.State[tracer.Operation], ctx *executor.Context) error {
	p.counter++
	switch state.Data.Op {
	case tracer.AddBalanceID:
		value, reason := state.Data.Data[0].(*uint256.Int), state.Data.Data[1].(tracing.BalanceChangeReason)
		ctx.State.AddBalance(state.Data.Addr, value, reason)
	case tracer.BeginBlockID:
		if err := ctx.State.BeginBlock(uint64(state.Block)); err != nil {
			return err
		}
	case tracer.BeginSyncPeriodID:
		ctx.State.BeginSyncPeriod(state.Data.Data[0].(uint64))
	case tracer.BeginTransactionID:
		if err := ctx.State.BeginTransaction(uint32(state.Transaction)); err != nil {
			return err
		}
	case tracer.CreateAccountID:
		ctx.State.CreateAccount(state.Data.Addr)
	case tracer.EmptyID:
		ctx.State.Empty(state.Data.Addr)
	case tracer.EndBlockID:
		if err := ctx.State.EndBlock(); err != nil {
			return err
		}
	case tracer.EndSyncPeriodID:
		ctx.State.EndSyncPeriod()
	case tracer.EndTransactionID:
		if err := ctx.State.EndTransaction(); err != nil {
			return err
		}
	case tracer.ExistID:
		ctx.State.Exist(state.Data.Addr)
	case tracer.GetBalanceID:
		ctx.State.GetBalance(state.Data.Addr)
	case tracer.GetCodeHashID:
		ctx.State.GetCodeHash(state.Data.Addr)
	case tracer.GetCodeID:
		ctx.State.GetCode(state.Data.Addr)
	case tracer.GetCodeSizeID:
		ctx.State.GetCodeSize(state.Data.Addr)
	case tracer.GetCommittedStateID:
		ctx.State.GetCommittedState(state.Data.Addr, state.Data.Key)
	case tracer.GetNonceID:
		ctx.State.GetNonce(state.Data.Addr)
	case tracer.GetStateID:
		ctx.State.GetState(state.Data.Addr, state.Data.Key)
	case tracer.HasSelfDestructedID:
		ctx.State.HasSelfDestructed(state.Data.Addr)
	case tracer.RevertToSnapshotID:
		ctx.State.RevertToSnapshot(int(state.Data.Data[0].(uint32)))
	case tracer.SetCodeID:
		ctx.State.SetCode(state.Data.Addr, state.Data.Data[0].([]byte))
	case tracer.SetNonceID:
		value, reason := state.Data.Data[0].(uint64), state.Data.Data[1].(tracing.NonceChangeReason)
		ctx.State.SetNonce(state.Data.Addr, value, reason)
	case tracer.SetStateID:
		ctx.State.SetState(state.Data.Addr, state.Data.Key, state.Data.Value)
	case tracer.SnapshotID:
		ctx.State.Snapshot()
	case tracer.SubBalanceID:
		value, reason := state.Data.Data[0].(*uint256.Int), state.Data.Data[1].(tracing.BalanceChangeReason)
		ctx.State.SubBalance(state.Data.Addr, value, reason)
	case tracer.SelfDestructID:
		ctx.State.SelfDestruct(state.Data.Addr)
	case tracer.CreateContractID:
		ctx.State.CreateContract(state.Data.Addr)
	case tracer.GetStorageRootID:
		ctx.State.GetStorageRoot(state.Data.Addr)
	case tracer.GetTransientStateID:
		ctx.State.GetTransientState(state.Data.Addr, state.Data.Key)
	case tracer.SetTransientStateID:
		ctx.State.SetTransientState(state.Data.Addr, state.Data.Key, state.Data.Value)
	case tracer.SelfDestruct6780ID:
		ctx.State.SelfDestruct6780(state.Data.Addr)
	case tracer.SubRefundID:
		value := state.Data.Data[0].(uint64)
		ctx.State.SubRefund(value)
	case tracer.GetRefundID:
		ctx.State.GetRefund()
	case tracer.AddRefundID:
		value := state.Data.Data[0].(uint64)
		ctx.State.AddRefund(value)
	case tracer.PrepareID:
		rules := state.Data.Data[0].(params.Rules)
		sender := state.Data.Data[1].(common.Address)
		coinbase := state.Data.Data[2].(common.Address)
		dest := state.Data.Data[3].(*common.Address)
		precompiles := state.Data.Data[4].([]common.Address)
		accessList := state.Data.Data[5].(types.AccessList)
		ctx.State.Prepare(rules, sender, coinbase, dest, precompiles, accessList)
	case tracer.AddAddressToAccessListID:
		ctx.State.AddAddressToAccessList(state.Data.Addr)
	case tracer.AddressInAccessListID:
		ctx.State.AddressInAccessList(state.Data.Addr)
	case tracer.SlotInAccessListID:
		ctx.State.SlotInAccessList(state.Data.Addr, state.Data.Key)
	case tracer.AddSlotToAccessListID:
		ctx.State.AddSlotToAccessList(state.Data.Addr, state.Data.Key)
	case tracer.AddLogID:
		log := state.Data.Data[0].(*types.Log)
		ctx.State.AddLog(log)
	case tracer.GetLogsID:
		hash, blk := state.Data.Data[0].(common.Hash), state.Data.Data[1].(uint64)
		blkHash, blkTimestamp := state.Data.Data[2].(common.Hash), state.Data.Data[3].(uint64)
		ctx.State.GetLogs(hash, blk, blkHash, blkTimestamp)
	case tracer.PointCacheID:
		ctx.State.PointCache()
	case tracer.WitnessID:
		ctx.State.Witness()
	case tracer.AddPreimageID:
		hash, image := state.Data.Data[0].(common.Hash), state.Data.Data[1].([]byte)
		ctx.State.AddPreimage(hash, image)
	case tracer.SetTxContextID:
		hash, txIndex := state.Data.Data[0].(common.Hash), state.Data.Data[1].(uint32)
		ctx.State.SetTxContext(hash, int(txIndex))
	case tracer.FinaliseID:
		b := state.Data.Data[0].(bool)
		ctx.State.Finalise(b)
	case tracer.IntermediateRootID:
		b := state.Data.Data[0].(bool)
		ctx.State.IntermediateRoot(b)
	case tracer.CommitID:
		blkNum, b := state.Data.Data[0].(uint64), state.Data.Data[1].(bool)
		if _, err := ctx.State.Commit(blkNum, b); err != nil {
			return err
		}
	case tracer.CloseID:
		if err := ctx.State.Close(); err != nil {
			return err
		}
	case tracer.AccessEventsID:
		ctx.State.AccessEvents()
	case tracer.GetHashID:
		if _, err := ctx.State.GetHash(); err != nil {
			return err
		}
	case tracer.GetSubstatePostAllocID:
		ctx.State.GetSubstatePostAlloc()
	case tracer.PrepareSubstateID:
		ws := state.Data.Data[0].(txcontext.AidaWorldState)
		ctx.State.PrepareSubstate(ws, uint64(state.Block))
	case tracer.GetArchiveStateID:
		blkNum := state.Data.Data[0].(uint64)
		_, err := ctx.State.GetArchiveState(blkNum)
		if err != nil {
			return err
		}
	case tracer.GetArchiveBlockHeightID:
		_, _, err := ctx.State.GetArchiveBlockHeight()
		if err != nil {
			return err
		}
	default:
		return errors.New("invalid operation")
	}
	return nil
}
