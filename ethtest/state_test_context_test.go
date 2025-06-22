package ethtest

import (
	"fmt"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/params"
	"github.com/stretchr/testify/assert"
)

func newTestStateTestContext() *StateTestContext {
	msg := &core.Message{}
	alloc := types.GenesisAlloc{}
	stJson := &stJSON{
		path: "testpath.json",
		Pre:  alloc,
	}
	post := stPost{
		RootHash:        common.HexToHash("0x1234"),
		ExpectException: "error",
		TxBytes:         hexutil.Bytes{0x01, 0x02},
		LogsHash:        common.HexToHash("0xabcd"),
	}
	chainCfg := &params.ChainConfig{}
	ctx := newStateTestTxContext(stJson, msg, post, chainCfg, "label", "fork", 7)
	return ctx.(*StateTestContext)
}

func TestStateTestContext_GetTxBytes(t *testing.T) {
	stCtx := newTestStateTestContext()
	assert.Equal(t, hexutil.Bytes{0x01, 0x02}, stCtx.GetTxBytes())
}

func TestStateTestContext_GetLogsHash(t *testing.T) {
	stCtx := newTestStateTestContext()
	assert.Equal(t, common.HexToHash("0xabcd"), stCtx.GetLogsHash())
}

func TestStateTestContext_GetStateHash(t *testing.T) {
	stCtx := newTestStateTestContext()
	assert.Equal(t, common.HexToHash("0x1234"), stCtx.GetStateHash())
}

func TestStateTestContext_GetOutputState(t *testing.T) {
	stCtx := newTestStateTestContext()
	assert.Nil(t, stCtx.GetOutputState())
}

func TestStateTestContext_GetInputState(t *testing.T) {
	stCtx := newTestStateTestContext()
	assert.NotNil(t, stCtx.GetInputState())
}

func TestStateTestContext_GetBlockEnvironment(t *testing.T) {
	stCtx := newTestStateTestContext()
	assert.NotNil(t, stCtx.GetBlockEnvironment())
}

func TestStateTestContext_GetMessage(t *testing.T) {
	stCtx := newTestStateTestContext()
	assert.NotNil(t, stCtx.GetMessage())
}

func TestStateTestContext_GetResult(t *testing.T) {
	stCtx := newTestStateTestContext()
	res := stCtx.GetResult()
	assert.Equal(t, "error", res.(stateTestResult).expectedErr)
}

func TestStateTestContext_String(t *testing.T) {
	stCtx := newTestStateTestContext()
	str := stCtx.String()
	assert.Contains(t, str, "testpath.json")
	assert.Contains(t, str, "label")
	assert.Contains(t, str, "fork")
	assert.Contains(t, str, fmt.Sprint(7))
}
