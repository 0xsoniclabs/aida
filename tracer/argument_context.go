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
	"github.com/ethereum/go-ethereum/common"
)

// ArgumentContext keeps track of previously used arguments in StateDB operations
type ArgumentContext struct {
	// Contract-address queue
	contracts Queue[common.Address]

	// Storage-key queue
	keys Queue[common.Hash]

	// Storage-value queue
	values Queue[common.Hash]

	file FileHandler
}

// NewArgumentContext creates a new event registry.
func NewArgumentContext(file FileHandler) ArgumentContext {
	return ArgumentContext{
		contracts: NewQueue[common.Address](),
		keys:      NewQueue[common.Hash](),
		values:    NewQueue[common.Hash](),
		file:      file,
	}
}

// WriteOp registers an operation with no simulation arguments
func (ctx *ArgumentContext) WriteOp(op uint16, data []byte) error {
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
func (ctx *ArgumentContext) WriteAddressOp(op uint16, address *common.Address, data []byte) error {
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
func (ctx *ArgumentContext) WriteKeyOp(op uint16, address *common.Address, key *common.Hash, data []byte) error {
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
func (ctx *ArgumentContext) WriteValueOp(op uint16, address *common.Address, key *common.Hash, value *common.Hash) error {
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

func (ctx *ArgumentContext) Close() error {
	return ctx.file.Close()
}

func (ctx *ArgumentContext) writeClassifiedOp(class uint8, idx int, data Byter) error {
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
	}
	return fmt.Errorf("unexpected argument classification: %d", class)
}

type Byter interface {
	Bytes() []byte
}
