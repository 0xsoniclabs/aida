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
	"encoding/json"
	"math/big"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBigInt_MarshalJSON(t *testing.T) {
	b := newBigInt(1234567890)
	data, err := b.MarshalJSON()
	assert.NoError(t, err)
	var s string
	assert.NoError(t, json.Unmarshal(data, &s))
	assert.Equal(t, "0x499602d2", s)
}

func TestBigInt_UnmarshalJSON(t *testing.T) {
	var b BigInt
	jsonStr := `"0x499602d2"`
	assert.NoError(t, b.UnmarshalJSON([]byte(jsonStr)))
	exp := big.NewInt(1234567890)
	assert.Equal(t, 0, b.Cmp(exp))
}

func TestBigInt_Convert(t *testing.T) {
	b := newBigInt(42)
	bi := b.Convert()
	assert.Equal(t, 0, bi.Cmp(big.NewInt(42)))
	var nilB *BigInt
	bi2 := nilB.Convert()
	assert.Equal(t, 0, bi2.Cmp(big.NewInt(0)))
}

func TestBigInt_JSONRoundTrip(t *testing.T) {
	b := newBigInt(9876543210)
	data, err := json.Marshal(b)
	assert.NoError(t, err)
	var b2 BigInt
	assert.NoError(t, json.Unmarshal(data, &b2))
	assert.True(t, reflect.DeepEqual(b.Convert(), b2.Convert()))
}
