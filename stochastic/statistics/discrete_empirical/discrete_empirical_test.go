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

package discrete_empiricial

import (
	"math"
	"testing"
)

func TestSample_BasicCDFSteps(t *testing.T) {
	pdf := []float64{0.2, 0.3, 0.5}
	if got := Sample(pdf, 0.0); got != 0 {
		t.Fatalf("u=0.0: want 0, got %d", got)
	}
	if got := Sample(pdf, 0.2); got != 0 {
		t.Fatalf("u=0.2 (boundary): want 0, got %d", got)
	}
	if got := Sample(pdf, 0.4); got != 1 {
		t.Fatalf("u=0.4: want 1, got %d", got)
	}
	if got := Sample(pdf, 0.8); got != 2 {
		t.Fatalf("u=0.8: want 2, got %d", got)
	}
}

func TestSample_ReturnsLastPositiveWhenUSurpassesTotal(t *testing.T) {
	pdf := []float64{0.1, 0.0, 0.2}
	if got := Sample(pdf, 0.999); got != 2 {
		t.Fatalf("u>sum: want last positive index 2, got %d", got)
	}
	pdf2 := []float64{0.0, 0.7, 0.0}
	if got := Sample(pdf2, 0.9); got != 1 {
		t.Fatalf("u>sum: want last positive index 1, got %d", got)
	}
}

func TestSample_AllZerosAndEmpty(t *testing.T) {
	pdfZeros := []float64{0.0, 0.0, 0.0}
	if got := Sample(pdfZeros, 0.5); got != 0 {
		t.Fatalf("all zeros: want 0, got %d", got)
	}
	var pdfEmpty []float64
	if got := Sample(pdfEmpty, 0.3); got != 0 {
		t.Fatalf("empty pdf: want 0, got %d", got)
	}
}

func TestSample_NumericalStabilityKahanPathIsExercised(t *testing.T) {
	pdf := []float64{
		1e-16, 1e-16, 1e-16, 1e-16,
		0.25, 0.25, 0.25, 0.25,
	}
	if got := Sample(pdf, 5e-16); got != 4 {
		t.Fatalf("u~tiny: want 4, got %d", got)
	}
	if got := Sample(pdf, 0.4); got != 5 {
		t.Fatalf("u=0.4: want 5, got %d", got)
	}
	if got := Sample(pdf, 1.0-math.SmallestNonzeroFloat64); got != 7 {
		t.Fatalf("u≈1: want 7, got %d", got)
	}
}

