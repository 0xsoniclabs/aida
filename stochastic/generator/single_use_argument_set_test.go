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

// TestSingleUseArgumentSetRemoveArgument tests deletion of an argument
func TestSingleUseArgSetRemoveArgument(t *testing.T) {
	mockCtl := gomock.NewController(t)
	defer mockCtl.Finish()
	mockArgSetRandomizer := NewMockArgSetRandomizer(mockCtl)
	n := ArgumentType(1000)
	// needed to fill the queue
	mockArgSetRandomizer.EXPECT().SampleDistribution(n - 1).Return(ArgumentType(0)).Times(statistics.QueueLen)
	ra := NewReusableArgumentSet(n, mockArgSetRandomizer)
	ia := NewSingleUseArgumentSet(ra)
	idx := int64(500) // choose an argument in the middle of the range

	// Remove previous element
	// expect a randomizer call during removal to refresh queue entries
	mockArgSetRandomizer.EXPECT().SampleDistribution(n - 2).Return(ArgumentType(48)).Times(1)
	err := ia.Remove(idx)
	if err != nil {
		t.Fatalf("Deletion failed (%v).", err)
	}

	// check whether argument still exists
	for i := int64(0); i < ia.Size(); i++ {
		if ia.translation[i] == idx {
			t.Fatalf("argument still exists.")
		}
	}
}

// TestSingleUseChoosePropagatesUnderlyingError tests error propagation from underlying Choose.
func TestSingleUseChoosePropagatesUnderlyingError(t *testing.T) {
	mockCtl := gomock.NewController(t)
	defer mockCtl.Finish()
	mockArgSet := NewMockArgumentSet(mockCtl)
	mockArgSet.EXPECT().Size().Return(ArgumentType(5)).AnyTimes()
	ia := NewSingleUseArgumentSet(mockArgSet)
	chooseErr := errors.New("choose failed")
	mockArgSet.EXPECT().Choose(statistics.PrevArgID).Return(ArgumentType(0), chooseErr)
	if _, err := ia.Choose(statistics.PrevArgID); err == nil {
		t.Fatalf("expected error to propagate from underlying Choose")
	}
}

// TestSingleUseChooseTranslationargumentOutOfRangeLow tests translation argument out of range (low).
func TestSingleUseChooseTranslationargumentOutOfRangeLow(t *testing.T) {
	mockCtl := gomock.NewController(t)
	defer mockCtl.Finish()
	mockArgSet := NewMockArgumentSet(mockCtl)
	mockArgSet.EXPECT().Size().Return(ArgumentType(5)).AnyTimes()
	ia := NewSingleUseArgumentSet(mockArgSet)
	mockArgSet.EXPECT().Choose(statistics.PrevArgID).Return(ArgumentType(0), nil)
	if _, err := ia.Choose(statistics.PrevArgID); err == nil {
		t.Fatalf("expected translation argument out of range error for v<=0")
	}
}

// TestSingleUseChooseTranslationargumentOutOfRangeHigh tests out-of-range error for v>len(translation).
func TestSingleUseChooseTranslationargumentOutOfRangeHigh(t *testing.T) {
	mockCtl := gomock.NewController(t)
	defer mockCtl.Finish()
	mockArgSet := NewMockArgumentSet(mockCtl)
	mockArgSet.EXPECT().Size().Return(ArgumentType(5)).AnyTimes()
	ia := NewSingleUseArgumentSet(mockArgSet)
	mockArgSet.EXPECT().Choose(statistics.PrevArgID).Return(ArgumentType(6), nil)
	if _, err := ia.Choose(statistics.PrevArgID); err == nil {
		t.Fatalf("expected translation argument out of range error for v>len(translation)")
	}
}

func TestSingleUseChooseDefaultReturnsTranslatedValue(t *testing.T) {
	mockCtl := gomock.NewController(t)
	defer mockCtl.Finish()
	mockArgSet := NewMockArgumentSet(mockCtl)
	mockArgSet.EXPECT().Size().Return(ArgumentType(5)).AnyTimes()
	ia := NewSingleUseArgumentSet(mockArgSet)
	mockArgSet.EXPECT().Choose(statistics.PrevArgID).Return(ArgumentType(2), nil)
	v, err := ia.Choose(statistics.PrevArgID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v != 2 {
		t.Fatalf("expected translated value 2, got %d", v)
	}
}

// TestSingleUseRemoveZeroIsNoop tests that removing argument 0 is a no-op.
func TestSingleUseRemoveZeroIsNoop(t *testing.T) {
	mockCtl := gomock.NewController(t)
	defer mockCtl.Finish()
	mockArgSet := NewMockArgumentSet(mockCtl)
	mockArgSet.EXPECT().Size().Return(ArgumentType(5)).AnyTimes()
	ia := NewSingleUseArgumentSet(mockArgSet)
	if err := ia.Remove(0); err != nil {
		t.Fatalf("expected nil error for k==0, got %v", err)
	}
}

// TestSingleUseRemoveargumentNotFound tests removal of a non-existing argument.
func TestSingleUseRemoveargumentNotFound(t *testing.T) {
	mockCtl := gomock.NewController(t)
	defer mockCtl.Finish()
	mockArgSet := NewMockArgumentSet(mockCtl)
	mockArgSet.EXPECT().Size().Return(ArgumentType(5)).AnyTimes()
	ia := NewSingleUseArgumentSet(mockArgSet)
	if err := ia.Remove(ArgumentType(999)); err == nil {
		t.Fatalf("expected error when removing non-existing argument")
	}
}

// TestSingleUseRemovePropagatesUnderlyingError tests error propagation from underlying Remove.
func TestSingleUseRemovePropagatesUnderlyingError(t *testing.T) {
	mockCtl := gomock.NewController(t)
	defer mockCtl.Finish()
	mockArgSet := NewMockArgumentSet(mockCtl)
	mockArgSet.EXPECT().Size().Return(ArgumentType(5)).AnyTimes()
	ia := NewSingleUseArgumentSet(mockArgSet)
	remErr := errors.New("remove failed")
	mockArgSet.EXPECT().Remove(ArgumentType(3)).Return(remErr)
	if err := ia.Remove(ArgumentType(3)); err == nil {
		t.Fatalf("expected error propagation from underlying Remove")
	}
}

// TestSingleUseRemoveSuccess tests successful removal of an argument.
func TestSingleUseRemoveSuccess(t *testing.T) {
	mockCtl := gomock.NewController(t)
	defer mockCtl.Finish()
	mockArgSet := NewMockArgumentSet(mockCtl)
	mockArgSet.EXPECT().Size().Return(ArgumentType(5)).AnyTimes()
	ia := NewSingleUseArgumentSet(mockArgSet)
	mockArgSet.EXPECT().Remove(ArgumentType(1)).Return(nil)
	if err := ia.Remove(ArgumentType(1)); err != nil {
		t.Fatalf("expected nil error on successful remove, got %v", err)
	}
}

// TestSingleUseArgumentSetSimple tests basic functionality of SingleUseArgumentSet
func TestSingleUseArgumentSetSimple(t *testing.T) {
	mockCtl := gomock.NewController(t)
	defer mockCtl.Finish()
	mockArgSetRandomizer := NewMockArgSetRandomizer(mockCtl)
	n := ArgumentType(1000)
	mockArgSetRandomizer.EXPECT().SampleDistribution(n - 1).Return(ArgumentType(0)).Times(statistics.QueueLen)
	ia := NewSingleUseArgumentSet(NewReusableArgumentSet(n, mockArgSetRandomizer))
	if _, err := ia.Choose(statistics.NoArgID); err == nil {
		t.Fatalf("expected an error message")
	}

	// check zero argument class (must be zero)
	if idx, err := ia.Choose(statistics.ZeroArgID); idx != 0 || err != nil {
		t.Fatalf("expected a zero value")
	}

	// check a new value (must be equal to the number of elements
	// in the argument set and must be greater than zero).
	if idx, err := ia.Choose(statistics.NewArgID); idx != ia.Size() || idx < 1 || err != nil {
		t.Fatalf("expected a new argument (%v, %v)", idx, ia.Size())
	}

	// run check again.
	if idx, err := ia.Choose(statistics.NewArgID); idx != ia.Size() || idx < 1 || err != nil {
		t.Fatalf("expected a new argument (%v, %v)", idx, ia.Size())
	}
}

// TestSingleUseArgumentSetRecentAccess tests previous accesses
func TestSingleUseArgumentSetRecentAccess(t *testing.T) {
	mockCtl := gomock.NewController(t)
	defer mockCtl.Finish()
	mockArgSetRandomizer := NewMockArgSetRandomizer(mockCtl)
	n := ArgumentType(1000)
	// needed to fill the queue
	mockArgSetRandomizer.EXPECT().SampleDistribution(n - 1).Return(ArgumentType(0)).Times(statistics.QueueLen)
	ra := NewReusableArgumentSet(n, mockArgSetRandomizer)
	ia := NewSingleUseArgumentSet(ra)

	// check a new value (must be equal to the number of elements
	// in the argument set and must be greater than zero).
	idx1, err1 := ia.Choose(statistics.NewArgID)
	if idx1 != ra.n || idx1 < 1 || err1 != nil {
		t.Fatalf("expected a new argument")
	}
	idx2, err2 := ia.Choose(statistics.PrevArgID)
	if idx1 != idx2 || err2 != nil {
		t.Fatalf("previous argument access failed. (%v, %v)", idx1, idx2)
	}
	idx3, err3 := ia.Choose(statistics.PrevArgID)
	if idx2 != idx3 || err3 != nil {
		t.Fatalf("previous argument access failed.")
	}
	// in the argument set and must be greater than zero).
	idx4, err4 := ia.Choose(statistics.NewArgID)
	if idx4 != ra.n || idx4 < 1 || err4 != nil {
		t.Fatalf("expected a new argument")
	}
	idx5, err5 := ia.Choose(statistics.PrevArgID)
	if idx5 == idx3 || err5 != nil {
		t.Fatalf("previous previous argument access must not be identical.")
	}
}
