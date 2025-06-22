package ethtest

import (
	"testing"

	"github.com/0xsoniclabs/aida/utils"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"

	"github.com/0xsoniclabs/aida/logger"
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
	cfg := &utils.Config{LogLevel: "info", Fork: "London"}
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
