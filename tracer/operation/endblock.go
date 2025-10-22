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
	"io"
	"time"

	"github.com/0xsoniclabs/aida/state"

	"github.com/0xsoniclabs/aida/tracer/context"
)

// Endblock data structure
type EndBlock struct {
}

// GetId returns the end-block operation identifier.
func (op *EndBlock) GetId() byte {
	return EndBlockID
}

// NewEndBlock creates a new end-block operation.
func NewEndBlock() *EndBlock {
	return &EndBlock{}
}

// ReadEndBlock reads an end-block operation from file.
func ReadEndBlock(f io.Reader) (Operation, error) {
	return new(EndBlock), nil
}

// Write the end-block operation to file.
func (op *EndBlock) Write(f io.Writer) error {
	return binary.Write(f, binary.LittleEndian, *op)
}

// Execute the end-block operation.
func (op *EndBlock) Execute(db state.StateDB, ctx *context.Replay) (time.Duration, error) {
	start := time.Now()
	err := db.EndBlock()
	if err != nil {
		return 0, err
	}
	return time.Since(start), nil
}

// Debug prints a debug message for the end-block operation.
func (op *EndBlock) Debug(ctx *context.Context) {
}
