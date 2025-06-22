package ethtest

import (
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/params"
	"github.com/stretchr/testify/assert"
)

func TestStJSON_setPath(t *testing.T) {
	s := &stJSON{}
	s.setPath("abc.json")
	assert.Equal(t, "abc.json", s.path)
}

func TestStJSON_setTestLabel(t *testing.T) {
	s := &stJSON{}
	s.setTestLabel("label1")
	assert.Equal(t, "label1", s.testLabel)
}

func TestStJSON_CreateEnv(t *testing.T) {
	chainCfg := &params.ChainConfig{ChainID: new(big.Int).SetInt64(1)}
	s := &stJSON{
		Env: stBlockEnvironment{},
	}
	env := s.CreateEnv(chainCfg, "fork1")
	assert.Equal(t, chainCfg, env.chainCfg)
	assert.Equal(t, "fork1", env.fork)
}

func TestStPost_Fields(t *testing.T) {
	post := stPost{
		RootHash:        common.HexToHash("0x1234"),
		LogsHash:        common.HexToHash("0xabcd"),
		TxBytes:         hexutil.Bytes{0x01, 0x02},
		ExpectException: "err",
		Indexes:         Index{Data: 1, Gas: 2, Value: 3},
	}
	assert.Equal(t, common.HexToHash("0x1234"), post.RootHash)
	assert.Equal(t, common.HexToHash("0xabcd"), post.LogsHash)
	assert.Equal(t, hexutil.Bytes{0x01, 0x02}, post.TxBytes)
	assert.Equal(t, "err", post.ExpectException)
	assert.Equal(t, 1, post.Indexes.Data)
	assert.Equal(t, 2, post.Indexes.Gas)
	assert.Equal(t, 3, post.Indexes.Value)
}

func TestIndex_Fields(t *testing.T) {
	idx := Index{Data: 5, Gas: 6, Value: 7}
	assert.Equal(t, 5, idx.Data)
	assert.Equal(t, 6, idx.Gas)
	assert.Equal(t, 7, idx.Value)
}
