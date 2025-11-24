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

package ethtest

import (
	"fmt"
	"math/big"

	"github.com/0xsoniclabs/aida/logger"
	"github.com/0xsoniclabs/aida/txcontext"
	"github.com/0xsoniclabs/aida/utils"
	"github.com/ethereum/go-ethereum/params"
	"github.com/ethereum/go-ethereum/tests"
	"golang.org/x/exp/maps"
)

type Transaction struct {
	Fork string
	Ctx  txcontext.TxContext
}

var usableForks = map[string]struct{}{
	"Osaka":        {},
	"Prague":       {},
	"Cancun":       {},
	"Shanghai":     {},
	"Paris":        {},
	"Bellatrix":    {},
	"GrayGlacier":  {},
	"ArrowGlacier": {},
	"Altair":       {},
	"London":       {},
	"Berlin":       {},
	"Istanbul":     {},
	"MuirGlacier":  {},
	"TestNetwork":  {},
}

// NewTestCaseSplitter opens all JSON tests within path
func NewTestCaseSplitter(cfg *utils.Config) (*TestCaseSplitter, error) {
	tests, err := getTestsWithinPath(cfg, utils.StateTests)
	if err != nil {
		return nil, err
	}
	log := logger.NewLogger(cfg.LogLevel, "eth-test-decoder")

	return &TestCaseSplitter{
		enabledForks: sortForks(log, cfg.Fork),
		log:          log,
		jsons:        tests,
		chainConfigs: make(map[string]*params.ChainConfig),
	}, nil
}

func sortForks(log logger.Logger, cfgFork string) (forks []string) {
	cfgFork = utils.ToTitleCase(cfgFork)
	if cfgFork == "All" {
		forks = maps.Keys(usableForks)
	} else {
		if _, ok := usableForks[cfgFork]; !ok {
			log.Warningf("Unknown name fork name %v, removing", cfgFork)
		} else {
			forks = append(forks, cfgFork)
		}
	}
	return forks
}

type TestCaseSplitter struct {
	enabledForks []string  // Which forks are enabled by user (default is all)
	jsons        []*stJSON // Decoded json fil
	log          logger.Logger
	chainConfigs map[string]*params.ChainConfig
}

// SplitStateTests iterates unmarshalled Geth-State test-files and divides them by 1) fork and
// 2) tests cases. Each file contains 1..N enabledForks, one block environment (marked as 'env') and one
// input alloc (marked as 'env'). Each fork within a file contains 1..N tests (marked as 'post').
func (s *TestCaseSplitter) SplitStateTests() (dividedTests []Transaction, err error) {
	var overall int

	// Iterate all JSONs
	for _, stJson := range s.jsons {
		baseFee := stJson.Env.BaseFee
		if baseFee == nil {
			// ethereum uses `0x10` for genesis baseFee. Therefore, it defaults to
			// parent - 2 : 0xa as the basefee for 'this' context.
			baseFee = &BigInt{*big.NewInt(0x0a)}
		}

		// Iterate all usable forks within one JSON file
		for _, fork := range s.enabledForks {
			posts, ok := stJson.Post[fork]
			if !ok {
				continue
			}
			chainCfg, err := s.getChainConfig(fork)
			if err != nil {
				return nil, err
			}
			// Iterate all tests within one fork
			for postNumber, post := range posts {
				msg, err := stJson.Tx.toMessage(post, baseFee)
				if err != nil {
					s.log.Warningf("Path: %v, fork: %v, test postNumber: %v\n"+
						"cannot decode tx to message: %v", stJson.path, fork, postNumber, err)
					continue
				}

				if fork == "Paris" {
					fork = "Merge"
				}
				txCtx := newStateTestTxContext(stJson, msg, post, chainCfg, stJson.testLabel, fork, postNumber)
				dividedTests = append(dividedTests, Transaction{
					fork,
					txCtx,
				})
				overall++
			}
		}
	}

	s.log.Noticef("Found %v runnable state tests...", overall)

	return dividedTests, err
}

func (s *TestCaseSplitter) getChainConfig(fork string) (*params.ChainConfig, error) {
	if cfg, ok := s.chainConfigs[fork]; ok {
		return cfg, nil
	}
	cfg, _, err := tests.GetChainConfig(fork)
	if err != nil {
		return nil, fmt.Errorf("cannot get chain config: %w", err)
	}

	s.chainConfigs[fork] = cfg
	return cfg, nil
}
