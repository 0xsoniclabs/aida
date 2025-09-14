// Copyright 2024 Fantom Foundation
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

package continuous_empirical

import (
	"math"
	"testing"
)

func almostEqual(a, b float64) bool {
	const eps = 1e-12
	return math.Abs(a-b) <= eps
}

func TestCDF_PiecewiseInterpolationAndBoundaries(t *testing.T) {
	f := [][2]float64{
		{0.0, 0.0},
		{0.25, 0.1},
		{0.6, 0.7},
		{1.0, 1.0},
	}
	if v := CDF(f, 0.0); !almostEqual(v, 0.0) {
		t.Fatalf("CDF at x=0.0: want 0.0, got %g", v)
	}
	if v := CDF(f, 0.125); !almostEqual(v, 0.05) {
		t.Fatalf("CDF at x=0.125: want 0.05, got %g", v)
	}
	if v := CDF(f, 0.25); !almostEqual(v, 0.1) {
		t.Fatalf("CDF at x=0.25 (boundary): want 0.1, got %g", v)
	}
	if v := CDF(f, 0.40); !almostEqual(v, 0.35714285714285715) {
		t.Fatalf("CDF at x=0.40: want ~0.3571428571, got %g", v)
	}
	if v := CDF(f, 1.2); !almostEqual(v, 1.0) {
		t.Fatalf("CDF at x=1.2 (>1): want 1.0, got %g", v)
	}
}

func TestCDF_ExtrapolatesForNegativeX(t *testing.T) {
	f := [][2]float64{
		{0.0, 0.0},
		{0.5, 0.5},
		{1.0, 1.0},
	}
	if v := CDF(f, -0.1); !almostEqual(v, -0.1) {
		t.Fatalf("CDF at x=-0.1: want -0.1 (extrapolated), got %g", v)
	}
}

func TestCheckPiecewiseLinearCDF_Valid(t *testing.T) {
	f := [][2]float64{
		{0.0, 0.0},
		{0.2, 0.1},
		{0.8, 0.9},
		{1.0, 1.0},
	}
	if ok := CheckPiecewiseLinearCDF(f); !ok {
		t.Fatalf("expected valid CDF definition")
	}
}

func TestCheckPiecewiseLinearCDF_BadStart(t *testing.T) {
	f := [][2]float64{
		{0.001, 0.0},
		{0.5, 0.4},
		{1.0, 1.0},
	}
	if ok := CheckPiecewiseLinearCDF(f); ok {
		t.Fatalf("expected invalid CDF due to bad start point")
	}
}

func TestCheckPiecewiseLinearCDF_BadEnd(t *testing.T) {
	f := [][2]float64{
		{0.0, 0.0},
		{0.5, 0.4},
		{0.999, 0.999},
	}
	if ok := CheckPiecewiseLinearCDF(f); ok {
		t.Fatalf("expected invalid CDF due to bad end point")
	}
}

func TestCheckPiecewiseLinearCDF_NonIncreasingX(t *testing.T) {
	f := [][2]float64{
		{0.0, 0.0},
		{0.5, 0.4},
		{0.5, 0.6},
		{1.0, 1.0},
	}
	if ok := CheckPiecewiseLinearCDF(f); ok {
		t.Fatalf("expected invalid CDF due to non-increasing x")
	}
}
