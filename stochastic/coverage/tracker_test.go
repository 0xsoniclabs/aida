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

package coverage

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTracker_ApplySnapshot(t *testing.T) {
	hash := [16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}

	tracker := &Tracker{
		metaHash:   hash,
		totalUnits: 100,
		units: map[CounterKey]unitInfo{
			{Pkg: 0, Func: 0, Unit: 0}: {
				PackagePath: "github.com/test/pkg",
				FuncName:    "TestFunc",
				File:        "test.go",
				StartLine:   10,
				EndLine:     15,
				lineKeys:    []string{"test.go:10", "test.go:11", "test.go:12", "test.go:13", "test.go:14", "test.go:15"},
			},
			{Pkg: 0, Func: 0, Unit: 1}: {
				PackagePath: "github.com/test/pkg",
				FuncName:    "TestFunc",
				File:        "test.go",
				StartLine:   20,
				EndLine:     22,
				lineKeys:    []string{"test.go:20", "test.go:21", "test.go:22"},
			},
			{Pkg: 0, Func: 1, Unit: 0}: {
				PackagePath: "github.com/test/pkg",
				FuncName:    "AnotherFunc",
				File:        "other.go",
				StartLine:   5,
				EndLine:     8,
				lineKeys:    []string{"other.go:5", "other.go:6", "other.go:7", "other.go:8"},
			},
		},
		lastSnapshot:    make(map[CounterKey]uint32),
		coveredLineKeys: make(map[string]struct{}),
	}

	// Initial snapshot - all zeros
	current := map[CounterKey]uint32{
		{Pkg: 0, Func: 0, Unit: 0}: 0,
		{Pkg: 0, Func: 0, Unit: 1}: 0,
		{Pkg: 0, Func: 1, Unit: 0}: 0,
	}

	delta := tracker.applySnapshot(current)
	require.Equal(t, 0, delta.NewUnits)
	require.Equal(t, 0, delta.NewLines)
	require.Equal(t, 0.0, delta.CoverageIncrease)
	require.Equal(t, 0.0, delta.CoverageNow)

	// Second snapshot - one unit becomes non-zero
	current = map[CounterKey]uint32{
		{Pkg: 0, Func: 0, Unit: 0}: 5, // now covered!
		{Pkg: 0, Func: 0, Unit: 1}: 0,
		{Pkg: 0, Func: 1, Unit: 0}: 0,
	}

	delta = tracker.applySnapshot(current)
	require.Equal(t, 1, delta.NewUnits)
	require.Equal(t, 6, delta.NewLines)                     // lines 10-15
	require.InDelta(t, 0.01, delta.CoverageIncrease, 0.001) // 1/100 = 0.01
	require.InDelta(t, 0.01, delta.CoverageNow, 0.001)

	// Third snapshot - two more units covered
	current = map[CounterKey]uint32{
		{Pkg: 0, Func: 0, Unit: 0}: 10, // increased
		{Pkg: 0, Func: 0, Unit: 1}: 1,  // newly covered
		{Pkg: 0, Func: 1, Unit: 0}: 3,  // newly covered
	}

	delta = tracker.applySnapshot(current)
	require.Equal(t, 2, delta.NewUnits)                     // units 1 and 2 are new
	require.Equal(t, 7, delta.NewLines)                     // 3 lines from unit 1 + 4 lines from unit 2
	require.InDelta(t, 0.02, delta.CoverageIncrease, 0.001) // went from 1/100 to 3/100
	require.InDelta(t, 0.03, delta.CoverageNow, 0.001)

	// Fourth snapshot - no new coverage (just counter increases)
	current = map[CounterKey]uint32{
		{Pkg: 0, Func: 0, Unit: 0}: 15, // just increased
		{Pkg: 0, Func: 0, Unit: 1}: 2,  // just increased
		{Pkg: 0, Func: 1, Unit: 0}: 5,  // just increased
	}

	delta = tracker.applySnapshot(current)
	require.Equal(t, 0, delta.NewUnits)
	require.Equal(t, 0, delta.NewLines)
	require.Equal(t, 0.0, delta.CoverageIncrease)
	require.InDelta(t, 0.03, delta.CoverageNow, 0.001)
}

func TestTracker_ApplySnapshot_DuplicateLines(t *testing.T) {
	// Test that duplicate line keys across units don't get counted multiple times
	hash := [16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}

	tracker := &Tracker{
		metaHash:   hash,
		totalUnits: 10,
		units: map[CounterKey]unitInfo{
			{Pkg: 0, Func: 0, Unit: 0}: {
				File:     "test.go",
				lineKeys: []string{"test.go:10", "test.go:11", "test.go:12"},
			},
			{Pkg: 0, Func: 0, Unit: 1}: {
				File:     "test.go",
				lineKeys: []string{"test.go:11", "test.go:12", "test.go:13"}, // overlap with unit 0
			},
		},
		lastSnapshot:    make(map[CounterKey]uint32),
		coveredLineKeys: make(map[string]struct{}),
	}

	// Cover unit 0
	current := map[CounterKey]uint32{
		{Pkg: 0, Func: 0, Unit: 0}: 1,
		{Pkg: 0, Func: 0, Unit: 1}: 0,
	}

	delta := tracker.applySnapshot(current)
	require.Equal(t, 1, delta.NewUnits)
	require.Equal(t, 3, delta.NewLines) // 10, 11, 12

	// Cover unit 1 - should only count line 13 as new (11 and 12 already covered)
	current = map[CounterKey]uint32{
		{Pkg: 0, Func: 0, Unit: 0}: 1,
		{Pkg: 0, Func: 0, Unit: 1}: 1,
	}

	delta = tracker.applySnapshot(current)
	require.Equal(t, 1, delta.NewUnits)
	require.Equal(t, 1, delta.NewLines) // only line 13 is new
}

func TestTracker_ApplySnapshot_EmptyLineKeys(t *testing.T) {
	// Test units with no line keys (edge case)
	tracker := &Tracker{
		totalUnits: 5,
		units: map[CounterKey]unitInfo{
			{Pkg: 0, Func: 0, Unit: 0}: {
				File:     "",
				lineKeys: nil,
			},
			{Pkg: 0, Func: 0, Unit: 1}: {
				File:     "test.go",
				lineKeys: []string{""}, // empty string
			},
		},
		lastSnapshot:    make(map[CounterKey]uint32),
		coveredLineKeys: make(map[string]struct{}),
	}

	current := map[CounterKey]uint32{
		{Pkg: 0, Func: 0, Unit: 0}: 1,
		{Pkg: 0, Func: 0, Unit: 1}: 1,
	}

	delta := tracker.applySnapshot(current)
	require.Equal(t, 2, delta.NewUnits)
	require.Equal(t, 0, delta.NewLines) // no valid line keys
}

func TestTracker_ApplySnapshot_UnknownUnit(t *testing.T) {
	// Test that unknown units in counter data don't crash
	tracker := &Tracker{
		totalUnits: 10,
		units: map[CounterKey]unitInfo{
			{Pkg: 0, Func: 0, Unit: 0}: {
				lineKeys: []string{"test.go:10"},
			},
		},
		lastSnapshot:    make(map[CounterKey]uint32),
		coveredLineKeys: make(map[string]struct{}),
	}

	current := map[CounterKey]uint32{
		{Pkg: 0, Func: 0, Unit: 0}:    1,
		{Pkg: 99, Func: 99, Unit: 99}: 1, // unknown unit
	}

	delta := tracker.applySnapshot(current)
	require.Equal(t, 2, delta.NewUnits) // both counted as new units
	require.Equal(t, 1, delta.NewLines) // only the known unit contributes lines
}

func TestTracker_TotalUnits(t *testing.T) {
	tracker := &Tracker{
		totalUnits: 42,
	}

	require.Equal(t, 42, tracker.TotalUnits())
}

func TestCounterKey(t *testing.T) {
	// Test that CounterKey can be used as map key
	m := make(map[CounterKey]int)

	key1 := CounterKey{Pkg: 1, Func: 2, Unit: 3}
	key2 := CounterKey{Pkg: 1, Func: 2, Unit: 3}
	key3 := CounterKey{Pkg: 1, Func: 2, Unit: 4}

	m[key1] = 100
	m[key2] = 200 // should overwrite key1
	m[key3] = 300

	require.Equal(t, 200, m[key1])
	require.Equal(t, 200, m[key2])
	require.Equal(t, 300, m[key3])
	require.Len(t, m, 2)
}

func TestDelta(t *testing.T) {
	// Test Delta structure
	delta := Delta{
		NewUnits:         5,
		NewLines:         10,
		CoverageIncrease: 0.05,
		CoverageNow:      0.15,
	}

	require.Equal(t, 5, delta.NewUnits)
	require.Equal(t, 10, delta.NewLines)
	require.Equal(t, 0.05, delta.CoverageIncrease)
	require.Equal(t, 0.15, delta.CoverageNow)
}
