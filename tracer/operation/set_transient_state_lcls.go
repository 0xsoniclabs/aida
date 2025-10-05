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
	"github.com/ethereum/go-ethereum/common"
)

// SetTransientStateLcls data structure
type SetTransientStateLcls struct {
	Value common.Hash // encoded storage value
}

// GetId returns the set-state-lcls identifier.
func (op *SetTransientStateLcls) GetId() byte {
	return SetTransientStateLclsID
}

// SetTransientStateLcls creates a new set-state-lcls operation.
func NewSetTransientStateLcls(value common.Hash) *SetTransientStateLcls {
	return &SetTransientStateLcls{Value: value}
}

// ReadSetTransientStateLcls reads a set-state-lcls operation from file.
func ReadSetTransientStateLcls(f io.Reader) (Operation, error) {
	data := new(SetTransientStateLcls)
	err := binary.Read(f, binary.LittleEndian, data)
	return data, err
}

// Write the set-state-lcls operation to file.
func (op *SetTransientStateLcls) Write(f io.Writer) error {
	err := binary.Write(f, binary.LittleEndian, *op)
	return err
}

// Execute the set-state-lcls operation.
func (op *SetTransientStateLcls) Execute(db state.StateDB, ctx *context.Replay) time.Duration {
	contract := ctx.PrevContract()
	storage := ctx.DecodeKeyCache(0)
	start := time.Now()
	db.SetTransientState(contract, storage, op.Value)
	return time.Since(start)
}

// Debug prints a debug message for the set-state-lcls operation.
func (op *SetTransientStateLcls) Debug(ctx *context.Context) {
	contract := ctx.PrevContract()
	storage := ctx.ReadKeyCache(0)
	fmt.Print(contract, storage, op.Value)
}
