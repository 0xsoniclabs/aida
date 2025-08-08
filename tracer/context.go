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

package tracer

import (
	"fmt"

	"github.com/0xsoniclabs/aida/utils/bigendian"
	"github.com/ethereum/go-ethereum/common"
)

//go:generate mockgen -source context.go -destination context_mock.go -package tracer

type Context interface {
	WriteOp(op uint8, data []byte) error
	WriteAddressOp(op uint8, address *common.Address, data []byte) error
	WriteKeyOp(op uint8, address *common.Address, key *common.Hash, data []byte) error
	WriteValueOp(op uint8, address *common.Address, key *common.Hash, value *common.Hash) error
	Close() error
}

// argumentContext keeps track of previously used arguments in StateDB operations
type argumentContext struct {
	// Contract-address queue
	contracts Queue[common.Address]
	// Storage-key queue
	keys Queue[common.Hash]
	// Storage-value queue
	values Queue[common.Hash]
	file   Writer
}

// NewContext creates a new event registry.
func NewContext(file Writer, first uint64, last uint64) (Context, error) {
	ctx := &argumentContext{
		contracts: NewQueue[common.Address](),
		keys:      NewQueue[common.Hash](),
		values:    NewQueue[common.Hash](),
		file:      file,
	}

	err := ctx.writeMetadata(first, last)
	if err != nil {
		return nil, err
	}

	return ctx, nil
}

// WriteOp registers an operation with no simulation arguments
func (ctx *argumentContext) WriteOp(op uint8, data []byte) error {
	argOp, err := EncodeArgOp(op, NoArgID, NoArgID, NoArgID)
	if err != nil {
		return err
	}
	if err = ctx.file.WriteUint16(argOp); err != nil {
		return err
	}
	if err = ctx.file.WriteData(data); err != nil {
		return fmt.Errorf("failed to write operation data: %w", err)
	}
	return nil
}

// WriteAddressOp registers an operation with a contract-address argument
func (ctx *argumentContext) WriteAddressOp(op uint8, address *common.Address, data []byte) error {
	addrClass, addrIdx := ctx.contracts.Classify(*address) // zero, previous, recent, address

	argOp, err := EncodeArgOp(op, addrClass, NoArgID, NoArgID)
	if err != nil {
		return err
	}
	// Write the operation code with argument classifications
	if err = ctx.file.WriteUint16(argOp); err != nil {
		return fmt.Errorf("failed to write addr encoded arg operation: %w", err)
	}

	// Write the address argument
	if err = ctx.writeClassifiedOp(addrClass, addrIdx, address); err != nil {
		return fmt.Errorf("failed to write addr operation: %w", err)
	}

	// Write the data argument
	if err = ctx.file.WriteData(data); err != nil {
		return err
	}
	return nil
}

// WriteKeyOp registers an operation with a contract-address and a storage-key arguments.
func (ctx *argumentContext) WriteKeyOp(op uint8, address *common.Address, key *common.Hash, data []byte) error {
	addrClass, addrIdx := ctx.contracts.Classify(*address)
	keyClass, keyIdx := ctx.keys.Classify(*key)

	argOp, err := EncodeArgOp(op, addrClass, keyClass, NoArgID)
	if err != nil {
		return err
	}
	// Write the operation code with argument classifications
	if err = ctx.file.WriteUint16(argOp); err != nil {
		return fmt.Errorf("failed to write key encoded arg operation: %w", err)
	}

	// Write the address and key arguments
	if err = ctx.writeClassifiedOp(addrClass, addrIdx, address); err != nil {
		return fmt.Errorf("failed to write addr operation: %w", err)
	}
	if err = ctx.writeClassifiedOp(keyClass, keyIdx, key); err != nil {
		return fmt.Errorf("failed to write key operation: %w", err)
	}

	// Write the data argument
	if err = ctx.file.WriteData(data); err != nil {
		return err
	}
	return nil
}

// WriteValueOp registers an operation with a contract-address, a storage-key and storage-value arguments.
func (ctx *argumentContext) WriteValueOp(op uint8, address *common.Address, key *common.Hash, value *common.Hash) error {
	addrClass, addrIdx := ctx.contracts.Classify(*address)
	keyClass, keyIdx := ctx.keys.Classify(*key)
	valueClass, valueIdx := ctx.values.Classify(*value)

	argOp, err := EncodeArgOp(op, addrClass, keyClass, valueClass)
	if err != nil {
		return err
	}
	if err = ctx.file.WriteUint16(argOp); err != nil {
		return fmt.Errorf("failed to write value encoded arg operation: %w", err)
	}

	// Write the address, key and value arguments
	if err = ctx.writeClassifiedOp(addrClass, addrIdx, address); err != nil {
		return fmt.Errorf("failed to write addr operation: %w", err)
	}
	if err = ctx.writeClassifiedOp(keyClass, keyIdx, key); err != nil {
		return fmt.Errorf("failed to write key operation: %w", err)
	}
	if err = ctx.writeClassifiedOp(valueClass, valueIdx, value); err != nil {
		return fmt.Errorf("failed to write value operation: %w", err)
	}
	return nil
}

func (ctx *argumentContext) Close() error {
	return ctx.file.Close()
}

func (ctx *argumentContext) writeClassifiedOp(class uint8, idx int, data Byter) error {
	switch class {
	case ZeroValueID:
	case PreviousValueID:
	case RecentValueID:
		if err := ctx.file.WriteUint8(uint8(idx)); err != nil {
			return err
		}
	case NewValueID:
		if err := ctx.file.WriteData(data.Bytes()); err != nil {
			return err
		}
	default:
		return fmt.Errorf("unexpected argument classification: %d", class)
	}
	return nil
}

// writeMetadata writes the metadata into the file. This should be called once and only at the beginning of the file.
func (ctx *argumentContext) writeMetadata(first uint64, last uint64) error {
	data := append(bigendian.Uint64ToBytes(first), bigendian.Uint64ToBytes(last)...)
	return ctx.file.WriteData(data)
}

type Byter interface {
	Bytes() []byte
}
