package executor

import (
	"fmt"
	"github.com/0xsoniclabs/aida/state"
	"github.com/0xsoniclabs/aida/tracer"
	"github.com/cockroachdb/errors"
	"github.com/ethereum/go-ethereum/common"
)

type traceProvider struct {
	file      tracer.FileReader
	contracts tracer.Queue[common.Address]
	keys      tracer.Queue[common.Hash]
	values    tracer.Queue[common.Hash]
	db        state.StateDB
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
			return fmt.Errorf("cannot read operation from file: %w", err)
		}
		operation, err := p.readOperation(argOp)
		if err != nil {
			return err
		}

		// We need to read data of BeginBlock and BeginTransaction beforehand
		switch operation.Op {
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
		case tracer.BeginTransactionID:
			tx, err = p.file.ReadUint32()
			if err != nil {
				return fmt.Errorf("cannot read transaction number: %w", err)
			}
		default:
			// do nothing
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
	case tracer.NewValueID:
		addr, err = p.file.ReadAddr()
		if err != nil {
			return tracer.Operation{}, fmt.Errorf("cannot read address from file: %w", err)
		}
		p.contracts.Place(addr)
	case tracer.PreviousValueID:
		addr, err = p.contracts.Get(1)
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
		p.contracts.Place(addr)
	default:
		return tracer.Operation{}, fmt.Errorf("wrong address class: %d", addrCl)
	}

	switch keyCl {
	case tracer.NoArgID:
	case tracer.NewValueID:
		key, err = p.file.ReadHash()
		if err != nil {
			return tracer.Operation{}, fmt.Errorf("cannot read key hash from file: %w", err)
		}
		p.keys.Place(key)
	case tracer.PreviousValueID:
		key, err = p.keys.Get(1)
		if err != nil {
			return tracer.Operation{}, fmt.Errorf("cannot get previous key hash from contracts queue: %w", err)
		}
	case tracer.RecentValueID:
		idx, err := p.file.ReadUint8()
		if err != nil {
			return tracer.Operation{}, fmt.Errorf("cannot read addrekey hash idx from file: %w", err)
		}
		key, err = p.keys.Get(int(idx))
		if err != nil {
			return tracer.Operation{}, fmt.Errorf("cannot get recent key hash from contracts queue: %w", err)
		}
		p.keys.Place(key)
	default:
		return tracer.Operation{}, fmt.Errorf("wrong key class: %d", keyCl)
	}

	switch valueCl {
	case tracer.NoArgID:
	case tracer.NewValueID:
		val, err = p.file.ReadHash()
		if err != nil {
			return tracer.Operation{}, fmt.Errorf("cannot read val hash from file: %w", err)
		}
		p.values.Place(val)
	case tracer.PreviousValueID:
		val, err = p.values.Get(1)
		if err != nil {
			return tracer.Operation{}, fmt.Errorf("cannot get previous val hash from contracts queue: %w", err)
		}
	case tracer.RecentValueID:
		idx, err := p.file.ReadUint8()
		if err != nil {
			return tracer.Operation{}, fmt.Errorf("cannot read addrekey val idx from file: %w", err)
		}
		val, err = p.values.Get(int(idx))
		if err != nil {
			return tracer.Operation{}, fmt.Errorf("cannot get recent key val from contracts queue: %w", err)
		}
		p.values.Place(val)
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
