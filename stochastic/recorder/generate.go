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
	"github.com/0xsoniclabs/aida/logger"
	"github.com/0xsoniclabs/aida/stochastic/operations"
	"github.com/0xsoniclabs/aida/stochastic/statistics/classifier"
	"github.com/0xsoniclabs/aida/utils"
)

// GenerateUniformRegistry produces a uniformly distributed simulation file.
func GenerateUniformRegistry(cfg *utils.Config, log logger.Logger) *EventRegistry {
	r := NewEventRegistry()

	// generate a uniform distribution for contracts, storage keys/values, and snapshots

	log.Infof("Number of contract addresses for priming: %v", cfg.ContractNumber)
	for i := int64(0); i < cfg.ContractNumber; i++ {
		for j := i - classifier.QueueLen - 1; j <= i; j++ {
			if j >= 0 {
				r.contracts.Place(operations.ToAddress(j))
			}
		}
	}

	log.Infof("Number of storage keys for priming: %v", cfg.KeysNumber)
	for i := int64(0); i < cfg.KeysNumber; i++ {
		for j := i - classifier.QueueLen - 1; j <= i; j++ {
			if j >= 0 {
				r.keys.Place(operations.ToHash(j))
			}
		}
	}

	log.Infof("Number of storage values for priming: %v", cfg.ValuesNumber)
	for i := int64(0); i < cfg.ValuesNumber; i++ {
		for j := i - classifier.QueueLen - 1; j <= i; j++ {
			if j >= 0 {
				r.values.Place(operations.ToHash(j))
			}
		}
	}

	log.Infof("Snapshot depth: %v", cfg.KeysNumber)
	for i := 0; i < cfg.SnapshotDepth; i++ {
		r.snapshotFreq[i] = 1
	}

	for i := 0; i < operations.NumArgOps; i++ {
		if operations.IsValidArgOp(i) {
			r.argOpFreq[i] = 1 // set frequency to greater than zero to emit operation
			opI, _, _, _ := operations.DecodeArgOp(i)
			switch opI {
			case operations.BeginSyncPeriodID:
				j := operations.EncodeArgOp(operations.BeginBlockID, classifier.NoArgID, classifier.NoArgID, classifier.NoArgID)
				r.transitFreq[i][j] = 1
			case operations.BeginBlockID:
				j := operations.EncodeArgOp(operations.BeginTransactionID, classifier.NoArgID, classifier.NoArgID, classifier.NoArgID)
				r.transitFreq[i][j] = 1
			case operations.EndTransactionID:
				j1 := operations.EncodeArgOp(operations.BeginTransactionID, classifier.NoArgID, classifier.NoArgID, classifier.NoArgID)
				j2 := operations.EncodeArgOp(operations.EndBlockID, classifier.NoArgID, classifier.NoArgID, classifier.NoArgID)
				r.transitFreq[i][j1] = cfg.BlockLength - 1
				r.transitFreq[i][j2] = 1
			case operations.EndBlockID:
				j1 := operations.EncodeArgOp(operations.BeginBlockID, classifier.NoArgID, classifier.NoArgID, classifier.NoArgID)
				j2 := operations.EncodeArgOp(operations.EndSyncPeriodID, classifier.NoArgID, classifier.NoArgID, classifier.NoArgID)
				r.transitFreq[i][j1] = cfg.SyncPeriodLength - 1
				r.transitFreq[i][j2] = 1
			case operations.EndSyncPeriodID:
				j := operations.EncodeArgOp(operations.BeginSyncPeriodID, classifier.NoArgID, classifier.NoArgID, classifier.NoArgID)
				r.transitFreq[i][j] = 1
			default:
				for j := 0; j < operations.NumArgOps; j++ {
					if operations.IsValidArgOp(j) {
						opJ, _, _, _ := operations.DecodeArgOp(j)
						if opJ != operations.BeginSyncPeriodID &&
							opJ != operations.BeginBlockID &&
							opJ != operations.BeginTransactionID &&
							opJ != operations.EndTransactionID &&
							opJ != operations.EndBlockID &&
							opJ != operations.EndSyncPeriodID {
							r.transitFreq[i][j] = cfg.TransactionLength - 1
						} else if opJ == operations.EndTransactionID {
							r.transitFreq[i][j] = 1
						}
					}
				}
			}
		}
	}
	return &r
}
