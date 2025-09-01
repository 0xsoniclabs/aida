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

package generator

import (
	"math/rand"
	"testing"

	"github.com/0xsoniclabs/aida/stochastic/statistics"
)

// containsQ checks whether an element is in the queue (ignoring the previous value).
func containsIndirectQ(slice []int64, x int64) bool {
	for i, n := range slice {
		if x == n && i != 0 {
			return true
		}
	}
	return false
}

// TestIndirectAccessSimple tests indirect access generator for indexes.
func TestIndirectAccessSimple(t *testing.T) {
	// create random generator with fixed seed value
	rg := rand.New(rand.NewSource(999))

	// create a random access index generator
	// with a zero probability distribution.
	qpdf := make([]float64, statistics.QueueLen)
	ia := NewIndirectAccess(NewRandomAccess(1000, NewExpRandomizer(rg, 5.0, qpdf)))

	// check no argument class (must be always -1)
	if _, err := ia.NextIndex(statistics.NoArgID); err == nil {
		t.Fatalf("expected an error message")
	}

	// check zero argument class (must be zero)
	if idx, err := ia.NextIndex(statistics.ZeroValueID); idx != 0 || err != nil {
		t.Fatalf("expected a zero value")
	}

	// check a new value (must be equal to the number of elements
	// in the index set and must be greater than zero).
	if idx, err := ia.NextIndex(statistics.NewValueID); idx != ia.NumElem() || idx < 1 || err != nil {
		t.Fatalf("expected a new index (%v, %v)", idx, ia.NumElem())
	}

	// run check again.
	if idx, err := ia.NextIndex(statistics.NewValueID); idx != ia.NumElem() || idx < 1 || err != nil {
		t.Fatalf("expected a new index (%v, %v)", idx, ia.NumElem())
	}
}

// TestIndirectAccessRecentAccess tests previous accesses
func TestIndirectAccessRecentAccess(t *testing.T) {
	// create random generator with fixed seed value
	rg := rand.New(rand.NewSource(999))

	// create a random access index generator
	// with a zero probability distribution.
	qpdf := make([]float64, statistics.QueueLen)
	ra := NewRandomAccess(1000, NewExpRandomizer(rg, 5.0, qpdf))
	ia := NewIndirectAccess(ra)

	// check a new value (must be equal to the number of elements
	// in the index set and must be greater than zero).
	idx1, err1 := ia.NextIndex(statistics.NewValueID)
	if idx1 != ra.n || idx1 < 1 || err1 != nil {
		t.Fatalf("expected a new index")
	}
	idx2, err2 := ia.NextIndex(statistics.PreviousValueID)
	if idx1 != idx2 || err2 != nil {
		t.Fatalf("previous index access failed. (%v, %v)", idx1, idx2)
	}
	idx3, err3 := ia.NextIndex(statistics.PreviousValueID)
	if idx2 != idx3 || err3 != nil {
		t.Fatalf("previous index access failed.")
	}
	// in the index set and must be greater than zero).
	idx4, err4 := ia.NextIndex(statistics.NewValueID)
	if idx4 != ra.n || idx4 < 1 || err4 != nil {
		t.Fatalf("expected a new index")
	}
	idx5, err5 := ia.NextIndex(statistics.PreviousValueID)
	if idx5 == idx3 || err5 != nil {
		t.Fatalf("previous previous index access must not be identical.")
	}
}

// TestIndirectAccessDeleteIndex tests deletion of an index
func TestIndirectAcessDeleteIndex(t *testing.T) {
	// create random generator with fixed seed value
	rg := rand.New(rand.NewSource(999))

	// create a random access index generator
	// with a zero probability distribution.
	qpdf := make([]float64, statistics.QueueLen)
	ra := NewRandomAccess(1000, NewExpRandomizer(rg, 5.0, qpdf))
	ia := NewIndirectAccess(ra)
	idx := int64(500) // choose an index in the middle of the range

	// delete previous element
	err := ia.DeleteIndex(idx)
	if err != nil {
		t.Fatalf("Deletion failed (%v).", err)
	}

	// check whether index still exists
	for i := int64(0); i < ia.NumElem(); i++ {
		if ia.translation[i] == idx {
			t.Fatalf("index still exists.")
		}
	}
}
