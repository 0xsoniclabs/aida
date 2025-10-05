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
	"fmt"

	"github.com/0xsoniclabs/aida/logger"
	"github.com/0xsoniclabs/aida/stochastic"
	"github.com/0xsoniclabs/aida/stochastic/operations"
	"github.com/0xsoniclabs/aida/utils"
)

// GenerateUniformStats produces a uniformly distributed state file.
func GenerateUniformStats(cfg *utils.Config, log logger.Logger) (*Stats, error) {
	if err := validateUniformConfig(cfg); err != nil {
		return nil, err
	}
	s := NewStats()
	log.Infof("Number of contract addresses for priming: %v", cfg.ContractNumber)
	for i := int64(0); i < cfg.ContractNumber; i++ {
		for j := i - stochastic.QueueLen - 1; j <= i; j++ {
			if j >= 0 {
				addr, err := operations.ToAddress(j)
				if err != nil {
					return nil, err
				}
				s.contracts.Classify(addr)
			}
		}
	}
	log.Infof("Number of storage keys for priming: %v", cfg.KeysNumber)
	for i := int64(0); i < cfg.KeysNumber; i++ {
		for j := i - stochastic.QueueLen - 1; j <= i; j++ {
			if j >= 0 {
				key, err := operations.ToHash(j)
				if err != nil {
					return nil, err
				}
				s.keys.Classify(key)
			}
		}
	}
	log.Infof("Number of storage values for priming: %v", cfg.ValuesNumber)
	for i := int64(0); i < cfg.ValuesNumber; i++ {
		for j := i - stochastic.QueueLen - 1; j <= i; j++ {
			if j >= 0 {
				value, err := operations.ToHash(j)
				if err != nil {
					return nil, err
				}
				s.values.Classify(value)
			}
		}
	}
	log.Infof("Snapshot depth: %v", cfg.SnapshotDepth)
	for i := 0; i < cfg.SnapshotDepth; i++ {
		s.snapshotFreq[i] = 1
	}
	for i := 0; i < operations.NumArgOps; i++ {
		if !operations.IsValidArgOp(i) {
			continue
		}
		opI, _, _, _, err := operations.DecodeArgOp(i)
		if err != nil {
			return nil, err
		}
		s.argOpFreq[i] = 1 // set frequency to greater than zero to emit operation
		switch opI {
		case operations.BeginSyncPeriodID:
			j, err := operations.EncodeArgOp(operations.BeginBlockID, stochastic.NoArgID, stochastic.NoArgID, stochastic.NoArgID)
			if err != nil {
				return nil, err
			}
			s.transitFreq[i][j] = 1
		case operations.BeginBlockID:
			j, err := operations.EncodeArgOp(operations.BeginTransactionID, stochastic.NoArgID, stochastic.NoArgID, stochastic.NoArgID)
			if err != nil {
				return nil, err
			}
			s.transitFreq[i][j] = 1
		case operations.EndTransactionID:
			j1, err := operations.EncodeArgOp(operations.BeginTransactionID, stochastic.NoArgID, stochastic.NoArgID, stochastic.NoArgID)
			if err != nil {
				return nil, err
			}
			j2, err := operations.EncodeArgOp(operations.EndBlockID, stochastic.NoArgID, stochastic.NoArgID, stochastic.NoArgID)
			if err != nil {
				return nil, err
			}
			s.transitFreq[i][j1] = cfg.BlockLength - 1
			s.transitFreq[i][j2] = 1
		case operations.EndBlockID:
			j1, err := operations.EncodeArgOp(operations.BeginBlockID, stochastic.NoArgID, stochastic.NoArgID, stochastic.NoArgID)
			if err != nil {
				return nil, err
			}
			j2, err := operations.EncodeArgOp(operations.EndSyncPeriodID, stochastic.NoArgID, stochastic.NoArgID, stochastic.NoArgID)
			if err != nil {
				return nil, err
			}
			s.transitFreq[i][j1] = cfg.SyncPeriodLength - 1
			s.transitFreq[i][j2] = 1
		case operations.EndSyncPeriodID:
			j, err := operations.EncodeArgOp(operations.BeginSyncPeriodID, stochastic.NoArgID, stochastic.NoArgID, stochastic.NoArgID)
			if err != nil {
				return nil, err
			}
			s.transitFreq[i][j] = 1
		default:
			for j := 0; j < operations.NumArgOps; j++ {
				if !operations.IsValidArgOp(j) {
					continue
				}
				opJ, _, _, _, err := operations.DecodeArgOp(j)
				if err != nil {
					return nil, err
				}
				if opJ != operations.BeginSyncPeriodID &&
					opJ != operations.BeginBlockID &&
					opJ != operations.BeginTransactionID &&
					opJ != operations.EndTransactionID &&
					opJ != operations.EndBlockID &&
					opJ != operations.EndSyncPeriodID {
					s.transitFreq[i][j] = cfg.TransactionLength - 1
				} else if opJ == operations.EndTransactionID {
					s.transitFreq[i][j] = 1
				}
			}
		}
	}
	return &s, nil
}

func validateUniformConfig(cfg *utils.Config) error {
	if cfg.BlockLength == 0 {
		return fmt.Errorf("block-length must be greater than zero")
	}
	if cfg.SyncPeriodLength == 0 {
		return fmt.Errorf("sync-period must be greater than zero")
	}
	if cfg.TransactionLength == 0 {
		return fmt.Errorf("transaction-length must be greater than zero")
	}
	if cfg.ContractNumber <= 0 {
		return fmt.Errorf("num-contracts must be greater than zero")
	}
	if cfg.KeysNumber <= 0 {
		return fmt.Errorf("num-keys must be greater than zero")
	}
	if cfg.ValuesNumber <= 0 {
		return fmt.Errorf("num-values must be greater than zero")
	}
	if cfg.SnapshotDepth <= 0 {
		return fmt.Errorf("snapshot-depth must be greater than zero")
	}
	return nil
}
