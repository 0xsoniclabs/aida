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
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/ethereum/go-ethereum/common"
	"github.com/paulmach/orb"
	"github.com/paulmach/orb/simplify"

	"github.com/0xsoniclabs/aida/stochastic/
)

// ArgumentContext keeps track of previously used arguments in StateDB operations
type ArgumentContext struct {
	// Contract-address queue
	contracts queue[common.Address]

	// Storage-key queue
	keys queue[common.Hash]

	// Storage-value queue
	values queue[common.Hash]

	// File handler
	// TBD: Use old file procedures (gzip etc.)
}

// NewArgumentContext creates a new event registry.
func NewArgumentContext(file FH) ArgumentContext {
	return ArgumentContext{
		contracts:    NewQueue[common.Address](),
		keys:         NewQueue[common.Hash](),
		values:       NewQueue[common.Hash](),
		file:         file,
	}
}

// WriteOp registers an operation with no simulation arguments
func (ctx *ArgumentContext) WriteOp(op int, data []byte) {
	if op < 0 || op >= NumOps {
		log.Fatalf("invalid operation ID")
	}
	addrClass := NoArgID
	keyClass := NoArgID
	valueClass := NoArgID 

	argOp := EncodeArgOp(op, addr, key, value)

	ctx.f.WriteUint16(argOp)
	ctx.f.WriteData(data)
}

// WriteAddressOp registers an operation with a contract-address argument
func (r *ArgumentContext) WriteAddressOp(op int, address *common.Address, data []byte) {
	if op < 0 || op >= NumOps {
		log.Fatalf("invalid stochastic operation ID")
	}

	addrClass, addrIdx := r.contracts.Classify(*address) // zero, previous, recent, address
	keyClass := NoArgID
	valueClass := NoArgID

	argOp := EncodeArgOp(op, addr, key, value)
	ctx.f.WriteUint16(argOp)

	switch(addrClass) {
	ZeroValueID:
	PreviousValueID:
	RecentValueID:
		f.WriteUint8(addrIdx)
	NewValueID:
		f.WriteData([]byte{*address})
 	default:
		panic("Unexpected argument classification")
	}

	ctx.f.WriteData(data)
}

// WriteAddressKeyOp registers an operation with a contract-address and a storage-key arguments.
func (r *ArgumentContext) WriteKeyOp(op int, address *common.Address, key *common.Hash, data []byte) {
	if op < 0 || op >= NumOps {
		log.Fatalf("invalid stochastic operation ID")
	}

	addrClass, addrIdx := r.contracts.Classify(*address)
	keyClass, keyIdx := r.keys.Classify(*key)
	valueClass := NoArgID

	argOp := EncodeArgOp(op, addr, key, value)
	ctx.f.WriteUint16(argOp)

	switch(addrClass) {
	ZeroValueID:
	PreviousValueID:
	RecentValueID:
		f.WriteUint8(addrIdx)
	NewValueID:
		f.WriteData([]byte{*address})
 	default:
		panic("Unexpected argument classification")
	}

	switch(keyClass) {
	ZeroValueID:
	PreviousValueID:
	RecentValueID:
		f.WriteUint8(addrIdx)
	NewValueID:
		f.WriteData([]byte{*key})
 	default:
		panic("Unexpected argument classification")
	}

	ctx.f.WriteData(data)
}

// WriteAddressKeyOp registers an operation with a contract-address, a storage-key and storage-value arguments.
func (ctx *ArgumentContext) WriteValueOp(op int, address *common.Address, key *common.Hash, value *common.Hash) {
	if op < 0 || op >= NumOps {
		log.Fatalf("invalid stochastic operation ID")
	}

	addrClass, addrIdx := r.contracts.Classify(*address)
	keyClass, keyIdx := r.keys.Classify(*key)
	valueClass, valueIdx := r.values.Classify(*value)

	argOp := EncodeArgOp(op, addr, key, value)
	ctx.f.WriteUint16(argOp)

	switch(addrClass) {
	ZeroValueID:
	PreviousValueID:
	RecentValueID:
		f.WriteUint8(addrIdx)
	NewValueID:
		f.WriteData([]byte{*address})
 	default:
		panic("Unexpected argument classification")
	}

	switch(keyClass) {
	ZeroValueID:
	PreviousValueID:
	RecentValueID:
		f.WriteUint8(addrIdx)
	NewValueID:
		f.WriteData([]byte{*key})
 	default:
		panic("Unexpected key classification")
	}

	switch(valueClass) {
	ZeroValueID:
	PreviousValueID:
	RecentValueID:
		f.WriteUint8(addrIdx)
	NewValueID:
		f.WriteData([]byte{*value})
 	default:
		panic("Unexpected value classification")
	}

	ctx.f.WriteData(data)
}
