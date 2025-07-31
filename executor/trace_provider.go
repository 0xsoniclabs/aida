package executor

import (
	"fmt"
	"github.com/0xsoniclabs/aida/tracer"
	"github.com/cockroachdb/errors"
	"github.com/ethereum/go-ethereum/common"
	"io"
)

func NewTraceProvider(file tracer.FileReader) Provider[tracer.Operation] {
	return &traceProvider{
		file:      file,
		contracts: tracer.NewQueue[common.Address](),
		keys:      tracer.NewQueue[common.Hash](),
		values:    tracer.NewQueue[common.Hash](),
	}
}

type traceProvider struct {
	file                                tracer.FileReader
	contracts                           tracer.Queue[common.Address]
	keys                                tracer.Queue[common.Hash]
	values                              tracer.Queue[common.Hash]
	prevAddrIdx, prevKeyIdx, prevValIdx int
}

func (p *traceProvider) Run(from int, to int, consumer Consumer[tracer.Operation]) (err error) {
	defer func() {
		err = errors.Join(err, p.file.Close())
	}()
	var (
		blk uint64
		tx  uint32
	)
	for {
		// read 16-bit operation number
		argOp, err := p.file.ReadUint16()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return nil
			}
			return fmt.Errorf("cannot read operation from file: %w", err)
		}
		operation, err := p.readOperation(argOp)
		if err != nil {
			return err
		}

		switch operation.Op {
		case tracer.AddBalanceID:
			value, reason, err := p.file.ReadBalanceChange()
			if err != nil {
				return err
			}
			operation.Data = []any{value, reason}
		case tracer.BeginBlockID:
			blk, err = p.file.ReadUint64()
			if err != nil {
				return fmt.Errorf("cannot read block number: %w", err)
			}
			if blk < uint64(from) {
				continue
			}
			if blk > uint64(to) {
				return nil
			}
			// reset tx number to 0
			tx = 0
		case tracer.BeginSyncPeriodID:
			value, err := p.file.ReadUint64()
			if err != nil {
				return err
			}
			operation.Data = []any{value}
		case tracer.BeginTransactionID:
			tx, err = p.file.ReadUint32()
			if err != nil {
				return fmt.Errorf("cannot read transaction number: %w", err)
			}
		case tracer.CreateAccountID:
		case tracer.EmptyID:
		case tracer.EndBlockID:
		case tracer.EndSyncPeriodID:
		case tracer.EndTransactionID:
		case tracer.ExistID:
		case tracer.GetBalanceID:
		case tracer.GetCodeHashID:
		case tracer.GetCodeID:
		case tracer.GetCodeSizeID:
		case tracer.GetCommittedStateID:
		case tracer.GetNonceID:
		case tracer.GetStateID:
		case tracer.HasSelfDestructedID:
		case tracer.RevertToSnapshotID:
			value, err := p.file.ReadUint32()
			if err != nil {
				return err
			}
			operation.Data = []any{value}
		case tracer.SetCodeID:
			code, err := p.file.ReadVariableSizeData()
			if err != nil {
				return err
			}
			operation.Data = []any{code}
		case tracer.SetNonceID:
			value, reason, err := p.file.ReadNonceChange()
			if err != nil {
				return err
			}
			operation.Data = []any{value, reason}
		case tracer.SetStateID:
		case tracer.SnapshotID:
		case tracer.SubBalanceID:
			value, reason, err := p.file.ReadBalanceChange()
			if err != nil {
				return err
			}
			operation.Data = []any{value, reason}
		case tracer.SelfDestructID:
		case tracer.CreateContractID:
		case tracer.GetStorageRootID:
		case tracer.GetTransientStateID:
		case tracer.SetTransientStateID:
		case tracer.SelfDestruct6780ID:
		case tracer.SubRefundID:
			refund, err := p.file.ReadUint64()
			if err != nil {
				return err
			}
			operation.Data = []any{refund}
		case tracer.GetRefundID:
		case tracer.AddRefundID:
			refund, err := p.file.ReadUint64()
			if err != nil {
				return err
			}
			operation.Data = []any{refund}
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
			isDestPresent, err := p.file.ReadBool()
			if err != nil {
				return err
			}
			var dest *common.Address
			if isDestPresent {
				a, err := p.file.ReadAddr()
				if err != nil {
					return err
				}
				dest = &a
			}
			numAddr, err := p.file.ReadUint16()
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
			operation.Data = []any{rules, sender, coinbase, dest, precompiles, accessList}
		case tracer.AddAddressToAccessListID:
		case tracer.AddressInAccessListID:
		case tracer.SlotInAccessListID:
		case tracer.AddSlotToAccessListID:
		case tracer.AddLogID:
			log, err := p.file.ReadLog()
			if err != nil {
				return err
			}
			operation.Data = []any{log}
		case tracer.GetLogsID:
			hash, err := p.file.ReadHash()
			if err != nil {
				return err
			}
			blkNum, err := p.file.ReadUint64()
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
			operation.Data = []any{hash, blkNum, blkHash, timestamp}
		case tracer.PointCacheID:
		case tracer.WitnessID:
		case tracer.AddPreimageID:
			hash, err := p.file.ReadHash()
			if err != nil {
				return err
			}
			image, err := p.file.ReadVariableSizeData()
			if err != nil {
				return err
			}
			operation.Data = []any{hash, image}
		case tracer.SetTxContextID:
			hash, err := p.file.ReadHash()
			if err != nil {
				return err
			}
			txIndex, err := p.file.ReadUint32()
			if err != nil {
				return err
			}
			operation.Data = []any{hash, txIndex}
		case tracer.FinaliseID:
			b, err := p.file.ReadBool()
			if err != nil {
				return err
			}
			operation.Data = []any{b}
		case tracer.IntermediateRootID:
			b, err := p.file.ReadBool()
			if err != nil {
				return err
			}
			operation.Data = []any{b}
		case tracer.CommitID:
			b, err := p.file.ReadBool()
			if err != nil {
				return err
			}
			blkNum, err := p.file.ReadUint64()
			if err != nil {
				return err
			}
			operation.Data = []any{blkNum, b}
		case tracer.CloseID:
		case tracer.AccessEventsID:
		case tracer.GetHashID:
		case tracer.GetSubstatePostAllocID:
		case tracer.PrepareSubstateID:
			ws, err := p.file.ReadWorldState()
			if err != nil {
				return err
			}
			operation.Data = []any{ws}
		case tracer.GetArchiveStateID:
			blkNum, err := p.file.ReadUint64()
			if err != nil {
				return err
			}
			operation.Data = []any{blkNum}
		case tracer.GetArchiveBlockHeightID:
		default:
			return fmt.Errorf("invalid operation %d/%s", operation.Op, tracer.OpText[operation.Op])
		}

		err = consumer(TransactionInfo[tracer.Operation]{
			Block:       int(blk),
			Transaction: int(tx),
			Data:        operation,
		})
		if err != nil {
			return err
		}
	}

}

// readOperation decodes argOp and reads all necessary arguments from the file.
func (p *traceProvider) readOperation(argOp uint16) (tracer.Operation, error) {
	var (
		addr common.Address
		key  common.Hash
		val  common.Hash
	)
	// readOperation opcode
	op, addrCl, keyCl, valueCl, err := tracer.DecodeArgOp(argOp)
	if err != nil {
		return tracer.Operation{}, fmt.Errorf("cannot decode operation: %w", err)
	}

	switch addrCl {
	case tracer.NoArgID:
	case tracer.ZeroValueID:
	case tracer.NewValueID:
		addr, err = p.file.ReadAddr()
		if err != nil {
			return tracer.Operation{}, fmt.Errorf("cannot read address from file: %w", err)
		}
		p.contracts.Place(addr)
		p.prevAddrIdx = p.contracts.Find(addr)
	case tracer.PreviousValueID:
		addr, err = p.contracts.Get(p.prevAddrIdx)
		if err != nil {
			return tracer.Operation{}, fmt.Errorf("cannot get previous address from contracts queue: %w", err)
		}
	case tracer.RecentValueID:
		idx, err := p.file.ReadUint8()
		if err != nil {
			return tracer.Operation{}, fmt.Errorf("cannot read address idx from file: %w", err)
		}
		addr, err = p.contracts.Get(int(idx))
		if err != nil {
			return tracer.Operation{}, fmt.Errorf("cannot get recent address from contracts queue: %w", err)
		}
	default:
		return tracer.Operation{}, fmt.Errorf("wrong address class: %d", addrCl)
	}

	switch keyCl {
	case tracer.NoArgID:
	case tracer.ZeroValueID:
	case tracer.NewValueID:
		key, err = p.file.ReadHash()
		if err != nil {
			return tracer.Operation{}, fmt.Errorf("cannot read key hash from file: %w", err)
		}
		p.keys.Place(key)
		p.prevKeyIdx = p.keys.Find(key)
	case tracer.PreviousValueID:
		key, err = p.keys.Get(p.prevKeyIdx)
		if err != nil {
			return tracer.Operation{}, fmt.Errorf("cannot get previous key from keys queue: %w", err)
		}
	case tracer.RecentValueID:
		idx, err := p.file.ReadUint8()
		if err != nil {
			return tracer.Operation{}, fmt.Errorf("cannot read key idx from file: %w", err)
		}
		key, err = p.keys.Get(int(idx))
		if err != nil {
			return tracer.Operation{}, fmt.Errorf("cannot get recent key from keys queue: %w", err)
		}
	default:
		return tracer.Operation{}, fmt.Errorf("wrong key class: %d", keyCl)
	}

	switch valueCl {
	case tracer.NoArgID:
	case tracer.ZeroValueID:
	case tracer.NewValueID:
		val, err = p.file.ReadHash()
		if err != nil {
			return tracer.Operation{}, fmt.Errorf("cannot read value hash from file: %w", err)
		}
		p.values.Place(val)
		p.prevValIdx = p.values.Find(val)
	case tracer.PreviousValueID:
		val, err = p.values.Get(p.prevValIdx)
		if err != nil {
			return tracer.Operation{}, fmt.Errorf("cannot get previous value hash from values queue: %w", err)
		}
	case tracer.RecentValueID:
		idx, err := p.file.ReadUint8()
		if err != nil {
			return tracer.Operation{}, fmt.Errorf("cannot read value val idx from file: %w", err)
		}
		val, err = p.values.Get(int(idx))
		if err != nil {
			return tracer.Operation{}, fmt.Errorf("cannot get recent value from values queue: %w", err)
		}
	default:
		return tracer.Operation{}, fmt.Errorf("wrong value class: %d", valueCl)
	}

	return tracer.Operation{
		Op:    op,
		Addr:  addr,
		Key:   key,
		Value: val,
	}, nil
}

func (p *traceProvider) Close() {
}
