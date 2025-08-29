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

package rpc

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// Package-level function tests
func TestRpc_NewResult(t *testing.T) {
	res := []byte("test result")
	err := error(nil)
	gasUsed := uint64(100)

	out := NewResult(res, err, gasUsed)

	// cast back
	r, ok := out.(*result)
	assert.True(t, ok)
	assert.NotNil(t, out)
	assert.Equal(t, err, r.err)
	assert.Equal(t, res, r.result)
	assert.Equal(t, gasUsed, r.gasUsed)
}

// Result struct method tests
func TestResult_GetReceipt(t *testing.T) {
	res := []byte("test result")
	err := error(nil)
	gasUsed := uint64(100)

	r := &result{
		gasUsed: gasUsed,
		result:  res,
		err:     err,
	}
	out := r.GetReceipt()
	assert.Nil(t, out)
}

func TestResult_GetRawResult(t *testing.T) {
	res := []byte("test result")
	mockErr := error(nil)
	gasUsed := uint64(100)

	r := &result{
		gasUsed: gasUsed,
		result:  res,
		err:     mockErr,
	}
	out, err := r.GetRawResult()
	assert.Equal(t, mockErr, err)
	assert.Equal(t, res, out)
}

func TestResult_GetGasUsed(t *testing.T) {
	res := []byte("test result")
	err := error(nil)
	gasUsed := uint64(100)

	r := &result{
		gasUsed: gasUsed,
		result:  res,
		err:     err,
	}
	out := r.GetGasUsed()
	assert.Equal(t, gasUsed, out)
}

func TestResult_String(t *testing.T) {
	res := []byte("test result")
	err := error(nil)
	gasUsed := uint64(100)

	r := &result{
		gasUsed: gasUsed,
		result:  res,
		err:     err,
	}
	out := r.String()
	assert.Contains(t, out, "Result: test result")
}
