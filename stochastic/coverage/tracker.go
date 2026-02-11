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
	"bytes"
	"fmt"
	"os"
	"sync"

	runtimeCoverage "runtime/coverage"
)

// CounterKey uniquely identifies a coverage counter within the instrumented program.
type CounterKey struct {
	Pkg  uint32
	Func uint32
	Unit uint32
}

// Delta captures the incremental coverage gained since the previous snapshot.
type Delta struct {
	NewUnits         int
	NewLines         int
	CoverageIncrease float64
	CoverageNow      float64
}

// Tracker observes runtime code coverage counters and reports incremental changes.
type Tracker struct {
	metaHash   [16]byte
	units      map[CounterKey]unitInfo
	totalUnits int

	mu              sync.Mutex
	lastSnapshot    map[CounterKey]uint32
	coveredUnits    int
	coveredLineKeys map[string]struct{}
	lastCoverage    float64
}

var (
	writeMetaFn     = runtimeCoverage.WriteMeta
	writeCountersFn = runtimeCoverage.WriteCounters
	parseMetaFn     = parseMetaFile
)

// NewTracker constructs a tracker for the currently running instrumented program.
func NewTracker() (*Tracker, error) {
	var metaBuf bytes.Buffer
	if err := writeMetaFn(&metaBuf); err != nil {
		return nil, fmt.Errorf("coverage: metadata unavailable (build without -covermode=atomic?): %w", err)
	}

	meta, err := parseMetaFn(metaBuf.Bytes())
	if err != nil {
		return nil, err
	}

	carmenUnits := filterCarmenUnits(meta.unitDetails)
	if len(carmenUnits) == 0 {
		fmt.Fprintf(os.Stderr, "coverage: no valid Carmen packages found in metadata (expected prefix %q); rebuild with -coverpkg=github.com/0xsoniclabs/carmen/go/... to enable Carmen-only filtering\n", carmenModulePrefix)
		carmenUnits = meta.unitDetails
	}

	initialCounts, err := snapshotCountersFn(meta.hash)
	if err != nil {
		return nil, err
	}
	initialCounts = filterCountersForUnits(initialCounts, carmenUnits)

	tracker := &Tracker{
		metaHash:        meta.hash,
		units:           carmenUnits,
		totalUnits:      len(carmenUnits),
		lastSnapshot:    initialCounts,
		coveredLineKeys: make(map[string]struct{}),
	}

	for key, value := range initialCounts {
		if value > 0 {
			tracker.coveredUnits++
			for _, lineKey := range meta.unitDetails[key].lineKeys {
				tracker.coveredLineKeys[lineKey] = struct{}{}
			}
		}
	}
	if tracker.totalUnits > 0 {
		tracker.lastCoverage = float64(tracker.coveredUnits) / float64(tracker.totalUnits)
	}

	return tracker, nil
}

// Snapshot captures the latest coverage counters and returns the delta against the previous snapshot.
func (t *Tracker) Snapshot() (Delta, error) {
	counts, err := snapshotCountersFn(t.metaHash)
	if err != nil {
		return Delta{}, err
	}
	counts = filterCountersForUnits(counts, t.units)

	t.mu.Lock()
	defer t.mu.Unlock()

	return t.applySnapshot(counts), nil
}

// TotalUnits returns the total number of coverage counters recorded for the binary.
func (t *Tracker) TotalUnits() int {
	return t.totalUnits
}

func (t *Tracker) applySnapshot(current map[CounterKey]uint32) Delta {
	newUnits := 0
	newLineKeys := make(map[string]struct{})

	for key, value := range current {
		prev := t.lastSnapshot[key]
		t.lastSnapshot[key] = value

		if prev == 0 && value > 0 {
			newUnits++
			info, ok := t.units[key]
			if !ok {
				continue
			}
			for _, lineKey := range info.lineKeys {
				if lineKey == "" {
					continue
				}
				if _, seen := t.coveredLineKeys[lineKey]; !seen {
					t.coveredLineKeys[lineKey] = struct{}{}
					newLineKeys[lineKey] = struct{}{}
				}
			}
		}
	}

	t.coveredUnits += newUnits

	coverageNow := t.lastCoverage
	if t.totalUnits > 0 {
		coverageNow = float64(t.coveredUnits) / float64(t.totalUnits)
	}
	delta := Delta{
		NewUnits:         newUnits,
		NewLines:         len(newLineKeys),
		CoverageIncrease: coverageNow - t.lastCoverage,
		CoverageNow:      coverageNow,
	}
	t.lastCoverage = coverageNow

	return delta
}

func snapshotCounters(expectedHash [16]byte) (map[CounterKey]uint32, error) {
	var buf bytes.Buffer
	if err := writeCountersFn(&buf); err != nil {
		return nil, fmt.Errorf("coverage: counters unavailable (build without -covermode=atomic?): %w", err)
	}
	return parseCounterFile(expectedHash, buf.Bytes())
}

// snapshotCountersFn allows tests to stub counter collection while leaving production code unchanged.
var snapshotCountersFn = snapshotCounters
