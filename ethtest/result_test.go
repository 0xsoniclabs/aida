package ethtest

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStateTestResult_GetReceipt(t *testing.T) {
	res := stateTestResult{}
	assert.Nil(t, res.GetReceipt())
}

func TestStateTestResult_GetRawResult_NoError(t *testing.T) {
	res := stateTestResult{expectedErr: ""}
	result, err := res.GetRawResult()
	assert.Nil(t, result)
	assert.NoError(t, err)
}

func TestStateTestResult_GetRawResult_WithError(t *testing.T) {
	errMsg := "some error"
	res := stateTestResult{expectedErr: errMsg}
	result, err := res.GetRawResult()
	assert.Nil(t, result)
	assert.Error(t, err)
	assert.EqualError(t, err, errMsg)
}

func TestStateTestResult_GetGasUsed(t *testing.T) {
	res := stateTestResult{}
	assert.Equal(t, uint64(0), res.GetGasUsed())
}
