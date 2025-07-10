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
	"log"

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

	// File handler
	// TBD: Use old file procedures (gzip etc.)
	file *FileHandler
}

// NewArgumentContext creates a new event registry.
func NewArgumentContext(file *FileHandler) ArgumentContext {
	return ArgumentContext{
		contracts: NewQueue[common.Address](),
		keys:      NewQueue[common.Hash](),
		values:    NewQueue[common.Hash](),
		file:      file,
	}
}

// WriteOp registers an operation with no simulation arguments
func (ctx *ArgumentContext) WriteOp(op uint16, data []byte) {
	if op < 0 || op >= NumOps {
		log.Fatalf("invalid operation ID")
	}
	addrClass := NoArgID
	keyClass := NoArgID
	valueClass := NoArgID

	argOp := EncodeArgOp(op, addrClass, keyClass, valueClass)

	ctx.file.WriteUint16(argOp)
	ctx.file.WriteData(data)
}

// WriteAddressOp registers an operation with a contract-address argument
func (ctx *ArgumentContext) WriteAddressOp(op uint16, address *common.Address, data []byte) {
	if op < 0 || op >= NumOps {
		log.Fatalf("invalid stochastic operation ID")
	}

	addrClass, addrIdx := ctx.contracts.Classify(*address) // zero, previous, recent, address

	argOp := EncodeArgOp(op, addrClass, NoArgID, NoArgID)
	ctx.file.WriteUint16(argOp)

	switch addrClass {
	case ZeroValueID:
	case PreviousValueID:
	case RecentValueID:
		ctx.file.WriteUint8(uint8(addrIdx))
	case NewValueID:
		ctx.file.WriteData(address.Bytes())
	default:
		panic("Unexpected argument classification")
	}

	ctx.file.WriteData(data)
}

// WriteKeyOp registers an operation with a contract-address and a storage-key arguments.
func (ctx *ArgumentContext) WriteKeyOp(op uint16, address *common.Address, key *common.Hash, data []byte) {
	if op < 0 || op >= NumOps {
		log.Fatalf("invalid stochastic operation ID")
	}

	addrClass, addrIdx := ctx.contracts.Classify(*address)
	keyClass, keyIdx := ctx.keys.Classify(*key)

	argOp := EncodeArgOp(op, addrClass, keyClass, NoArgID)
	ctx.file.WriteUint16(argOp)

	switch addrClass {
	case ZeroValueID:
	case PreviousValueID:
	case RecentValueID:
		ctx.file.WriteUint8(uint8(addrIdx))
	case NewValueID:
		ctx.file.WriteData(address.Bytes())
	default:
		panic("Unexpected argument classification")
	}

	switch keyClass {
	case ZeroValueID:
	case PreviousValueID:
	case RecentValueID:
		ctx.file.WriteUint8(uint8(keyIdx))
	case NewValueID:
		ctx.file.WriteData(key.Bytes())
	default:
		panic("Unexpected argument classification")
	}

	ctx.file.WriteData(data)
}

// WriteValueOp registers an operation with a contract-address, a storage-key and storage-value arguments.
func (ctx *ArgumentContext) WriteValueOp(op uint16, address *common.Address, key *common.Hash, value *common.Hash) {
	if op < 0 || op >= NumOps {
		log.Fatalf("invalid stochastic operation ID")
	}

	addrClass, addrIdx := ctx.contracts.Classify(*address)
	keyClass, keyIdx := ctx.keys.Classify(*key)
	valueClass, valueIdx := ctx.values.Classify(*value)

	argOp := EncodeArgOp(op, addrClass, keyClass, valueClass)
	ctx.file.WriteUint16(argOp)

	switch addrClass {
	case ZeroValueID:
	case PreviousValueID:
	case RecentValueID:
		ctx.file.WriteUint8(uint8(addrIdx))
	case NewValueID:
		ctx.file.WriteData(address.Bytes())
	default:
		panic("Unexpected argument classification")
	}

	switch keyClass {
	case ZeroValueID:
	case PreviousValueID:
	case RecentValueID:
		ctx.file.WriteUint8(uint8(keyIdx))
	case NewValueID:
		ctx.file.WriteData(key.Bytes())
	default:
		panic("Unexpected key classification")
	}

	switch valueClass {
	case ZeroValueID:
	case PreviousValueID:
	case RecentValueID:
		ctx.file.WriteUint8(uint8(valueIdx))
	case NewValueID:
		ctx.file.WriteData(value.Bytes())
	default:
		panic("Unexpected value classification")
	}
}

func (ctx *ArgumentContext) Close() error {
	return ctx.file.Close()
}
