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
	"testing"

	"github.com/0xsoniclabs/aida/config"
	"github.com/0xsoniclabs/aida/logger"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/params"
	"github.com/stretchr/testify/assert"
)

func TestTransaction_Fields(t *testing.T) {
	tx := Transaction{Fork: "London", Ctx: nil}
	assert.Equal(t, "London", tx.Fork)
	assert.Nil(t, tx.Ctx)
}

func TestSortForks_All(t *testing.T) {
	log := logger.NewLogger("info", "test-sort-forks")
	forks := sortForks(log, "all")
	assert.ElementsMatch(t, forks, []string{
		"Prague", "Cancun", "Shanghai", "Paris", "Bellatrix", "GrayGlacier", "ArrowGlacier", "Altair", "London", "Berlin", "Istanbul", "MuirGlacier", "TestNetwork",
	})
}

func TestSortForks_Unknown(t *testing.T) {
	log := logger.NewLogger("info", "test-sort-forks")
	forks := sortForks(log, "unknownFork")
	assert.Empty(t, forks)
}

func TestSortForks_Single(t *testing.T) {
	log := logger.NewLogger("info", "test-sort-forks")
	forks := sortForks(log, "London")
	assert.Equal(t, []string{"London"}, forks)
}

func TestTestCaseSplitter_getChainConfig(t *testing.T) {
	ts := &TestCaseSplitter{chainConfigs: make(map[string]*params.ChainConfig)}
	cfg, err := ts.getChainConfig("London")
	assert.NoError(t, err)
	assert.NotNil(t, cfg)
	cfg2, err2 := ts.getChainConfig("London")
	assert.NoError(t, err2)
	assert.Equal(t, cfg, cfg2)
}

func TestTestCaseSplitter_SplitStateTests_Empty(t *testing.T) {
	ts := &TestCaseSplitter{jsons: []*stJSON{}, enabledForks: []string{"London"}, chainConfigs: make(map[string]*params.ChainConfig), log: logger.NewLogger("info", "splitter")}
	dt, err := ts.SplitStateTests()
	assert.NoError(t, err)
	assert.Empty(t, dt)
}

func TestTestCaseSplitter_SplitStateTests_BaseFeeNil(t *testing.T) {
	stJson := &stJSON{
		Env:  stBlockEnvironment{BaseFee: nil},
		Post: map[string][]stPost{},
	}
	ts := &TestCaseSplitter{
		jsons:        []*stJSON{stJson},
		enabledForks: []string{"London"},
		chainConfigs: make(map[string]*params.ChainConfig),
		log:          logger.NewLogger("info", "splitter"),
	}
	dt, err := ts.SplitStateTests()
	assert.NoError(t, err)
	assert.Empty(t, dt)
}

func TestNewTestCaseSplitter_Error(t *testing.T) {
	cfg := &config.Config{LogLevel: "info", Fork: "London"}
	_, err := NewTestCaseSplitter(cfg)
	assert.Error(t, err)
}

func TestTestCaseSplitter_SplitStateTests_PostNotFound(t *testing.T) {
	stJson := &stJSON{
		Env:  stBlockEnvironment{BaseFee: newBigInt(1)},
		Post: map[string][]stPost{},
	}
	ts := &TestCaseSplitter{
		jsons:        []*stJSON{stJson},
		enabledForks: []string{"London"},
		chainConfigs: make(map[string]*params.ChainConfig),
		log:          logger.NewLogger("info", "splitter"),
	}
	dt, err := ts.SplitStateTests()
	assert.NoError(t, err)
	assert.Empty(t, dt)
}

func TestTestCaseSplitter_SplitStateTests(t *testing.T) {
	stJson := &stJSON{
		Env: stBlockEnvironment{BaseFee: newBigInt(1)},
		Tx: stTransaction{
			Data:                 []string{"0x1234"},
			Value:                []string{"1234"},
			GasLimit:             []*BigInt{newBigInt(1000000)},
			Nonce:                newBigInt(1),
			MaxFeePerGas:         newBigInt(1),
			MaxPriorityFeePerGas: newBigInt(1),
			BlobGasFeeCap:        newBigInt(1),
		},
		Post: map[string][]stPost{
			"London": {
				{
					RootHash:        common.HexToHash("0x1234"),
					LogsHash:        common.HexToHash("0xabcd"),
					TxBytes:         hexutil.Bytes{0x01, 0x02},
					ExpectException: "err",
					Indexes:         Index{Data: 0, Gas: 0, Value: 0},
				},
			},
			"Paris": {
				{
					RootHash:        common.HexToHash("0x1234"),
					LogsHash:        common.HexToHash("0xabcd"),
					TxBytes:         hexutil.Bytes{0x01, 0x02},
					ExpectException: "err",
					Indexes:         Index{Data: 0, Gas: 0, Value: 0},
				},
			},
		},
	}
	ts := &TestCaseSplitter{
		jsons:        []*stJSON{stJson},
		enabledForks: []string{"London", "Paris"},
		chainConfigs: make(map[string]*params.ChainConfig),
		log:          logger.NewLogger("info", "splitter"),
	}
	dt, err := ts.SplitStateTests()
	assert.NoError(t, err)
	assert.Len(t, dt, len(ts.enabledForks))
}
