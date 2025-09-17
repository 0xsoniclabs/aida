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

package recorder

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/ethereum/go-ethereum/common"

	"github.com/0xsoniclabs/aida/stochastic"
	"github.com/0xsoniclabs/aida/stochastic/operations"
	"github.com/0xsoniclabs/aida/stochastic/recorder/arguments"
	"github.com/0xsoniclabs/aida/stochastic/statistics/continuous"
)

// State counts states and transitions for a Markov-Process
// and classifies occurring arguments including snapshot deltas.
type State struct {
	// Frequency of argument-encoded operations
	argOpFreq [operations.NumArgOps]uint64

	// Transition frequencies between two subsequent argument-encoded operations
	transitFreq [operations.NumArgOps][operations.NumArgOps]uint64

	// Contract-address access statistics
	contracts arguments.Classifier[common.Address]

	// Storage-key access statistics
	keys arguments.Classifier[common.Hash]

	// Storage-value access statistics
	values arguments.Classifier[common.Hash]

	// Previous argument-encoded operation
	prevArgOp int

	// Snapshot deltas
	snapshotFreq map[int]uint64
}

// NewState creates a new state registry.
func NewState() State {
	return State{
		prevArgOp:    operations.NumArgOps,
		contracts:    arguments.NewClassifier[common.Address](),
		keys:         arguments.NewClassifier[common.Hash](),
		values:       arguments.NewClassifier[common.Hash](),
		snapshotFreq: map[int]uint64{},
	}
}

// RegisterOp registers an operation with no arguments
func (r *State) RegisterOp(op int) {
	r.updateFreq(
		op,
		stochastic.NoArgID,
		stochastic.NoArgID,
		stochastic.NoArgID,
	)
}

// RegisterSnapshot registers the delta between snapshot identifiers.
func (r *State) RegisterSnapshot(delta int) {
	r.snapshotFreq[delta]++
}

// RegisterAddressOp registers an operation with a contract-address argument
func (r *State) RegisterAddressOp(op int, address *common.Address) {
	r.updateFreq(
		op,
		r.contracts.Classify(*address),
		stochastic.NoArgID,
		stochastic.NoArgID,
	)
}

// RegisterAddressKeyOp registers an operation with a contract-address and a storage-key arguments.
func (r *State) RegisterKeyOp(op int, address *common.Address, key *common.Hash) {
	r.updateFreq(
		op,
		r.contracts.Classify(*address),
		r.keys.Classify(*key),
		stochastic.NoArgID,
	)
}

// RegisterAddressKeyOp registers an operation with a contract-address, a storage-key and storage-value arguments.
func (r *State) RegisterValueOp(op int, address *common.Address, key *common.Hash, value *common.Hash) {
	r.updateFreq(
		op,
		r.contracts.Classify(*address),
		r.keys.Classify(*key),
		r.values.Classify(*value),
	)
}

// updateFreq updates operation and transition frequency.
func (r *State) updateFreq(op int, addr int, key int, value int) {
	// encode argument classes to compute specialized operation using a Horner's scheme
	argOp, err := operations.EncodeArgOp(op, addr, key, value)
	if err != nil {
		panic(fmt.Sprintf("StateRegistry: cannot encode operation %v with arguments %v %v %v; Error %v", op, addr, key, value, err))
	}

	// increment operation's frequency depending on argument class
	r.argOpFreq[argOp]++

	// counting transition frequency (if not first state)
	if r.prevArgOp < operations.NumArgOps {
		r.transitFreq[r.prevArgOp][argOp] = r.transitFreq[r.prevArgOp][argOp] + 1
	}

	// remember current operation as previous operation
	r.prevArgOp = argOp
}

// StateJSON is the JSON struct for a recorded markov process
type StateJSON struct {
	FileId           string      `json:"FileId"`           // file identification
	Operations       []string    `json:"operations"`       // name of operations with argument classes
	StochasticMatrix [][]float64 `json:"stochasticMatrix"` // observed stochastic matrix

	// argument arguments statistics for contracts, keys, and values
	Contracts arguments.ClassifierJSON `json:"contractStats"`
	Keys      arguments.ClassifierJSON `json:"keyStats"`
	Values    arguments.ClassifierJSON `json:"valueSats"`

	// snapshot delta distribution
	SnapshotECDF [][2]float64 `json:"snapshotEcdf"`
}

// JSON produces the JSON struct for a recorded markov process
func (r *State) JSON() StateJSON {
	// generate labels for observable operations
	label := []string{}
	for argop := 0; argop < operations.NumArgOps; argop++ {
		if r.argOpFreq[argop] > 0 {
			op, addr, key, value, _ := operations.DecodeArgOp(argop)
			opc, _ := operations.EncodeOpcode(op, addr, key, value)
			label = append(label, opc)
		}
	}

	// compute stochastic matrix for observable operations with their arguments
	A := [][]float64{}
	for i := 0; i < operations.NumArgOps; i++ {
		if r.argOpFreq[i] > 0 {
			row := []float64{}
			// find row total of row (i.e. state i)
			total := uint64(0)
			for j := 0; j < operations.NumArgOps; j++ {
				total += r.transitFreq[i][j]
			}
			// normalize row
			for j := 0; j < operations.NumArgOps; j++ {
				if r.argOpFreq[j] > 0 {
					row = append(row, float64(r.transitFreq[i][j])/float64(total))
				}
			}
			A = append(A, row)
		}
	}

	return StateJSON{
		FileId:           "state",
		Operations:       label,
		StochasticMatrix: A,
		Contracts:        r.contracts.JSON(),
		Keys:             r.keys.JSON(),
		Values:           r.values.JSON(),
		SnapshotECDF:     continuous.ToCountECDF(&r.snapshotFreq),
	}
}

// Read state from a file in JSON format.
func Read(filename string) (*StateJSON, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed opening state file %v; %v", filename, err)
	}
	defer func(file *os.File) {
		err = errors.Join(err, file.Close())
	}(file)
	contents, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed reading state file; %v", err)
	}
	var stateJSON StateJSON
	err = json.Unmarshal(contents, &stateJSON)
	if err != nil {
		return nil, fmt.Errorf("cannot unmarshal state; %v", err)
	}
	if stateJSON.FileId != "state" {
		return nil, fmt.Errorf("file %v is not an state file", filename)
	}
	return &stateJSON, nil
}

// Write a state in JSON format.
func (r *State) Write(filename string) (err error) {
	f, fErr := os.Create(filename)
	if fErr != nil {
		return fmt.Errorf("cannot open JSON file; %v", fErr)
	}
	defer func(f *os.File) {
		err = errors.Join(err, f.Close())
	}(f)
	jOut, err := json.MarshalIndent(r.JSON(), "", "    ")
	if err != nil {
		return fmt.Errorf("failed to convert JSON file; %v", err)
	}
	_, err = fmt.Fprintln(f, string(jOut))
	if err != nil {
		return fmt.Errorf("failed to convert JSON file; %v", err)
	}
	return nil
}
