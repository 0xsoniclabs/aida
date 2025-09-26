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
	"math/rand"
	"testing"

	"github.com/0xsoniclabs/aida/stochastic"
)

// TestRandomizer_FailNewRandomizer tests the failing NewRandomizer
func TestRandomizer_FailNewRandomizer(t *testing.T) {
	rg := rand.New(rand.NewSource(1))
	queuePMF := make([]float64, 1)
	argCDF := [][2]float64{{0.0, 0.0}, {1.0, 1.0}}
	_, err := NewRandomizer(rg, queuePMF, argCDF)
	if err == nil {
		t.Fatalf("expected to fail")
	}
}

// TestRandomizer_Simple tests the NewRandomizer, SampleArg, and SampleQueue functions.
func TestRandomizer_Simple(t *testing.T) {
	rg := rand.New(rand.NewSource(1))
	queuePMF := make([]float64, stochastic.QueueLen)
	x := 1.0 / float64(stochastic.QueueLen)
	for i := range stochastic.QueueLen {
		queuePMF[i] = x
	}
	n := int64(100)
	argCDF := [][2]float64{{0.0, 0.0}, {1.0, 1.0}}
	r, err := NewRandomizer(rg, queuePMF, argCDF)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if r == nil {
		t.Fatalf("unexpected nil Randomizer")
	}
	for range 10000 {
		v := r.SampleArg(n)
		if v < 0 || v >= n {
			t.Fatalf("sampled argument value out of range: %d", v)
		}
	}
	for range 10000 {
		if v := r.SampleQueue(); v < 1 || v >= stochastic.QueueLen {
			t.Fatalf("sampled queue value out of range [1,%d): %d", stochastic.QueueLen, v)
		}
	}
}

// TestRandomizer_SampleQueueRange ensures SampleQueue stays within [1,QueueLen-1].
func TestRandomizer_SampleQueueRange(t *testing.T) {
	rg := rand.New(rand.NewSource(1337))

	// Valid queuePMF: pdf[0] in (0,1), others positive and <1; shape doesn't matter for range.
	queuePMF := make([]float64, stochastic.QueueLen)
	queuePMF[0] = 0.1
	rest := 0.9 / float64(stochastic.QueueLen-1)
	for i := 1; i < len(queuePMF); i++ {
		queuePMF[i] = rest
	}

	// Simple argCDF; not used by SampleQueue but required by constructor.
	argCDF := [][2]float64{{0.0, 0.0}, {1.0, 1.0}}

	r, err := NewRandomizer(rg, queuePMF, argCDF)
	if err != nil {
		t.Fatalf("unexpected error constructing Randomizer: %v", err)
	}

	for range 1000 {
		v := r.SampleQueue()
		if v < 1 || v >= stochastic.QueueLen {
			t.Fatalf("queue index out of range: %d", v)
		}
	}
}
