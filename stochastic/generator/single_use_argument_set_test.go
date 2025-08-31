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
	"errors"
	"testing"

	"github.com/0xsoniclabs/aida/stochastic/statistics"
	"github.com/golang/mock/gomock"
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

// TestSingleUseArgumentSetSimple tests indirect access generator for indexes.
func TestSingleUseArgumentSetSimple(t *testing.T) {
	mockCtl := gomock.NewController(t)
	defer mockCtl.Finish()
	mockRandomizer := NewMockRandomizer(mockCtl)
	n := ArgumentType(1000)
	// needed to fill the queue
	mockRandomizer.EXPECT().SampleDistribution(n - 1).Return(ArgumentType(0)).Times(statistics.QueueLen)
	ia := NewSingleUseArgumentSet(NewReusableArgumentSet(n, mockRandomizer))

	// check no argument class (must be always -1)
	if _, err := ia.Choose(statistics.NoArgID); err == nil {
		t.Fatalf("expected an error message")
	}

	// check zero argument class (must be zero)
	if idx, err := ia.Choose(statistics.ZeroArgID); idx != 0 || err != nil {
		t.Fatalf("expected a zero value")
	}

	// check a new value (must be equal to the number of elements
	// in the index set and must be greater than zero).
	if idx, err := ia.Choose(statistics.NewArgID); idx != ia.Size() || idx < 1 || err != nil {
		t.Fatalf("expected a new index (%v, %v)", idx, ia.Size())
	}

	// run check again.
	if idx, err := ia.Choose(statistics.NewArgID); idx != ia.Size() || idx < 1 || err != nil {
		t.Fatalf("expected a new index (%v, %v)", idx, ia.Size())
	}
}

// TestSingleUseArgumentSetRecentAccess tests previous accesses
func TestSingleUseArgumentSetRecentAccess(t *testing.T) {
	mockCtl := gomock.NewController(t)
	defer mockCtl.Finish()
	mockRandomizer := NewMockRandomizer(mockCtl)
	n := ArgumentType(1000)
	// needed to fill the queue
	mockRandomizer.EXPECT().SampleDistribution(n - 1).Return(ArgumentType(0)).Times(statistics.QueueLen)
	ra := NewReusableArgumentSet(n, mockRandomizer)
	ia := NewSingleUseArgumentSet(ra)

	// check a new value (must be equal to the number of elements
	// in the index set and must be greater than zero).
	idx1, err1 := ia.Choose(statistics.NewArgID)
	if idx1 != ra.n || idx1 < 1 || err1 != nil {
		t.Fatalf("expected a new index")
	}
	idx2, err2 := ia.Choose(statistics.PrevArgID)
	if idx1 != idx2 || err2 != nil {
		t.Fatalf("previous index access failed. (%v, %v)", idx1, idx2)
	}
	idx3, err3 := ia.Choose(statistics.PrevArgID)
	if idx2 != idx3 || err3 != nil {
		t.Fatalf("previous index access failed.")
	}
	// in the index set and must be greater than zero).
	idx4, err4 := ia.Choose(statistics.NewArgID)
	if idx4 != ra.n || idx4 < 1 || err4 != nil {
		t.Fatalf("expected a new index")
	}
	idx5, err5 := ia.Choose(statistics.PrevArgID)
	if idx5 == idx3 || err5 != nil {
		t.Fatalf("previous previous index access must not be identical.")
	}
}

// TestSingleUseArgumentSetDeleteIndex tests deletion of an index
func TestIndirectAcessDeleteIndex(t *testing.T) {
	mockCtl := gomock.NewController(t)
	defer mockCtl.Finish()
	mockRandomizer := NewMockRandomizer(mockCtl)
	n := ArgumentType(1000)
	// needed to fill the queue
	mockRandomizer.EXPECT().SampleDistribution(n - 1).Return(ArgumentType(0)).Times(statistics.QueueLen)
	ra := NewReusableArgumentSet(n, mockRandomizer)
	ia := NewSingleUseArgumentSet(ra)
	idx := int64(500) // choose an index in the middle of the range

	// delete previous element
	// expect a randomizer call during removal to refresh queue entries
	mockRandomizer.EXPECT().SampleDistribution(n - 2).Return(ArgumentType(48)).Times(1)
	err := ia.Remove(idx)
	if err != nil {
		t.Fatalf("Deletion failed (%v).", err)
	}

	// check whether index still exists
	for i := int64(0); i < ia.Size(); i++ {
		if ia.translation[i] == idx {
			t.Fatalf("index still exists.")
		}
	}
}

func TestSingleUseChoosePropagatesUnderlyingError(t *testing.T) {
    mockCtl := gomock.NewController(t)
    defer mockCtl.Finish()

    mockAS := NewMockArgumentSet(mockCtl)
    mockAS.EXPECT().Size().Return(ArgumentType(5)).AnyTimes()

    ia := NewSingleUseArgumentSet(mockAS)

    chooseErr := errors.New("choose failed")
    mockAS.EXPECT().Choose(statistics.PrevArgID).Return(ArgumentType(0), chooseErr)

    if _, err := ia.Choose(statistics.PrevArgID); err == nil {
        t.Fatalf("expected error to propagate from underlying Choose")
    }
}

func TestSingleUseChooseTranslationIndexOutOfRangeLow(t *testing.T) {
    mockCtl := gomock.NewController(t)
    defer mockCtl.Finish()

    mockAS := NewMockArgumentSet(mockCtl)
    mockAS.EXPECT().Size().Return(ArgumentType(5)).AnyTimes()

    ia := NewSingleUseArgumentSet(mockAS)

    mockAS.EXPECT().Choose(statistics.PrevArgID).Return(ArgumentType(0), nil)

    if _, err := ia.Choose(statistics.PrevArgID); err == nil {
        t.Fatalf("expected translation index out of range error for v<=0")
    }
}

func TestSingleUseChooseTranslationIndexOutOfRangeHigh(t *testing.T) {
    mockCtl := gomock.NewController(t)
    defer mockCtl.Finish()

    mockAS := NewMockArgumentSet(mockCtl)
    mockAS.EXPECT().Size().Return(ArgumentType(5)).AnyTimes()

    ia := NewSingleUseArgumentSet(mockAS)

    // len(translation) == 5, return 6 to exceed bounds safely
    mockAS.EXPECT().Choose(statistics.PrevArgID).Return(ArgumentType(6), nil)

    if _, err := ia.Choose(statistics.PrevArgID); err == nil {
        t.Fatalf("expected translation index out of range error for v>len(translation)")
    }
}

func TestSingleUseChooseDefaultReturnsTranslatedValue(t *testing.T) {
    mockCtl := gomock.NewController(t)
    defer mockCtl.Finish()

    mockAS := NewMockArgumentSet(mockCtl)
    mockAS.EXPECT().Size().Return(ArgumentType(5)).AnyTimes()

    ia := NewSingleUseArgumentSet(mockAS)

    mockAS.EXPECT().Choose(statistics.PrevArgID).Return(ArgumentType(2), nil)

    v, err := ia.Choose(statistics.PrevArgID)
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    if v != 2 {
        t.Fatalf("expected translated value 2, got %d", v)
    }
}

func TestSingleUseRemoveZeroIsNoop(t *testing.T) {
    mockCtl := gomock.NewController(t)
    defer mockCtl.Finish()

    mockAS := NewMockArgumentSet(mockCtl)
    mockAS.EXPECT().Size().Return(ArgumentType(5)).AnyTimes()

    ia := NewSingleUseArgumentSet(mockAS)

    if err := ia.Remove(0); err != nil {
        t.Fatalf("expected nil error for k==0, got %v", err)
    }
}

func TestSingleUseRemoveIndexNotFound(t *testing.T) {
    mockCtl := gomock.NewController(t)
    defer mockCtl.Finish()

    mockAS := NewMockArgumentSet(mockCtl)
    mockAS.EXPECT().Size().Return(ArgumentType(5)).AnyTimes()

    ia := NewSingleUseArgumentSet(mockAS)

    if err := ia.Remove(ArgumentType(999)); err == nil {
        t.Fatalf("expected error when removing non-existing index")
    }
}

func TestSingleUseRemovePropagatesUnderlyingError(t *testing.T) {
    mockCtl := gomock.NewController(t)
    defer mockCtl.Finish()

    mockAS := NewMockArgumentSet(mockCtl)
    mockAS.EXPECT().Size().Return(ArgumentType(5)).AnyTimes()

    ia := NewSingleUseArgumentSet(mockAS)

    // Removing existing translated index 3 should call underlying Remove(3)
    remErr := errors.New("remove failed")
    mockAS.EXPECT().Remove(ArgumentType(3)).Return(remErr)

    if err := ia.Remove(ArgumentType(3)); err == nil {
        t.Fatalf("expected error propagation from underlying Remove")
    }
}

func TestSingleUseRemoveSuccess(t *testing.T) {
    mockCtl := gomock.NewController(t)
    defer mockCtl.Finish()

    mockAS := NewMockArgumentSet(mockCtl)
    mockAS.EXPECT().Size().Return(ArgumentType(5)).AnyTimes()

    ia := NewSingleUseArgumentSet(mockAS)

    mockAS.EXPECT().Remove(ArgumentType(1)).Return(nil)

    if err := ia.Remove(ArgumentType(1)); err != nil {
        t.Fatalf("expected nil error on successful remove, got %v", err)
    }
}
