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

	"github.com/stretchr/testify/assert"
)

func TestScalarSampler_FallbackToUniform(t *testing.T) {
	rg := rand.New(rand.NewSource(42))
	sampler := NewScalarSampler(rg, nil)

	for i := 0; i < 10; i++ {
		got := sampler.Sample(7)
		assert.GreaterOrEqual(t, got, int64(0))
		assert.Less(t, got, int64(7))
	}
	assert.Equal(t, int64(0), sampler.Sample(0))
}

func TestScalarSampler_UsesECDFAndClamps(t *testing.T) {
	rg := rand.New(rand.NewSource(7))
	ecdf := [][2]float64{
		{0.0, 0.0},
		{0.5, 0.5},
		{1.0, 1.0},
	}
	sampler := NewScalarSampler(rg, ecdf)

	for i := 0; i < 10; i++ {
		got := sampler.Sample(5)
		assert.GreaterOrEqual(t, got, int64(0))
		assert.Less(t, got, int64(5))
	}

	// Force uniform sampling, then swap back to the empirical distribution.
	sampler.Replace(nil)
	assert.Less(t, sampler.Sample(3), int64(3))

	sampler.Replace(ecdf)
	for i := 0; i < 10; i++ {
		got := sampler.Sample(1)
		assert.True(t, got == 0)
	}
}
