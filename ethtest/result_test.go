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
