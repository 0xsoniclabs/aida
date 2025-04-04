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
	"log"
	"math"
	"time"

	"github.com/0xsoniclabs/aida/state"

	"github.com/0xsoniclabs/aida/tracer/context"
)

// Snapshot data structure
type Snapshot struct {
	SnapshotID int32 // returned ID (for later mapping)
}

// GetId returns the snapshot operation identifier.
func (op *Snapshot) GetId() byte {
	return SnapshotID
}

// NewSnapshot creates a new snapshot operation.
func NewSnapshot(SnapshotID int32) *Snapshot {
	return &Snapshot{SnapshotID: SnapshotID}
}

// ReadSnapshot reads a snapshot operation from a file.
func ReadSnapshot(f io.Reader) (Operation, error) {
	data := new(Snapshot)
	err := binary.Read(f, binary.LittleEndian, data)
	return data, err
}

// Write the snapshot operation to file.
func (op *Snapshot) Write(f io.Writer) error {
	err := binary.Write(f, binary.LittleEndian, *op)
	return err
}

// Execute the snapshot operation.
func (op *Snapshot) Execute(db state.StateDB, ctx *context.Replay) time.Duration {
	start := time.Now()
	ID := db.Snapshot()
	elapsed := time.Since(start)
	if ID > math.MaxInt32 {
		log.Fatalf("Snapshot ID exceeds 32 bit")
	}
	ctx.AddSnapshot(op.SnapshotID, int32(ID))
	return elapsed
}

// Debug prints the details for the snapshot operation.
func (op *Snapshot) Debug(*context.Context) {
	fmt.Print(op.SnapshotID)
}
