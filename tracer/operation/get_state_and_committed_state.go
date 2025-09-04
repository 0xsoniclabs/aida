// Copyright 2025 Sonic Labs
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

	"github.com/0xsoniclabs/aida/tracer/context"
)

// GetCommittedState data structure
type GetStateAndCommittedState struct {
	Contract common.Address
	Key      common.Hash
}

// GetId returns the get-commited-state-operation identifier.
func (op *GetStateAndCommittedState) GetId() byte {
	return GetStateAndCommittedStateID
}

// NewGetStateAndCommittedState creates a new get-commited-state operation.
func NewGetStateAndCommittedState(contract common.Address, key common.Hash) *GetStateAndCommittedState {
	return &GetStateAndCommittedState{Contract: contract, Key: key}
}

// ReadGetStateAndCommittedState reads a get-commited-state operation from file.
func ReadGetStateAndCommittedState(f io.Reader) (Operation, error) {
	data := new(GetStateAndCommittedState)
	err := binary.Read(f, binary.LittleEndian, data)
	return data, err
}

// Write the get-commited-state operation to file.
func (op *GetStateAndCommittedState) Write(f io.Writer) error {
	err := binary.Write(f, binary.LittleEndian, *op)
	return err
}

// Execute the get-committed-state operation.
func (op *GetStateAndCommittedState) Execute(db state.StateDB, ctx *context.Replay) time.Duration {
	contract := ctx.DecodeContract(op.Contract)
	storage := ctx.DecodeKey(op.Key)
	start := time.Now()
	db.GetCommittedState(contract, storage)
	return time.Since(start)
}

// Debug prints debug message for the get-committed-state operation.
func (op *GetStateAndCommittedState) Debug(ctx *context.Context) {
	fmt.Print(op.Contract, op.Key)
}
