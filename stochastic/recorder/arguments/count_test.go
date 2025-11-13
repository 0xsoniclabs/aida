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

package arguments

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestArgCountSimple counts a single occurrence of an argument and checks whether it exists.
func TestArgCountSimple(t *testing.T) {
	stats := newCount[int]()
	data := 100
	if stats.exists(data) {
		t.Fatalf("Existence check failed")
	}
	stats.place(data)
	if !stats.exists(data) {
		t.Fatalf("Existence check failed")
	}
	if stats.freq[data] != 1 {
		t.Fatalf("Counting frequency failed; expected one, is %v", stats.freq[data])
	}
}

// TestArgCountSimple2 counts two occurrences of an argument and checks whether their
// frequency is two.
func TestArgCountSimple2(t *testing.T) {
	stats := newCount[int]()
	data := 200
	if stats.exists(data) {
		t.Fatalf("Existence check failed")
	}
	stats.place(data)
	stats.place(data)
	if !stats.exists(data) {
		t.Fatalf("Existence check failed")
	}
	if stats.freq[data] != 2 {
		t.Fatalf("Counting frequency failed; expected two; is %v", stats.freq[data])
	}
}

// TestArgCountSimple3 counts the single occurrence of two items and checks whether
// their frequencies are one and whether they exist.
func TestArgCountSimple3(t *testing.T) {
	stats := newCount[int]()
	data1 := 10
	data2 := 11
	if stats.exists(data1) {
		t.Fatalf("Existence check failed")
	}
	if stats.exists(data2) {
		t.Fatalf("Existence check failed")
	}
	stats.place(data1)
	stats.place(data2)
	if !stats.exists(data1) || !stats.exists(data2) {
		t.Fatalf("Existence check failed")
	}
	if stats.freq[data1] != 1 {
		t.Fatalf("Counting frequency failed; expected one, is %v", stats.freq[data1])
	}
	if stats.freq[data2] != 1 {
		t.Fatalf("Counting frequency failed; expected one, is %v", stats.freq[data2])
	}
}

// testArgStatJSON tests the JSON output for an argument counting statistics.
// It marshals the JSON output and unmarshals it again and checks whether
// the original and unmarshaled JSON output are identical.
func testArgStatJSON(stats count[int], t *testing.T) {
	jsonX, err := stats.json()
	assert.NoError(t, err)
	jOut, err := json.Marshal(jsonX)
	if err != nil {
		t.Fatalf("Marshalling failed to produce distribution")
	}
	var jsonY ArgStatsJSON
	if err := json.Unmarshal(jOut, &jsonY); err != nil {
		t.Fatalf("Unmarshalling failed to produce JSON")
	}
	if !reflect.DeepEqual(jsonX, jsonY) {
		t.Errorf("Unmarshaling mismatch. Expected:\n%+v\nActual:\n%+v", jsonX, jsonY)
	}
}

// TestArgCountJSON tests JSON output of distribution.
// It tests an empty counting statistics and a populated one.
func TestArgCountJSON(t *testing.T) {
	stats := newCount[int]()

	// test an empty counting statistics
	testArgStatJSON(stats, t)

	// test a populate counting statistics
	for i := 1; i <= 10; i++ {
		stats.place(i)
	}
	stats.place(1)
	stats.place(10)
	testArgStatJSON(stats, t)
}

func TestArgCountMarshalJSON(t *testing.T) {
	stats := newCount[int]()
	stats.place(1)
	stats.place(2)
	payload, err := stats.MarshalJSON()
	if err != nil {
		t.Fatalf("MarshalJSON failed: %v", err)
	}
	var decoded ArgStatsJSON
	if err := json.Unmarshal(payload, &decoded); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}
	value, err := stats.json()
	assert.NoError(t, err)
	assert.True(t, reflect.DeepEqual(decoded, value))
}
