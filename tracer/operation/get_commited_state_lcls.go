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
	"github.com/0xsoniclabs/aida/tracer/context"
)

// GetStateAndCommittedStateLcls data structure
type GetStateAndCommittedStateLcls struct {}

// GetId returns the get-commited-state-operation identifier.
func (op *GetStateAndCommittedStateLcls) GetId() byte {
	return GetStateAndCommittedStateLclsID
}

// NewGetStateAndCommittedStateLcls creates a new get-commited-state operation.
func NewGetStateAndCommittedStateLcls() *GetStateAndCommittedStateLcls {
	return new(GetStateAndCommittedStateLcls)
}

// ReadGetStateAndCommittedStateLcls reads a get-commited-state operation from file.
func ReadGetStateAndCommittedStateLcls(f io.Reader) (Operation, error) {
	data := new(GetStateAndCommittedStateLcls)
	err := binary.Read(f, binary.LittleEndian, data)
	return data, err
}

// Write the get-commited-state operation to file.
func (op *GetStateAndCommittedStateLcls) Write(f io.Writer) error {
	err := binary.Write(f, binary.LittleEndian, *op)
	return err
}

// Execute the get-committed-state operation.
func (op *GetStateAndCommittedStateLcls) Execute(db state.StateDB, ctx *context.Replay) time.Duration {
	contract := ctx.PrevContract()
	storage := ctx.DecodeKeyCache(0)
	start := time.Now()
	db.GetStateAndCommittedState(contract, storage)
	return time.Since(start)
}

// Debug prints debug message for the get-committed-state operation.
func (op *GetStateAndCommittedStateLcls) Debug(ctx *context.Context) {
	contract := ctx.PrevContract()
	storage := ctx.ReadKeyCache(0)
	fmt.Print(contract, storage)
}
