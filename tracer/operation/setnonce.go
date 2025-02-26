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

package operation

import (
	"encoding/binary"
	"fmt"
	"io"
	"time"

	"github.com/0xsoniclabs/aida/state"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/tracing"

	"github.com/0xsoniclabs/aida/tracer/context"
)

// SetNonce data structure
type SetNonce struct {
	Contract common.Address
	Nonce    uint64 // nonce
	Reason   tracing.NonceChangeReason
}

// GetId returns the set-nonce operation identifier.
func (op *SetNonce) GetId() byte {
	return SetNonceID
}

// NewSetNonce creates a new set-nonce operation.
func NewSetNonce(contract common.Address, nonce uint64, reason tracing.NonceChangeReason) *SetNonce {
	return &SetNonce{Contract: contract, Nonce: nonce, Reason: reason}
}

// ReadSetNonce reads a set-nonce operation from a file.
func ReadSetNonce(f io.Reader) (Operation, error) {
	data := new(SetNonce)
	err := binary.Read(f, binary.LittleEndian, data)
	return data, err
}

// Write the set-nonce operation to a file.
func (op *SetNonce) Write(f io.Writer) error {
	err := binary.Write(f, binary.LittleEndian, *op)
	return err
}

// Execute the set-nonce operation.
func (op *SetNonce) Execute(db state.StateDB, ctx *context.Replay) time.Duration {
	contract := ctx.DecodeContract(op.Contract)
	start := time.Now()
	db.SetNonce(contract, op.Nonce, op.Reason)
	return time.Since(start)
}

// Debug prints a debug message for the set-nonce operation.
func (op *SetNonce) Debug(ctx *context.Context) {
	fmt.Print(op.Contract, op.Nonce, op.Reason)
}
