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

package visualizer

import (
	"sort"

	"github.com/0xsoniclabs/aida/stochastic/operations"
	"github.com/0xsoniclabs/aida/stochastic/recorder"
	"github.com/0xsoniclabs/aida/stochastic/recorder/arguments"
	"github.com/0xsoniclabs/aida/stochastic/statistics/markov"
)

// StatsData contains the statistical visualisation data for the markov chain
type StatsData struct {
	Contracts ArgumentData // contract-address view model
	Keys      ArgumentData // storage-key view model
	Values    ArgumentData // storage-value view model
	Snapshot  SnapshotData // snapshot view model

	Stationary          []OpData                                      // stationary distribution model
	TxOperation         []OpData                                      // average number of operations per Tx
	TxPerBlock          float64                                       // average number of transactions per block
	BlocksPerSyncPeriod float64                                       // average number of blocks per sync-period
	OperationLabel      []string                                      // operation labels for stochastic matrix
	StochasticMatrix    [][]float64                                   // stochastic Matrix
	SimplifiedMatrix    [operations.NumOps][operations.NumOps]float64 // simplified stochastic matrix
}

// ArgumentData contains the statistical data for access statistics that is used for visualization.
type ArgumentData struct {
	A_CDF [][2]float64 // empirical cumulative distribution function of counting stats
	Q_PMF []float64    // queuing distribution function
}

// SnapshotData contains the statistical data for snapshot deltas used for visualization.
type SnapshotData struct {
	ECdf [][2]float64 // empirical cumulative distribution function
}

// OpData stores a single operation and its probability (for stead-state)
type OpData struct {
	label string  // operation's label
	value float64 // operation's value (either probability or frequency)
}

// stats is the singleton for the viewing model.
var stats StatsData

// GetData returns the pointer to the singleton.
func GetData() *StatsData {
	return &stats
}

// PopulateData populates the model from the JSON struct
func (e *StatsData) PopulateData(d *recorder.StatsJSON) {

	// populate access stats for contract addresses
	e.Contracts.PopulateAccess(&d.Contracts)

	// populate access stats for storage keys
	e.Keys.PopulateAccess(&d.Keys)

	// populate access stats for storage values
	e.Values.PopulateAccess(&d.Values)

	// populate access stats for storage values
	e.Snapshot.PopulateSnapshotStats(d)

	// Sort entries of the stationary distribution and populate
	n := len(d.Operations)
	mc, mc_err := markov.New(d.StochasticMatrix, d.Operations)
	if mc_err != nil {
		panic("PopulateData: Expected a new markov chain. Error: %v")
	}
	stationary, _ := mc.Stationary()
	data := []OpData{}
	for i := 0; i < n; i++ {
		data = append(data, OpData{
			label: d.Operations[i],
			value: stationary[i],
		})
	}
	sort.Slice(data, func(i, j int) bool {
		return data[i].value < data[j].value
	})
	e.Stationary = data

	// compute average number of operations per transaction

	// find the BeginTransaction probability in the stationary distribution
	txProb := 0.0
	blockProb := 0.0
	syncPeriodProb := 0.0
	for i := 0; i < n; i++ {
		sop, _, _, _, _ := operations.DecodeOpcode(d.Operations[i])
		if sop == operations.BeginTransactionID {
			txProb = stationary[i]
		}
		if sop == operations.BeginBlockID {
			blockProb = stationary[i]
		}
		if sop == operations.BeginSyncPeriodID {
			syncPeriodProb = stationary[i]
		}
	}
	if blockProb > 0.0 {
		e.TxPerBlock = txProb / blockProb
	}
	if syncPeriodProb > 0.0 {
		e.BlocksPerSyncPeriod = blockProb / syncPeriodProb
	}

	txData := []OpData{}
	if txProb > 0.0 {
		for op := 0; op < operations.NumOps; op++ {
			// exclude scoping operations
			if op != operations.BeginBlockID && op != operations.EndBlockID && op != operations.BeginSyncPeriodID &&
				op != operations.EndSyncPeriodID && op != operations.BeginTransactionID && op != operations.EndTransactionID {
				// sum all versions of an operation and normalize the value with the transaction's probability
				sum := 0.0
				for i := 0; i < n; i++ {
					if sop, _, _, _, _ := operations.DecodeOpcode(d.Operations[i]); sop == op {
						sum += stationary[i]
					}
				}
				txData = append(txData, OpData{
					label: operations.OpMnemo(op),
					value: sum / txProb})
			}
		}
	}
	// sort expected operation frequencies
	sort.Slice(txData, func(i, j int) bool {
		return txData[i].value > txData[j].value
	})
	e.TxOperation = txData

	// Populate stochastic matrix
	e.OperationLabel = make([]string, len(d.Operations))
	copy(e.OperationLabel, d.Operations)
	e.StochasticMatrix = make([][]float64, len(d.StochasticMatrix))
	for i := range d.StochasticMatrix {
		e.StochasticMatrix[i] = make([]float64, len(d.StochasticMatrix[i]))
		copy(e.StochasticMatrix[i], d.StochasticMatrix[i])
	}

	// reduce stochastic matrix to a simplified matrix
	for i := 0; i < n; i++ {
		iop, _, _, _, _ := operations.DecodeOpcode(d.Operations[i])
		for j := 0; j < n; j++ {
			jop, _, _, _, _ := operations.DecodeOpcode(d.Operations[j])
			e.SimplifiedMatrix[iop][jop] += d.StochasticMatrix[i][j]
		}
	}

	// normalize row data after reduction
	for i := 0; i < operations.NumOps; i++ {
		sum := 0.0
		for j := 0; j < operations.NumOps; j++ {
			sum += e.SimplifiedMatrix[i][j]
		}
		for j := 0; j < operations.NumOps; j++ {
			e.SimplifiedMatrix[i][j] /= sum
		}
	}
}

// PopulateAccess populates access stats model
func (a *ArgumentData) PopulateAccess(d *arguments.ClassifierJSON) {
	a.A_CDF = make([][2]float64, len(d.Counting.ECDF))
	copy(a.A_CDF, d.Counting.ECDF)
	a.Q_PMF = make([]float64, len(d.Queuing.Distribution))
	copy(a.Q_PMF, d.Queuing.Distribution)
}

// PopulateSnapStats populates snapshot stats model
func (s *SnapshotData) PopulateSnapshotStats(d *recorder.StatsJSON) {
	s.ECdf = make([][2]float64, len(d.SnapshotECDF))
	copy(s.ECdf, d.SnapshotECDF)
}
