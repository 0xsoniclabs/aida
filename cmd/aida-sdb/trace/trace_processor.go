package trace

import (
	"errors"
	"github.com/0xsoniclabs/aida/executor"
	"github.com/0xsoniclabs/aida/tracer"
	"github.com/0xsoniclabs/aida/utils"
	"github.com/ethereum/go-ethereum/common"
)

type traceProcessor struct {
	cfg  *utils.Config
	file tracer.FileReader
}

func (p *traceProcessor) Process(state executor.State[tracer.Operation], ctx *executor.Context) error {
	switch state.Data.Op {
	case tracer.AddBalanceID:
		value, reason, err := p.file.ReadBalanceChange()
		if err != nil {
			return err
		}
		ctx.State.AddBalance(state.Data.Addr, value, reason)
	case tracer.BeginBlockID:
		if err := ctx.State.BeginBlock(uint64(state.Block)); err != nil {
			return err
		}
	case tracer.BeginSyncPeriodID:
		val, err := p.file.ReadUint64()
		if err != nil {
			return err
		}
		ctx.State.BeginSyncPeriod(val)
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
		value, err := p.file.ReadUint32()
		if err != nil {
			return err
		}
		ctx.State.RevertToSnapshot(int(value))
	case tracer.SetCodeID:
		code, err := p.file.ReadVariableSizeData()
		if err != nil {
			return err
		}
		ctx.State.SetCode(state.Data.Addr, code)
	case tracer.SetNonceID:
		value, reason, err := p.file.ReadNonceChange()
		if err != nil {
			return err
		}
		ctx.State.SetNonce(state.Data.Addr, value, reason)
	case tracer.SetStateID:
		ctx.State.SetState(state.Data.Addr, state.Data.Key, state.Data.Value)
	case tracer.SnapshotID:
		ctx.State.Snapshot()
	case tracer.SubBalanceID:
		value, reason, err := p.file.ReadBalanceChange()
		if err != nil {
			return err
		}
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
		ctx.State.GetTransientState(state.Data.Addr, state.Data.Value)
	case tracer.SelfDestruct6780ID:
		ctx.State.SelfDestruct6780(state.Data.Addr)
	case tracer.SubRefundID:
		refund, err := p.file.ReadUint64()
		if err != nil {
			return err
		}
		ctx.State.SubRefund(refund)
	case tracer.GetRefundID:
		ctx.State.GetRefund()
	case tracer.AddRefundID:
		refund, err := p.file.ReadUint64()
		if err != nil {
			return err
		}
		ctx.State.AddRefund(refund)
	case tracer.PrepareID:
		rules, err := p.file.ReadRules()
		if err != nil {
			return err
		}
		sender, err := p.file.ReadAddr()
		if err != nil {
			return err
		}
		coinbase, err := p.file.ReadAddr()
		if err != nil {
			return err
		}
		dest, err := p.file.ReadAddr()
		if err != nil {
			return err
		}
		numAddr, err := p.file.ReadUint32()
		if err != nil {
			return err
		}
		precompiles := make([]common.Address, numAddr)
		for idx := range numAddr {
			addr, err := p.file.ReadAddr()
			if err != nil {
				return err
			}
			precompiles[idx] = addr
		}
		accessList, err := p.file.ReadAccessList()
		if err != nil {
			return err
		}
		ctx.State.Prepare(rules, sender, coinbase, &dest, precompiles, accessList)
	case tracer.AddAddressToAccessListID:
		ctx.State.AddAddressToAccessList(state.Data.Addr)
	case tracer.AddressInAccessListID:
		ctx.State.AddressInAccessList(state.Data.Addr)
	case tracer.SlotInAccessListID:
		ctx.State.SlotInAccessList(state.Data.Addr, state.Data.Key)
	case tracer.AddSlotToAccessListID:
		ctx.State.AddSlotToAccessList(state.Data.Addr, state.Data.Key)
	case tracer.AddLogID:
	case tracer.GetLogsID:
		hash, err := p.file.ReadHash()
		if err != nil {
			return err
		}
		blk, err := p.file.ReadUint64()
		if err != nil {
			return err
		}
		blkHash, err := p.file.ReadHash()
		if err != nil {
			return err
		}
		timestamp, err := p.file.ReadUint64()
		if err != nil {
			return err
		}
		ctx.State.GetLogs(hash, blk, blkHash, timestamp)
	case tracer.PointCacheID:
		ctx.State.PointCache()
	case tracer.WitnessID:
		ctx.State.Witness()
	case tracer.AddPreimageID:
		hash, err := p.file.ReadHash()
		if err != nil {
			return err
		}
		image, err := p.file.ReadVariableSizeData()
		if err != nil {
			return err
		}
		ctx.State.AddPreimage(hash, image)
	case tracer.SetTxContextID:
		hash, err := p.file.ReadHash()
		if err != nil {
			return err
		}
		txIndex, err := p.file.ReadUint32()
		if err != nil {
			return err
		}
		ctx.State.SetTxContext(hash, int(txIndex))
	case tracer.FinaliseID:
		b, err := p.file.ReadBool()
		if err != nil {
			return err
		}
		ctx.State.Finalise(b)
	case tracer.IntermediateRootID:
		b, err := p.file.ReadBool()
		if err != nil {
			return err
		}
		ctx.State.IntermediateRoot(b)
	case tracer.CommitID:
		b, err := p.file.ReadBool()
		if err != nil {
			return err
		}
		blk, err := p.file.ReadUint64()
		if err != nil {
			return err
		}
		if _, err = ctx.State.Commit(blk, b); err != nil {
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
		// ignored
		ctx.State.PrepareSubstate(nil, 0)
	default:
		return errors.New("invalid operation")
	}
	return nil
}
