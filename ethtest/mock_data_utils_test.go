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

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
)

func TestEthTest_CreateTestTransaction(t *testing.T) {
	tx := CreateTestTransaction(t)
	assert.NotNil(t, tx)
}

func TestEthTest_CreateTestStJson(t *testing.T) {
	st := CreateTestStJson(t)
	assert.NotNil(t, st)
	assert.Equal(t, "test/path", st.path)
}

func TestEthTest_CreateErrorTestTransaction(t *testing.T) {
	tx := CreateErrorTestTransaction(t)
	assert.NotNil(t, tx)
	st, ok := tx.(*StateTestContext)
	assert.True(t, ok)
	assert.Equal(t, "err", st.expectedError)
}

func TestEthTest_CreateNoErrorTestTransaction(t *testing.T) {
	tx := CreateNoErrorTestTransaction(t)
	assert.NotNil(t, tx)
	st, ok := tx.(*StateTestContext)
	assert.True(t, ok)
	assert.Equal(t, "", st.expectedError)
}

func TestEthTest_CreateTransactionThatFailsBlobGasExceedCheck(t *testing.T) {
	tx := CreateTransactionThatFailsBlobGasExceedCheck(t)
	assert.NotNil(t, tx)
	st, ok := tx.(*StateTestContext)
	assert.True(t, ok)
	assert.NotNil(t, st.msg)
	assert.GreaterOrEqual(t, len(st.msg.BlobHashes), 19)
	assert.NotNil(t, st.env)
	assert.Equal(t, "cancun", st.env.GetFork())
}

func TestEthTest_CreateTestTransactionWithHash(t *testing.T) {
	hash := common.HexToHash("0x1234")
	tx := CreateTestTransactionWithHash(t, hash)
	assert.NotNil(t, tx)
	st, ok := tx.(*StateTestContext)
	assert.True(t, ok)
	assert.Equal(t, hash, st.rootHash)
}

func TestEthTest_CreateTransactionWithInvalidTxBytes(t *testing.T) {
	tx := CreateTransactionWithInvalidTxBytes(t)
	assert.NotNil(t, tx)
	st, ok := tx.(*StateTestContext)
	assert.True(t, ok)
	assert.NotNil(t, st.txBytes)
	assert.NotNil(t, st.msg)
	assert.NotNil(t, st.env)
	assert.Equal(t, "cancun", st.env.GetFork())
}

func TestEthTest_CreateTransactionThatFailsSenderValidation(t *testing.T) {
	tx := CreateTransactionThatFailsSenderValidation(t)
	assert.NotNil(t, tx)
	st, ok := tx.(*StateTestContext)
	assert.True(t, ok)
	assert.NotNil(t, st.txBytes)
	assert.NotNil(t, st.msg)
	assert.NotNil(t, st.env)
	assert.Equal(t, "Shanghai", st.env.GetFork())
}
