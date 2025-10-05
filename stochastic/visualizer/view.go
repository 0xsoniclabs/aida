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
	"fmt"
	"sort"
	"sync"

	"github.com/0xsoniclabs/aida/stochastic/operations"
	"github.com/0xsoniclabs/aida/stochastic/recorder"
	"github.com/0xsoniclabs/aida/stochastic/statistics/markov"
)

type opDatum struct {
	label string
	value float64
}

type viewState struct {
	stats               *recorder.StatsJSON
	stationary          []opDatum
	txOperation         []opDatum
	txPerBlock          float64
	blocksPerSyncPeriod float64
	simplifiedMatrix    [operations.NumOps][operations.NumOps]float64
}

var (
	currentMu    sync.RWMutex
	currentState *viewState
)

func setViewState(stats *recorder.StatsJSON) error {
	if stats == nil {
		return fmt.Errorf("visualizer: stats are nil")
	}
	derived, err := buildViewState(stats)
	if err != nil {
		return err
	}
	currentMu.Lock()
	currentState = derived
	currentMu.Unlock()
	return nil
}

func buildViewState(stats *recorder.StatsJSON) (*viewState, error) {
	mc, err := markov.New(stats.StochasticMatrix, stats.Operations)
	if err != nil {
		return nil, fmt.Errorf("visualizer: create markov chain: %w", err)
	}
	stationary, err := mc.Stationary()
	if err != nil {
		return nil, fmt.Errorf("visualizer: stationary distribution: %w", err)
	}

	stationaryData := make([]opDatum, len(stationary))
	for i := range stationary {
		stationaryData[i] = opDatum{
			label: stats.Operations[i],
			value: stationary[i],
		}
	}
	sort.Slice(stationaryData, func(i, j int) bool {
		return stationaryData[i].value < stationaryData[j].value
	})

	txOps, txPerBlock, blocksPerSyncPeriod, err := computeTxOperation(stats, stationary)
	if err != nil {
		return nil, err
	}
	simplified, err := computeSimplifiedMatrix(stats)
	if err != nil {
		return nil, err
	}

	return &viewState{
		stats:               stats,
		stationary:          stationaryData,
		txOperation:         txOps,
		txPerBlock:          txPerBlock,
		blocksPerSyncPeriod: blocksPerSyncPeriod,
		simplifiedMatrix:    simplified,
	}, nil
}

func computeTxOperation(stats *recorder.StatsJSON, stationary []float64) ([]opDatum, float64, float64, error) {
	n := len(stats.Operations)
	txProb := 0.0
	blockProb := 0.0
	syncPeriodProb := 0.0
	for i := 0; i < n; i++ {
		sop, _, _, _, err := operations.DecodeOpcode(stats.Operations[i])
		if err != nil {
			return nil, 0, 0, fmt.Errorf("visualizer: decode opcode %q: %w", stats.Operations[i], err)
		}
		switch sop {
		case operations.BeginTransactionID:
			txProb = stationary[i]
		case operations.BeginBlockID:
			blockProb = stationary[i]
		case operations.BeginSyncPeriodID:
			syncPeriodProb = stationary[i]
		}
	}

	txPerBlock := 0.0
	if blockProb > 0 {
		txPerBlock = txProb / blockProb
	}
	blocksPerSync := 0.0
	if syncPeriodProb > 0 {
		blocksPerSync = blockProb / syncPeriodProb
	}

	txData := []opDatum{}
	if txProb > 0 {
		for op := 0; op < operations.NumOps; op++ {
			if op == operations.BeginBlockID || op == operations.EndBlockID || op == operations.BeginSyncPeriodID ||
				op == operations.EndSyncPeriodID || op == operations.BeginTransactionID || op == operations.EndTransactionID {
				continue
			}
			sum := 0.0
			for i := 0; i < n; i++ {
				sop, _, _, _, err := operations.DecodeOpcode(stats.Operations[i])
				if err != nil {
					return nil, 0, 0, fmt.Errorf("visualizer: decode opcode %q: %w", stats.Operations[i], err)
				}
				if sop == op {
					sum += stationary[i]
				}
			}
			txData = append(txData, opDatum{
				label: operations.OpMnemo(op),
				value: sum / txProb,
			})
		}
	}
	sort.Slice(txData, func(i, j int) bool {
		return txData[i].value > txData[j].value
	})

	return txData, txPerBlock, blocksPerSync, nil
}

func computeSimplifiedMatrix(stats *recorder.StatsJSON) ([operations.NumOps][operations.NumOps]float64, error) {
	var simplified [operations.NumOps][operations.NumOps]float64
	n := len(stats.Operations)
	for i := 0; i < n; i++ {
		iop, _, _, _, err := operations.DecodeOpcode(stats.Operations[i])
		if err != nil {
			return simplified, fmt.Errorf("visualizer: decode opcode %q: %w", stats.Operations[i], err)
		}
		for j := 0; j < n; j++ {
			jop, _, _, _, err := operations.DecodeOpcode(stats.Operations[j])
			if err != nil {
				return simplified, fmt.Errorf("visualizer: decode opcode %q: %w", stats.Operations[j], err)
			}
			simplified[iop][jop] += stats.StochasticMatrix[i][j]
		}
	}

	for i := 0; i < operations.NumOps; i++ {
		sum := 0.0
		for j := 0; j < operations.NumOps; j++ {
			sum += simplified[i][j]
		}
		if sum == 0 {
			continue
		}
		for j := 0; j < operations.NumOps; j++ {
			simplified[i][j] /= sum
		}
	}
	return simplified, nil
}

func currentView() (*viewState, error) {
	currentMu.RLock()
	defer currentMu.RUnlock()
	if currentState == nil {
		return nil, fmt.Errorf("visualizer: statistics not initialised")
	}
	return currentState, nil
}
