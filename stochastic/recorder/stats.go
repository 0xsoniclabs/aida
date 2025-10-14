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

package recorder

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"os"
	"sort"

	"github.com/ethereum/go-ethereum/common"

	"github.com/0xsoniclabs/aida/stochastic"
	"github.com/0xsoniclabs/aida/stochastic/operations"
	"github.com/0xsoniclabs/aida/stochastic/recorder/arguments"
	"github.com/0xsoniclabs/aida/stochastic/statistics/continuous"
)

type scalarStats struct {
	freq map[int64]uint64
}

func newScalarStats() scalarStats {
	return scalarStats{freq: map[int64]uint64{}}
}

func (s *scalarStats) record(value int64) {
	if value < 0 {
		return
	}
	s.freq[value]++
}

type ScalarStatsJSON struct {
	Max  int64        `json:"max"`
	ECDF [][2]float64 `json:"ecdf"`
}

func (s scalarStats) json() ScalarStatsJSON {
	if len(s.freq) == 0 {
		return ScalarStatsJSON{
			Max:  0,
			ECDF: [][2]float64{{0.0, 0.0}, {1.0, 1.0}},
		}
	}
	values := make([]int64, 0, len(s.freq))
	var total uint64
	var maxVal int64
	for value, freq := range s.freq {
		if freq == 0 {
			continue
		}
		values = append(values, value)
		total += freq
		if value > maxVal {
			maxVal = value
		}
	}
	sort.Slice(values, func(i, j int) bool { return values[i] < values[j] })
	domain := float64(maxVal + 1)
	pdf := make([][2]float64, 0, len(values))
	for _, value := range values {
		prob := float64(s.freq[value]) / float64(total)
		x := (float64(value) + 0.5) / domain
		pdf = append(pdf, [2]float64{x, prob})
	}
	ecdf := continuous.PDFtoCDF(pdf)
	return ScalarStatsJSON{
		Max:  maxVal,
		ECDF: ecdf,
	}
}

// Stats counts states and transitions for a Markov-Process
// and classifies occurring arguments including snapshot deltas.
type Stats struct {
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

	balance scalarStats
	nonce   scalarStats
	code    scalarStats
}

// NewStats creates a new stats object for recording.
func NewStats() Stats {
	return Stats{
		prevArgOp:    operations.NumArgOps,
		contracts:    arguments.NewClassifier[common.Address](),
		keys:         arguments.NewClassifier[common.Hash](),
		values:       arguments.NewClassifier[common.Hash](),
		snapshotFreq: map[int]uint64{},
		balance:      newScalarStats(),
		nonce:        newScalarStats(),
		code:         newScalarStats(),
	}
}

// CountOp counts an operation with no arguments
func (r *Stats) CountOp(op int) {
	if err := r.updateFreq(
		op,
		stochastic.NoArgID,
		stochastic.NoArgID,
		stochastic.NoArgID,
	); err != nil {
		panic(fmt.Errorf("CountOp: %w", err))
	}
}

// CountSnapshot counts the delta between snapshot identifiers
// and the operation RevertToSnapshot.
func (r *Stats) CountSnapshot(delta int) {
	r.CountOp(operations.RevertToSnapshotID)
	r.snapshotFreq[delta]++
}

// CountAddressOp counts an operation with a contract-address argument
func (r *Stats) CountAddressOp(op int, address *common.Address) {
	if err := r.updateFreq(
		op,
		r.contracts.Classify(*address),
		stochastic.NoArgID,
		stochastic.NoArgID,
	); err != nil {
		panic(fmt.Errorf("CountAddressOp: %w", err))
	}
}

// CountKeyOp counts an operation with a contract-address and a storage-key arguments.
func (r *Stats) CountKeyOp(op int, address *common.Address, key *common.Hash) {
	if err := r.updateFreq(
		op,
		r.contracts.Classify(*address),
		r.keys.Classify(*key),
		stochastic.NoArgID,
	); err != nil {
		panic(fmt.Errorf("CountKeyOp: %w", err))
	}
}

// CountValueOp counts an operation with a contract-address, a storage-key and storage-value arguments.
func (r *Stats) CountValueOp(op int, address *common.Address, key *common.Hash, value *common.Hash) {
	if err := r.updateFreq(
		op,
		r.contracts.Classify(*address),
		r.keys.Classify(*key),
		r.values.Classify(*value),
	); err != nil {
		panic(fmt.Errorf("CountValueOp: %w", err))
	}
}

// RecordBalance tracks the magnitude used in balance updates.
func (r *Stats) RecordBalance(value int64) {
	if value < 0 {
		value = 0
	}
	r.balance.record(value)
}

// RecordNonce tracks nonce assignments.
func (r *Stats) RecordNonce(value uint64) {
	if value > math.MaxInt64 {
		r.nonce.record(math.MaxInt64)
		return
	}
	r.nonce.record(int64(value))
}

// RecordCodeSize tracks code sizes used when setting bytecode.
func (r *Stats) RecordCodeSize(size int) {
	if size < 0 {
		return
	}
	r.code.record(int64(size))
}

// updateFreq updates operation and transition frequency.
func (r *Stats) updateFreq(op int, addr int, key int, value int) error {
	// encode argument classes to compute specialized operation using a Horner's scheme
	argOp, err := operations.EncodeArgOp(op, addr, key, value)
	if err != nil {
		return fmt.Errorf("updateFreq: cannot encode operation %v with arguments %v %v %v: %w", op, addr, key, value, err)
	}

	// increment operation's frequency depending on argument class
	r.argOpFreq[argOp]++

	// counting transition frequency (if not first state)
	if r.prevArgOp < operations.NumArgOps {
		r.transitFreq[r.prevArgOp][argOp] = r.transitFreq[r.prevArgOp][argOp] + 1
	}

	// remember current operation as previous operation
	r.prevArgOp = argOp
	return nil
}

// StatsJSON is the JSON struct for a recorded markov process
type StatsJSON struct {
	FileId           string      `json:"FileId"`           // file identification
	Operations       []string    `json:"operations"`       // name of operations with argument classes
	StochasticMatrix [][]float64 `json:"stochasticMatrix"` // observed stochastic matrix

	// argument arguments statistics for contracts, keys, and values
	Contracts arguments.ClassifierJSON `json:"contractStats"`
	Keys      arguments.ClassifierJSON `json:"keyStats"`
	Values    arguments.ClassifierJSON `json:"valueSats"`

	// snapshot delta distribution
	SnapshotECDF [][2]float64 `json:"snapshotEcdf"`

	// scalar argument statistics
	Balance  ScalarStatsJSON `json:"balanceStats"`
	Nonce    ScalarStatsJSON `json:"nonceStats"`
	CodeSize ScalarStatsJSON `json:"codeSizeStats"`
}

const statsFileID = "stats"

// MarshalJSON ensures the FileId is populated before serialising.
func (s StatsJSON) MarshalJSON() ([]byte, error) {
	if s.FileId == "" {
		s.FileId = statsFileID
	}
	type alias StatsJSON
	return json.Marshal(alias(s))
}

// UnmarshalJSON validates the FileId while deserialising.
func (s *StatsJSON) UnmarshalJSON(data []byte) error {
	type alias StatsJSON
	var tmp alias
	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}
	if tmp.FileId == "" {
		return fmt.Errorf("StatsJSON: missing FileId")
	}
	if tmp.FileId != statsFileID {
		return fmt.Errorf("StatsJSON: unexpected FileId %q", tmp.FileId)
	}
	*s = StatsJSON(tmp)
	return nil
}

// JSON produces the JSON struct for a recorded markov process
func (r *Stats) JSON() StatsJSON {
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
					probability := float64(0.0)
					if total > 0 {
						probability = float64(r.transitFreq[i][j]) / float64(total)
					}
					row = append(row, probability)
				}
			}
			A = append(A, row)
		}
	}

	// compute PDF for snapshots distribution
	total := uint64(0)
	maxArg := 0
	for arg, freq := range r.snapshotFreq {
		total += freq
		if maxArg < arg {
			maxArg = arg
		}
	}
	pdf := [][2]float64{}
	for arg := range maxArg {
		x := (float64(arg) + 0.5) / float64(maxArg)
		f := float64(r.snapshotFreq[arg]) / float64(total)
		pdf = append(pdf, [2]float64{x, f})
	}
	ecdf := continuous.PDFtoCDF(pdf)
	return StatsJSON{
		FileId:           "stats",
		Operations:       label,
		StochasticMatrix: A,
		Contracts:        r.contracts.JSON(),
		Keys:             r.keys.JSON(),
		Values:           r.values.JSON(),
		SnapshotECDF:     ecdf,
		Balance:          r.balance.json(),
		Nonce:            r.nonce.json(),
		CodeSize:         r.code.json(),
	}
}

// Read stats from a file in JSON format.
func Read(filename string) (*StatsJSON, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed opening stats file %v; %v", filename, err)
	}
	defer func(file *os.File) {
		err = errors.Join(err, file.Close())
	}(file)
	contents, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed reading stats file; %v", err)
	}
	var statsJSON StatsJSON
	err = json.Unmarshal(contents, &statsJSON)
	if err != nil {
		return nil, fmt.Errorf("cannot unmarshal stats; %v", err)
	}
	if statsJSON.FileId != "stats" {
		return nil, fmt.Errorf("file %v is not an stats file", filename)
	}
	return &statsJSON, nil
}

// Write a stats in JSON format.
func (r *Stats) Write(filename string) (err error) {
	f, fErr := os.Create(filename)
	if fErr != nil {
		return fmt.Errorf("cannot open for writing JSON file; %v", fErr)
	}
	defer func(f *os.File) {
		err = errors.Join(err, f.Close())
	}(f)
	jOut, err := json.MarshalIndent(r.JSON(), "", "    ")
	if err != nil {
		return fmt.Errorf("failed to convert JSON; %v", err)
	}
	_, err = fmt.Fprintln(f, string(jOut))
	if err != nil {
		return fmt.Errorf("failed to write file; %v", err)
	}
	return nil
}
