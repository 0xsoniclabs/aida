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

package utils

import (
	"flag"
	"fmt"
	"github.com/stretchr/testify/require"
	"math"
	"math/big"
	"os"
	"reflect"
	"strconv"
	"testing"

	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/stretchr/testify/assert"

	"github.com/0xsoniclabs/aida/logger"
	"github.com/0xsoniclabs/substate/db"
	"github.com/0xsoniclabs/tosca/go/tosca"
	"github.com/urfave/cli/v2"
)

func prepareMockCliContext() *cli.Context {
	flagSet := flag.NewFlagSet("utils_config_test", 0)
	flagSet.Uint64(SyncPeriodLengthFlag.Name, 1000, "Number of blocks")
	flagSet.Bool(ValidateFlag.Name, true, "enables validation")
	flagSet.Bool(ValidateTxStateFlag.Name, true, "enables transaction state validation")
	flagSet.Bool(ContinueOnFailureFlag.Name, true, "continue execute after validation failure detected")
	flagSet.String(AidaDbFlag.Name, "./test.db", "set substate, updateset and deleted accounts directory")
	flagSet.String(logger.LogLevelFlag.Name, "info", "Level of the logging of the app action (\"critical\", \"error\", \"warning\", \"notice\", \"info\", \"debug\"; default: NOTICE)")

	ctx := cli.NewContext(cli.NewApp(), flagSet, nil)

	command := &cli.Command{Name: "test_command"}
	ctx.Command = command

	return ctx
}

func TestUtilsConfig_GetChainConfig(t *testing.T) {
	testCases := []ChainID{
		TestnetChainID,
		MainnetChainID,
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("ChainID: %d", tc), func(t *testing.T) {
			chainConfig, err := getChainConfig(tc, "")
			if err != nil {
				t.Fatalf("cannot get chain config: %v", err)
			}

			if tc == MainnetChainID && chainConfig.BerlinBlock.Cmp(new(big.Int).SetUint64(37455223)) != 0 {
				t.Fatalf("Incorrect Berlin fork block on chainID: %d; Block number: %d, should be: %d", MainnetChainID, chainConfig.BerlinBlock, 37455223)
			}

			if tc == MainnetChainID && chainConfig.LondonBlock.Cmp(new(big.Int).SetUint64(37534833)) != 0 {
				t.Fatalf("Incorrect London fork block on chainID: %d; Block number: %d, should be: %d", MainnetChainID, chainConfig.LondonBlock, 37534833)
			}

			if tc == TestnetChainID && chainConfig.BerlinBlock.Cmp(new(big.Int).SetUint64(1559470)) != 0 {
				t.Fatalf("Incorrect Berlin fork block on chainID: %d; Block number: %d, should be: %d", TestnetChainID, chainConfig.BerlinBlock, 1559470)
			}

			if tc == TestnetChainID && chainConfig.LondonBlock.Cmp(new(big.Int).SetUint64(7513335)) != 0 {
				t.Fatalf("Incorrect London fork block on chainID: %d; Block number: %d, should be: %d", TestnetChainID, chainConfig.LondonBlock, 7513335)
			}
		})
	}
}

func TestUtilsConfig_NewConfig(t *testing.T) {
	ctx := prepareMockCliContext()

	_, err := NewConfig(ctx, NoArgs)
	if err != nil {
		t.Fatalf("Failed to create new config: %v", err)
	}
}

func TestUtilsConfig_SetBlockRange(t *testing.T) {
	first, last, err := SetBlockRange("0", "40000000", 0)
	if err != nil {
		t.Fatalf("Failed to set block range (0-40000000): %v", err)
	}

	if first != uint64(0) {
		t.Fatalf("Failed to parse first block; expected: %d, have: %d", 0, first)
	}

	if last != uint64(40_000_000) {
		t.Fatalf("Failed to parse last block; expected: %d, have: %d", 40_000_000, last)
	}

	first, last, err = SetBlockRange("OpeRa", "berlin", MainnetChainID)
	if err != nil {
		t.Fatalf("Failed to set block range (opera-berlin on mainnet): %v", err)
	}

	if first != uint64(4_564_026) {
		t.Fatalf("Failed to parse first block; expected: %d, have: %d", 4_564_026, first)
	}

	if last != uint64(37_455_223) {
		t.Fatalf("Failed to parse last block; expected: %d, have: %d", 37_455_223, last)
	}

	first, last, err = SetBlockRange("zero", "London", TestnetChainID)
	if err != nil {
		t.Fatalf("Failed to set block range (zero-london on testnet): %v", err)
	}

	if first != uint64(0) {
		t.Fatalf("Failed to parse first block; expected: %d, have: %d", 0, first)
	}

	if last != uint64(7_513_335) {
		t.Fatalf("Failed to parse last block; expected: %d, have: %d", 7_513_335, last)
	}

	// test addition/subtraction
	first, last, err = SetBlockRange("opera+23456", "London-100", TestnetChainID)
	if err != nil {
		t.Fatalf("Failed to set block range (opera+23456-London-100 on mainnet): %v", err)
	}

	if first != uint64(502_783) {
		t.Fatalf("Failed to parse first block; expected: %d, have: %d", 502_783, first)
	}

	if last != uint64(7_513_235) {
		t.Fatalf("Failed to parse last block; expected: %d, have: %d", 7_513_235, last)
	}

	// test upper/lower cases
	first, last, err = SetBlockRange("berlin-1000", "LonDoN", MainnetChainID)
	if err != nil {
		t.Fatalf("Failed to set block range (berlin-1000-LonDoN on mainnet): %v", err)
	}

	if first != uint64(37_454_223) {
		t.Fatalf("Failed to parse first block; expected: %d, have: %d", 37_454_223, first)
	}

	if last != uint64(37_534_833) {
		t.Fatalf("Failed to parse last block; expected: %d, have: %d", 37_534_833, last)
	}

	// test first and last keyword. Since no metadata, default values are expected
	first, last, err = SetBlockRange("first", "last", MainnetChainID)
	if err != nil {
		t.Fatalf("Failed to set block range (first-last on mainnet): %v", err)
	}

	if first != uint64(0) {
		t.Fatalf("Failed to parse first block; expected: %d, have: %d", 0, first)
	}

	if last != maxLastBlock {
		t.Fatalf("Failed to parse last block; expected: %v, have: %v", uint64(math.MaxUint64), last)
	}

	// test lastpatch and last keyword
	first, last, err = SetBlockRange("lastpatch", "last", MainnetChainID)
	if err != nil {
		t.Fatalf("Failed to set block range (lastpatch-last on mainnet): %v", err)
	}

	if first != uint64(0) {
		t.Fatalf("Failed to parse first block; expected: %d, have: %d", 0, first)
	}

	if last != maxLastBlock {
		t.Fatalf("Failed to parse last block; expected: %v, have: %v", uint64(math.MaxUint64), last)
	}
}

func TestUtilsConfig_SetInvalidBlockRange(t *testing.T) {
	_, _, err := SetBlockRange("test", "40000000", 0)
	if err == nil {
		t.Fatalf("Failed to throw an error")
	}

	_, _, err = SetBlockRange("1000", "0", TestnetChainID)
	if err == nil {
		t.Fatalf("Failed to throw an error")
	}

	_, _, err = SetBlockRange("tokyo", "berlin", MainnetChainID)
	if err == nil {
		t.Fatalf("Failed to throw an error")
	}

	_, _, err = SetBlockRange("tokyo", "berlin", TestnetChainID)
	if err == nil {
		t.Fatalf("Failed to throw an error")
	}

	_, _, err = SetBlockRange("london-opera", "opera+london", MainnetChainID)
	if err == nil {
		t.Fatalf("Failed to throw an error")
	}

	_, _, err = SetBlockRange("london-opera", "opera+london", TestnetChainID)
	if err == nil {
		t.Fatalf("Failed to throw an error")
	}
}

func TestUtilsConfig_SetBlockRangeLastSmallerThanFirst(t *testing.T) {
	_, _, err := SetBlockRange("5", "0", 0)
	if err == nil {
		t.Fatalf("Failed to throw an error when last block number is smaller than first")
	}
}

func TestUtilsConfig_adjustBlockRange(t *testing.T) {
	var (
		chainId           ChainID
		first, last       uint64
		firstArg, lastArg uint64
		err               error
	)
	chainId = MainnetChainID
	KeywordBlocks[chainId]["first"] = 1000
	KeywordBlocks[chainId]["last"] = 2000

	cfg := &Config{ChainID: chainId, LogLevel: "NOTICE"}
	cc := NewConfigContext(cfg, nil)

	firstArg = 1100
	lastArg = 1900
	first, last, err = cc.adjustBlockRange(firstArg, lastArg)
	if err != nil {
		t.Fatalf("unexpected error; %v", err)
	}
	if first != firstArg && last != lastArg {
		t.Fatalf("wrong block range; expected %v:%v, have %v:%v", firstArg, lastArg, first, last)
	}

	firstArg = 3000
	lastArg = 4000
	first, last, err = cc.adjustBlockRange(firstArg, lastArg)
	if first != 0 && last != 0 {
		t.Fatalf("wrong block range; expected %v:%v, have %v:%v", 0, 0, first, last)
	}
	if err == nil {
		t.Fatalf("Ranges not overlapped. Expected an error.")
	}

	// check corner cases
	firstArg = 100
	lastArg = 1000
	first, last, err = cc.adjustBlockRange(firstArg, lastArg)
	if err != nil {
		t.Fatalf("unexpected error; %v", err)
	}
	if first != firstArg && last != lastArg {
		t.Fatalf("wrong block range; expected %v:%v, have %v:%v", lastArg, lastArg, first, last)
	}

	firstArg = 2000
	lastArg = 2200
	first, last, err = cc.adjustBlockRange(firstArg, lastArg)
	if err != nil {
		t.Fatalf("unexpected error; %v", err)
	}
	if first != firstArg && last != lastArg {
		t.Fatalf("wrong block range; expected %v:%v, have %v:%v", firstArg, firstArg, first, last)
	}
	// reset keywords for the following tests
	KeywordBlocks[chainId]["first"] = 0
	KeywordBlocks[chainId]["last"] = math.MaxUint64
}

func TestUtilsConfig_getMdBlockRange(t *testing.T) {
	// prepare components
	// create new leveldb
	var (
		logLevel   = "NOTICE"
		firstBlock = KeywordBlocks[MainnetChainID]["opera"]
		lastBlock  = uint64(20001704)
		chainId    = MainnetChainID
	)
	// prepare mock config
	cfg := &Config{AidaDb: "./test.db", LogLevel: logLevel, ChainID: chainId}

	// prepare config context
	cc := NewConfigContext(cfg, nil)

	// prepare fake AidaDb
	err := createFakeAidaDb(cfg)
	if err != nil {
		t.Fatalf("cannot create fake AidaDb; %v", err)
	}

	defer func() {
		err = os.RemoveAll(cfg.AidaDb)
		if err != nil {
			t.Fatalf("cannot remove db; %v", err)
		}
	}()
	cfg.AidaDb = "./test-wrong.db" //db doesn't exist

	// test getMdBlockRange
	// getMdBlockRange returns default values if unable to open
	first, last, lastpatch, err := cc.getMdBlockRange()
	if cc.hasMetadata || first != uint64(0) || last != math.MaxUint64 {
		t.Fatalf("wrong block range; expected %v:%v, have %v:%v", 0, uint64(math.MaxUint64), first, last)
	} else if err != nil {
		t.Fatalf("unexpected error; %v", err)
	} else if lastpatch != uint64(0) {
		t.Fatalf("wrong last patch block; expected %v, have %v", 0, lastpatch)
	}

	cfg.AidaDb = "./test.db" //db exists
	// open an existing AidaDb
	err = cc.setAidaDbRepositoryUrl()
	if err != nil {
		t.Fatalf("cannot set AidaDb repository url; %v", err)
	}
	first, last, lastpatch, err = cc.getMdBlockRange()
	if !cc.hasMetadata || first != firstBlock || last != lastBlock {
		t.Fatalf("wrong block range; expected %v:%v, have %v:%v", firstBlock, lastBlock, first, last)
	} else if err != nil {
		t.Fatalf("unexpected error; %v", err)
	} else if lastpatch != uint64(45640256) {
		t.Fatalf("wrong last patch block; expected %v, have %v", 45640256, lastpatch)
	}

	// aida url is not set; expected lastpatch is 0.
	AidaDbRepositoryUrl = ""
	first, last, lastpatch, err = cc.getMdBlockRange()
	if !cc.hasMetadata || first != firstBlock || last != lastBlock {
		t.Fatalf("wrong block range; expected %v:%v, have %v:%v", firstBlock, lastBlock, first, last)
	} else if err != nil {
		t.Fatalf("unexpected error; %v", err)
	} else if lastpatch != uint64(0) {
		t.Fatalf("wrong last patch block; expected %v, have %v", 0, lastpatch)
	}
}

// TestUtilsConfig_VmImplsAreRegistered checks if interpreters are correctly registered
func TestUtilsConfig_VmImplsAreRegistered(t *testing.T) {
	checkedImpls := []string{"lfvm", "lfvm-si", "evmzero", "evmone"}
	for _, interpreterImpl := range checkedImpls {
		factory := tosca.GetInterpreterFactory(interpreterImpl)
		if factory == nil {
			t.Errorf("interpreter %q is not registered", interpreterImpl)
		}
	}
}

// TestUtilsConfig_setChainIdFromDB tests if chainID can be loaded from AidaDB correctly
func TestUtilsConfig_setChainIdFromDB(t *testing.T) {
	// prepare components
	// create new leveldb
	var (
		logLevel = "NOTICE"
		chainId  = MainnetChainID
	)

	// prepare mock config
	cfg := &Config{AidaDb: "./test.db", LogLevel: logLevel}

	// prepare config context
	cc := NewConfigContext(cfg, &cli.Context{Command: &cli.Command{Name: "fake-name"}})

	// prepare fake AidaDb
	err := createFakeAidaDb(cfg)
	if err != nil {
		t.Fatalf("cannot create fake AidaDb; %v", err)
	}

	defer func() {
		err = os.RemoveAll(cfg.AidaDb)
		if err != nil {
			t.Fatalf("cannot remove db; %v", err)
		}
	}()

	// test getChainId function
	err = cc.setChainId()
	if err != nil {
		t.Fatalf("cannot get chain ID; %v", err)
	}

	if cfg.ChainID != chainId {
		t.Fatalf("failed to get chainId correctly from AidaDB; got: %v; expected: %v", cfg.ChainID, chainId)
	}
}

// TestUtilsConfig_getChainIdFromFlag tests if chainID can be loaded from flag correctly
func TestUtilsConfig_setChainIdFromFlag(t *testing.T) {
	// prepare components
	var (
		err      error
		logLevel = "NOTICE"
		chainId  = MainnetChainID
	)

	// prepare mock config
	cfg := &Config{AidaDb: "./test.db", LogLevel: logLevel, ChainID: chainId}

	// create config context
	cc := NewConfigContext(cfg, nil)

	// test getChainId function
	err = cc.setChainId()
	if err != nil {
		t.Fatalf("cannot get chain ID; %v", err)
	}

	if cfg.ChainID != chainId {
		t.Fatalf("failed to get chainId correctly from AidaDB; got: %v; expected: %v", cfg.ChainID, chainId)
	}
}

// TestUtilsConfig_getDefaultChainId tests if unknown chainID will be replaced with the mainnet chainID
func TestUtilsConfig_setDefaultChainId(t *testing.T) {
	// prepare components
	var (
		err      error
		logLevel = "NOTICE"
		chainId  = MainnetChainID
	)

	// prepare mock config
	cfg := &Config{AidaDb: "./test.db", LogLevel: logLevel}

	// create config context
	cc := NewConfigContext(cfg, &cli.Context{Command: &cli.Command{Name: "fake-name"}})

	// test getChainId function
	err = cc.setChainId()
	if err != nil {
		t.Fatalf("cannot get chain ID; %v", err)
	}

	if cfg.ChainID != chainId {
		t.Fatalf("failed to get chainId correctly from AidaDB; got: %v; expected: %v", cfg.ChainID, chainId)
	}
}

// TestUtilsConfig_updateConfigBlockRangeBlockRange tests correct parsing of cli arguments for block range
func TestUtilsConfig_updateConfigBlockRangeBlockRange(t *testing.T) {
	// prepare components
	var (
		logLevel = "NOTICE"
		mode     = BlockRangeArgs
		firstArg = "4564026"
		lastArg  = "5000000"
	)

	// prepare mock config
	cfg := &Config{AidaDb: "./test.db", LogLevel: logLevel, ChainID: MainnetChainID}

	// create config context
	cc := NewConfigContext(cfg, nil)

	// prepare fake AidaDb
	err := createFakeAidaDb(cfg)
	if err != nil {
		t.Fatalf("cannot create fake AidaDb; %v", err)
	}

	defer func() {
		err = os.RemoveAll(cfg.AidaDb)
		if err != nil {
			t.Fatalf("cannot remove db; %v", err)
		}
	}()

	// parse cli arguments slice
	err = cc.updateConfigBlockRange([]string{firstArg, lastArg}, mode)
	if err != nil {
		t.Fatalf("cannot parse the cli arguments; %v", err)
	}

	// check if the arguments were parsed correctly
	if parsedFirst, _ := strconv.ParseUint(firstArg, 10, 64); parsedFirst != cfg.First {
		t.Fatalf("failed to get first argument correctly; got: %d; expected: %s", parsedFirst, firstArg)
	}

	if parsedLast, _ := strconv.ParseUint(lastArg, 10, 64); parsedLast != cfg.Last {
		t.Fatalf("failed to get last argument correctly; got: %d; expected: %s", parsedLast, lastArg)
	}
}

// TestUtilsConfig_updateConfigBlockRangeBlockRangeInvalid tests parsing of invalid cli arguments length for block range
func TestUtilsConfig_updateConfigBlockRangeBlockRangeInvalid(t *testing.T) {
	// prepare components
	var (
		mode     = BlockRangeArgs
		logLevel = "NOTICE"
	)

	// prepare mock config
	cfg := &Config{AidaDb: "./test.db", LogLevel: logLevel}

	// create config context
	cc := NewConfigContext(cfg, nil)

	// parse cli arguments slice of insufficient length
	err := cc.updateConfigBlockRange([]string{"test"}, mode)
	if err == nil {
		t.Fatalf("failed to throw an error")
	}
}

// TestUtilsConfig_updateConfigBlockRangeLastBlock tests correct parsing of cli argument for last block number
func TestUtilsConfig_updateConfigBlockRangeLastBlock(t *testing.T) {
	// prepare components
	var (
		logLevel = "NOTICE"
		mode     = LastBlockArg
		lastArg  = "30"
	)

	// prepare mock config
	cfg := &Config{AidaDb: "./test.db", LogLevel: logLevel}

	// create config context
	cc := NewConfigContext(cfg, nil)

	// parse cli arguments slice
	err := cc.updateConfigBlockRange([]string{lastArg}, mode)
	if err != nil {
		t.Fatalf("cannot parse the cli arguments; %v", err)
	}

	// check if the arguments were parsed correctly
	if parsedLast, _ := strconv.ParseUint(lastArg, 10, 64); parsedLast != cfg.Last {
		t.Fatalf("failed to get last argument correctly; got: %d; expected: %s", parsedLast, lastArg)
	}
}

// TestUtilsConfig_updateConfigBlockRangeLastBlockInvalid tests parsing of invalid cli arguments length for last block number
func TestUtilsConfig_updateConfigBlockRangeLastBlockInvalid(t *testing.T) {
	// prepare components
	var (
		logLevel = "NOTICE"
		mode     = LastBlockArg
	)

	// prepare mock config
	cfg := &Config{AidaDb: "./test.db", LogLevel: logLevel}

	// create config context
	cc := NewConfigContext(cfg, nil)

	// parse cli arguments slice of insufficient length
	err := cc.updateConfigBlockRange([]string{"test"}, mode)
	if err == nil {
		t.Fatalf("failed to throw an error")
	}
}

// TestUtilsConfig_updateConfigBlockRangeOneToNInvalid tests parsing of invalid cli arguments length for last block number
func TestUtilsConfig_updateConfigBlockRangeOneToNInvalid(t *testing.T) {
	// prepare components
	var (
		logLevel = "NOTICE"
		mode     = OneToNArgs
	)

	// prepare mock config
	cfg := &Config{AidaDb: "./test.db", LogLevel: logLevel}

	// create config context
	cc := NewConfigContext(cfg, nil)

	// parse cli arguments slice of insufficient length
	err := cc.updateConfigBlockRange([]string{}, mode)
	if err == nil {
		t.Fatalf("failed to throw an error")
	}
}

// TestUtilsConfig_adjustMissingConfigValues tests if missing config values are set correctly
func TestUtilsConfig_adjustMissingConfigValues(t *testing.T) {
	// prepare components
	var (
		chainId           = MainnetChainID
		aidaDB            = "./test.db"
		dbImpl            = "carmen"
		dbVariant         = ""
		randomSeed int64  = -1
		first      uint64 = 0
	)

	// prepare mock config
	cfg := &Config{
		ChainID:    chainId,
		AidaDb:     aidaDB,
		DbImpl:     dbImpl,
		DbVariant:  dbVariant,
		RandomSeed: randomSeed,
		First:      first,
		LogLevel:   "NOTICE",
	}

	// create config context
	cc := NewConfigContext(cfg, nil)

	// prepare fake AidaDb
	err := createFakeAidaDb(cfg)
	if err != nil {
		t.Fatalf("cannot create fake AidaDb; %v", err)
	}

	defer func() {
		err = os.RemoveAll(cfg.AidaDb)
		if err != nil {
			t.Fatalf("cannot remove db; %v", err)
		}
	}()

	// set missing values
	err = cc.adjustMissingConfigValues()
	if err != nil {
		t.Fatalf("failed to adjust missing config values; %v", err)
	}

	// checks
	if cfg.DbVariant == dbVariant {
		t.Fatalf("failed to adjust carmen DBvariant; got: %s; expected: %s", cfg.DbVariant, dbVariant)
	}

	if cfg.RandomSeed <= 0 {
		t.Fatalf("failed to adjust random seed value; got: %d; expected: Random int64 greater than 0", cfg.RandomSeed)
	}

	if cfg.DeletionDb != cfg.AidaDb {
		t.Fatalf("failed to adjust deletion db path; got: %s; expected: %s", cfg.DeletionDb, aidaDB)
	}

	if cfg.SubstateDb != cfg.AidaDb {
		t.Fatalf("failed to adjust substate db path; got: %s; expected: %s", cfg.SubstateDb, aidaDB)
	}

	if cfg.UpdateDb != cfg.AidaDb {
		t.Fatalf("failed to adjust update db path; got: %s; expected: %s", cfg.UpdateDb, aidaDB)
	}
}

// TestUtilsConfig_adjustMissingConfigValuesValidationOn tests if missing config validation values are set correctly
func TestUtilsConfig_adjustMissingConfigValuesValidationOn(t *testing.T) {
	// prepare mock configs
	for _, cfg := range []Config{
		{
			Validate:          true,
			ValidateTxState:   false,
			ContinueOnFailure: false,
		},
		{
			Validate:          false,
			ValidateTxState:   true,
			ContinueOnFailure: false,
		},
		{
			Validate:          false,
			ValidateTxState:   false,
			ContinueOnFailure: true,
		},
		{
			Validate:          false,
			ValidateTxState:   true,
			ContinueOnFailure: true,
		},
		{
			Validate:          true,
			ValidateTxState:   true,
			ContinueOnFailure: true,
		},
	} {
		t.Run("validation adjustment", func(t *testing.T) {
			// set missing values
			cc := NewConfigContext(&cfg, nil)
			err := cc.adjustMissingConfigValues()
			if err != nil {
				t.Fatalf("failed to adjust missing config values; %v", err)
			}

			// checks
			if !cfg.ValidateTxState {
				t.Fatalf("failed to adjust ValidateTxState; got: %v; expected: %v", cfg.ValidateTxState, true)
			}

		})
	}
}

// TestUtilsConfig_adjustMissingConfigValuesValidationOff tests if missing config validation values are set correctly
func TestUtilsConfig_adjustMissingConfigValuesValidationOff(t *testing.T) {
	// prepare mock config
	cfg := &Config{
		Validate:          false,
		ValidateTxState:   false,
		ContinueOnFailure: false,
		LogLevel:          "NOTICE",
	}

	// prepare config context
	cc := NewConfigContext(cfg, nil)

	err := cc.adjustMissingConfigValues()
	if err != nil {
		t.Fatalf("failed to adjust missing config values; %v", err)
	}

	// checks
	if cfg.ValidateTxState {
		t.Fatalf("failed to adjust ValidateTxState; got: %v; expected: %v", cfg.ValidateTxState, true)
	}

}

// TestUtilsConfig_adjustMissingConfigValuesKeepStateDb tests if missing config keepDb value is set correctly
func TestUtilsConfig_adjustMissingConfigValuesKeepStateDb(t *testing.T) {
	// prepare mock config
	cfg := &Config{
		KeepDb:    true,
		DbVariant: "go-memory",
		LogLevel:  "NOTICE",
	}

	// prepare config context
	cc := NewConfigContext(cfg, nil)

	err := cc.adjustMissingConfigValues()
	if err != nil {
		t.Fatalf("failed to adjust missing config values; %v", err)
	}

	// checks
	if cfg.KeepDb {
		t.Fatalf("failed to adjust KeepDb value; got: %v; expected: %v", cfg.KeepDb, false)
	}
}

// TestUtilsConfig_adjustMissingConfigValuesWrongDbTmp tests if temporary db path doesn't exist, system temp location is used instead.
func TestUtilsConfig_adjustMissingConfigValuesWrongDbTmp(t *testing.T) {
	// prepare mock config
	cfg := &Config{
		DbTmp:    "./wrong_path",
		LogLevel: "NOTICE",
	}

	// prepare config context
	cc := NewConfigContext(cfg, nil)

	err := cc.adjustMissingConfigValues()
	if err != nil {
		t.Fatalf("failed to adjust missing config values; %v", err)
	}

	// checks
	if cfg.DbTmp != os.TempDir() {
		t.Fatalf("failed to adjust temporary database location; got: %v; expected: %v", cfg.DbTmp, os.TempDir())
	}
}

// TestUtilsConfig_ToTitleCase_Success tests if ToTitleCase function works correctly.
// It should return a string with the first letter capitalized.
// If a word Glaciers is contained, it should be returned as Glaciers.
func TestUtilsConfig_ToTitleCase_Success(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{
			input: "newglaciers",
			want:  "NewGlaciers",
		},
		{
			input: "all",
			want:  "All",
		},
		{
			input: "unKnoWn",
			want:  "Unknown",
		},
	}

	for _, test := range tests {
		got := ToTitleCase(test.input)
		if got != test.want {
			t.Fatalf("failed to capitalize first letter; got: %s; expected: %s", got, test.want)
		}
	}
}

func TestConfigContext_setVmConfig(t *testing.T) {
	for chainID, name := range RealChainIDs {
		t.Run(name, func(t *testing.T) {
			cfg := &Config{ChainID: chainID}
			ctx := NewConfigContext(cfg, nil)
			err := ctx.setVmConfig()
			require.NoError(t, err, "cannot set vm cfg")
			if IsEthereumNetwork(cfg.ChainID) {
				require.False(t, cfg.VmCfg.ChargeExcessGas)
				require.False(t, cfg.VmCfg.IgnoreGasFeeCap)
				require.False(t, cfg.VmCfg.InsufficientBalanceIsNotAnError)
				require.False(t, cfg.VmCfg.SkipTipPaymentToCoinbase)
			} else {
				require.True(t, cfg.VmCfg.ChargeExcessGas)
				require.True(t, cfg.VmCfg.IgnoreGasFeeCap)
				require.True(t, cfg.VmCfg.InsufficientBalanceIsNotAnError)
				require.True(t, cfg.VmCfg.SkipTipPaymentToCoinbase)
			}
		})
	}
}

func TestConfigContext_setVmConfig_EthereumEvmImpl(t *testing.T) {
	tests := []struct {
		name    string
		cfg     *Config
		require func(t require.TestingT, value bool, msgAndArgs ...any)
	}{
		{
			name:    "empty-evm-impl",
			cfg:     &Config{EvmImpl: ""},
			require: require.True,
		},
		{
			name:    "opera-evm-impl",
			cfg:     &Config{EvmImpl: "opera"},
			require: require.True,
		},
		{
			name:    "ethereum-evm-impl",
			cfg:     &Config{EvmImpl: "ethereum"},
			require: require.False,
		},
		{
			name:    "unknown",
			cfg:     &Config{EvmImpl: "unknown"},
			require: require.True,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctx := NewConfigContext(test.cfg, nil)
			err := ctx.setVmConfig()
			require.NoError(t, err, "cannot set vm cfg")
			test.require(t, test.cfg.VmCfg.ChargeExcessGas)
			test.require(t, test.cfg.VmCfg.IgnoreGasFeeCap)
			test.require(t, test.cfg.VmCfg.InsufficientBalanceIsNotAnError)
			test.require(t, test.cfg.VmCfg.SkipTipPaymentToCoinbase)
		})
	}
}

func TestConfigContext_setVmConfig_InvalidVmImplCausesError(t *testing.T) {
	cfg := &Config{VmImpl: "invalid"}
	ctx := NewConfigContext(cfg, nil)
	err := ctx.setVmConfig()
	require.Error(t, err, "error must be returned")
	require.Contains(t, err.Error(), "cannot get interpreter for \"invalid\"")
}

func TestNewTestConfig_CorrectlyFillsData(t *testing.T) {
	chainId := MainnetChainID
	first := uint64(123)
	last := uint64(456)
	fork := "london"

	cfg := NewTestConfig(t, chainId, first, last, true, fork)

	require.Equal(t, chainId, cfg.ChainID, "ChainID not set correctly")
	require.Equal(t, first, cfg.First, "First block not set correctly")
	require.Equal(t, last, cfg.Last, "Last block not set correctly")
	require.True(t, cfg.Validate, "Validate not set correctly")
	require.True(t, cfg.ValidateTxState, "ValidateTxState not set correctly")
	require.NotNil(t, cfg.chainCfg, "chainCfg should be set")
	require.Equal(t, "Critical", cfg.LogLevel, "LogLevel should be Critical")
	require.True(t, cfg.SkipPriming, "SkipPriming should be true")
	require.True(t, cfg.VmCfg.NoBaseFee, "VmCfg.NoBaseFee should be true")
	require.Nil(t, cfg.VmCfg.Tracer, "VmCfg.Tracer should be nil")
	require.Nil(t, cfg.VmCfg.Interpreter, "VmCfg.Interpreter should be nil")
}

// createFakeAidaDb creates fake empty aidaDB with metadata for testing purposes
func createFakeAidaDb(cfg *Config) error {
	// fake metadata values
	var (
		firstBlock        = KeywordBlocks[MainnetChainID]["opera"]
		lastBlock  uint64 = 20001704
		firstEpoch uint64 = 100
		lastEpoch  uint64 = 200
	)

	// open fake aidaDB
	testDb, err := db.NewDefaultBaseDB(cfg.AidaDb)
	if err != nil {
		return fmt.Errorf("cannot open patch db; %v", err)
	}

	// create fake metadata
	err = ProcessPatchLikeMetadata(testDb, cfg.LogLevel, firstBlock, lastBlock, firstEpoch, lastEpoch, cfg.ChainID, true, nil)
	if err != nil {
		return fmt.Errorf("cannot create a metadata; %v", err)
	}
	err = testDb.Close()
	if err != nil {
		return fmt.Errorf("cannot close db; %v", err)
	}

	return nil
}

func Test_SetChainConfig(t *testing.T) {
	// case 1
	cfg := configContext{
		cfg: &Config{
			ChainID: EthTestsChainID,
		},
	}
	err := cfg.setChainConfig()
	assert.NoError(t, err)

	// case 2
	cfg = configContext{
		cfg: &Config{
			ChainID: MainnetChainID,
		},
	}
	err = cfg.setChainConfig()
	assert.NoError(t, err)

	// case 3
	cfg = configContext{
		cfg: &Config{
			ChainID: ChainID(999),
		},
	}
	err = cfg.setChainConfig()
	assert.Error(t, err)
}

func Test_ReportNewConfig(t *testing.T) {
	cfg := configContext{
		log: logger.NewLogger("NOTICE", "Config"),
		cfg: &Config{
			Profile:        true,
			ProfileFile:    "test.prof",
			ProfileSqlite3: "test.db",
			RegisterRun:    "test",
			ShadowDb:       true,
			DbLogging:      "test.db",
		},
	}
	assert.NotPanicsf(t, cfg.reportNewConfig, "reportNewConfig panics")
}

func Test_AdjustMissingConfigValues(t *testing.T) {
	cfg := configContext{
		log: logger.NewLogger("NOTICE", "Config"),
		cfg: &Config{
			ChainID:      MainnetChainID,
			AidaDb:       "./test.db",
			DbImpl:       "carmen",
			DbVariant:    "",
			RandomSeed:   -1,
			First:        0,
			LogLevel:     "NOTICE",
			ErrorLogging: "test.db",
		},
	}

	err := cfg.adjustMissingConfigValues()
	assert.NoError(t, err)
	assert.Equal(t, true, cfg.cfg.ContinueOnFailure)
}

func Test_IsEthereumNetwork(t *testing.T) {
	isEthereum := IsEthereumNetwork(EthereumChainID)
	assert.True(t, isEthereum)

	isEthereum = IsEthereumNetwork(HoleskyChainID)
	assert.True(t, isEthereum)

	isEthereum = IsEthereumNetwork(HoodiChainID)
	assert.True(t, isEthereum)

	isEthereum = IsEthereumNetwork(SepoliaChainID)
	assert.True(t, isEthereum)

	isEthereum = IsEthereumNetwork(ChainID(999))
	assert.False(t, isEthereum)
}

func Test_getChainConfig(t *testing.T) {
	chainConfig, err := getChainConfig(EthTestsChainID, "Frontier")
	assert.NoError(t, err)
	assert.Equal(t, big.NewInt(1), chainConfig.ChainID)

	chainConfig, err = getChainConfig(EthereumChainID, "Frontier")
	assert.NoError(t, err)
	assert.Equal(t, big.NewInt(1), chainConfig.ChainID)
	assert.Equal(t, false, chainConfig.DAOForkSupport)

	chainConfig, err = getChainConfig(HoleskyChainID, "Frontier")
	assert.NoError(t, err)
	assert.Equal(t, big.NewInt(17000), chainConfig.ChainID)
	assert.Equal(t, false, chainConfig.DAOForkSupport)

	chainConfig, err = getChainConfig(HoodiChainID, "Frontier")
	assert.NoError(t, err)
	assert.Equal(t, big.NewInt(560048), chainConfig.ChainID)
	assert.Equal(t, false, chainConfig.DAOForkSupport)

	chainConfig, err = getChainConfig(SepoliaChainID, "Frontier")
	assert.NoError(t, err)
	assert.Equal(t, big.NewInt(11155111), chainConfig.ChainID)
	assert.Equal(t, false, chainConfig.DAOForkSupport)

	chainConfig, err = getChainConfig(MainnetChainID, "Frontier")
	assert.NoError(t, err)
	assert.Equal(t, big.NewInt(250), chainConfig.ChainID)
	assert.Equal(t, false, chainConfig.DAOForkSupport)

	chainConfig, err = getChainConfig(ChainID(999), "Frontier")
	assert.Error(t, err)
	assert.Nil(t, chainConfig)
}

func Test_SetStateDbSrcReadOnly(t *testing.T) {
	cfg := Config{}
	cfg.SetStateDbSrcReadOnly()
	assert.Equal(t, true, cfg.StateDbSrcDirectAccess)
	assert.Equal(t, true, cfg.StateDbSrcReadOnly)
}

func Test_setAidaDbRepositoryUrl(t *testing.T) {
	testCases := []struct {
		name     string
		chainID  ChainID
		expected string
	}{
		{"SonicMainnet", SonicMainnetChainID, AidaDbRepositorySonicUrl},
		{"Mainnet", MainnetChainID, AidaDbRepositoryOperaUrl},
		{"Testnet", TestnetChainID, AidaDbRepositoryTestnetUrl},
		{"Ethereum", EthereumChainID, AidaDbRepositoryEthereumUrl},
		{"Holesky", HoleskyChainID, AidaDbRepositoryHoleskyUrl},
		{"Hoodi", HoodiChainID, AidaDbRepositoryHoodiUrl},
		{"Sepolia", SepoliaChainID, AidaDbRepositorySepoliaUrl},
		{"Unknown", ChainID(999), AidaDbRepositorySonicUrl}, // Unknown chain ID defaults to Sonic
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Reset the global URL before each test
			AidaDbRepositoryUrl = ""

			// Create config context with the test chain ID
			cc := configContext{
				log: logger.NewLogger("NOTICE", "Config"),
				cfg: &Config{
					ChainID: tc.chainID,
				},
			}

			// Call the method being tested
			err := cc.setAidaDbRepositoryUrl()

			// Verify results
			assert.NoError(t, err)
			assert.Equal(t, tc.expected, AidaDbRepositoryUrl)
		})
	}
}

func Test_setFirstOperaBlock(t *testing.T) {
	cfg := configContext{
		cfg: &Config{
			ChainID: MainnetChainID,
		},
	}

	err := cfg.setFirstOperaBlock()
	assert.NoError(t, err)

	cfg = configContext{
		cfg: &Config{
			ChainID: ChainID(999),
		},
	}
	err = cfg.setFirstOperaBlock()
	assert.Error(t, err)
}

func Test_GetInterpreterFactory(t *testing.T) {
	// case 1
	method := func(evm *vm.EVM) vm.Interpreter {
		return nil
	}
	cfg := &Config{
		interpreterFactory: method,
	}
	factory, err := cfg.GetInterpreterFactory()
	assert.NoError(t, err)
	assert.Equal(t, reflect.ValueOf(method).Pointer(), reflect.ValueOf(factory).Pointer())

	// case 2
	cfg = &Config{
		VmImpl: "",
	}
	factory, err = cfg.GetInterpreterFactory()
	assert.NoError(t, err)
	assert.Nil(t, factory)

	// case 3
	cfg = &Config{
		VmImpl: "unknown",
	}
	factory, err = cfg.GetInterpreterFactory()
	assert.Error(t, err)
	assert.Nil(t, factory)

	// case 4
	err = tosca.RegisterInterpreterFactory("unknown", func(config any) (tosca.Interpreter, error) {
		return nil, nil
	})
	if err != nil {
		t.Fatalf("failed to register interpreter factory: %v", err)
	}
	cfg = &Config{
		VmImpl: "unknown",
	}
	factory, err = cfg.GetInterpreterFactory()
	assert.Nil(t, err)
	assert.NotNil(t, factory)
}

func Test_GetChainConfig(t *testing.T) {
	cfg := &Config{
		ChainID: MainnetChainID,
	}

	chainConfig, err := cfg.GetChainConfig("")
	assert.NoError(t, err)
	assert.NotNil(t, chainConfig)
	assert.Equal(t, big.NewInt(250), chainConfig.ChainID) // Example value for MainnetChainID
}
