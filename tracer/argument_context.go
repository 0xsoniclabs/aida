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
func (ctx *ArgumentContext) WriteOp(op uint16, data []byte) {
	argOp := EncodeArgOp(op, NoArgID, NoArgID, NoArgID)
	ctx.file.WriteUint16(argOp)
	ctx.file.WriteData(data)
}

// WriteAddressOp registers an operation with a contract-address argument
func (ctx *ArgumentContext) WriteAddressOp(op uint16, address *common.Address, data []byte) {
	addrClass, addrIdx := ctx.contracts.Classify(*address) // zero, previous, recent, address
	ctx.writeClassifiedOp(addrClass, addrIdx, address)

	argOp := EncodeArgOp(op, addrClass, NoArgID, NoArgID)
	ctx.file.WriteUint16(argOp)

	ctx.file.WriteData(data)
}

// WriteKeyOp registers an operation with a contract-address and a storage-key arguments.
func (ctx *ArgumentContext) WriteKeyOp(op uint16, address *common.Address, key *common.Hash, data []byte) {
	addrClass, addrIdx := ctx.contracts.Classify(*address)
	keyClass, keyIdx := ctx.keys.Classify(*key)

	argOp := EncodeArgOp(op, addrClass, keyClass, NoArgID)
	// Write the operation code with argument classifications
	ctx.file.WriteUint16(argOp)

	// Write the address and key arguments
	ctx.writeClassifiedOp(addrClass, addrIdx, address)
	ctx.writeClassifiedOp(keyClass, keyIdx, key)

	// Write the data argument
	ctx.file.WriteData(data)
}

// WriteValueOp registers an operation with a contract-address, a storage-key and storage-value arguments.
func (ctx *ArgumentContext) WriteValueOp(op uint16, address *common.Address, key *common.Hash, value *common.Hash) {
	addrClass, addrIdx := ctx.contracts.Classify(*address)
	keyClass, keyIdx := ctx.keys.Classify(*key)
	valueClass, valueIdx := ctx.values.Classify(*value)

	argOp := EncodeArgOp(op, addrClass, keyClass, valueClass)
	ctx.file.WriteUint16(argOp)

	// Write the address, key and value arguments
	ctx.writeClassifiedOp(addrClass, addrIdx, address)
	ctx.writeClassifiedOp(keyClass, keyIdx, key)
	ctx.writeClassifiedOp(valueClass, valueIdx, key)
}

func (ctx *ArgumentContext) Close() error {
	return ctx.file.Close()
}

func (ctx *ArgumentContext) writeClassifiedOp(class uint8, idx int, data Byter) {
	switch class {
	case ZeroValueID:
	case PreviousValueID:
	case RecentValueID:
		ctx.file.WriteUint8(uint8(idx))
	case NewValueID:
		ctx.file.WriteData(data.Bytes())
	default:
		panic("Unexpected argument classification")
	}
}

type Byter interface {
	Bytes() []byte
}
