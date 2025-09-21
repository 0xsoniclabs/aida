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
	"math"
	"testing"

	"github.com/0xsoniclabs/aida/stochastic"
	"go.uber.org/mock/gomock"
)

// TestReusableNew tests the creation of a new argument set
func TestReusableNew(t *testing.T) {
	mockCtl := gomock.NewController(t)
	defer mockCtl.Finish()
	mockRandomizer := NewMockRandomizer(mockCtl)
	n := int64(1000)
	mockRandomizer.EXPECT().SampleArg(n - 1).Return(int64(0)).Times(stochastic.QueueLen)
	as := NewReusable(n, mockRandomizer)
	if as == nil {
		t.Errorf("Expected an argument set, but got nil")
	}
	n = int64(minCardinality - 1)
	as = NewReusable(n, mockRandomizer)
	if as != nil {
		t.Errorf("Expected an error, but got an argument set")
	}
}

// TestReusableChooseNoArg tests no argument kind in the Choose function of an argument set
func TestReusableChooseNoArg(t *testing.T) {
	mockCtl := gomock.NewController(t)
	defer mockCtl.Finish()
	mockRandomizer := NewMockRandomizer(mockCtl)
	n := int64(1000)
	mockRandomizer.EXPECT().SampleArg(n - 1).Return(int64(0)).Times(stochastic.QueueLen)
	as := NewReusable(n, mockRandomizer)
	_, err := as.Choose(stochastic.NoArgID)
	if err == nil {
		t.Errorf("Expected an error for NoArgID, but got nil")
	}
}

// TestReusableChooseZeroARg tests zero argument kind in the Choose function of an argument set
func TestReusableChooseZeroArg(t *testing.T) {
	mockCtl := gomock.NewController(t)
	defer mockCtl.Finish()
	mockRandomizer := NewMockRandomizer(mockCtl)
	n := int64(1000)
	mockRandomizer.EXPECT().SampleArg(n - 1).Return(int64(0)).Times(stochastic.QueueLen)
	as := NewReusable(n, mockRandomizer)
	zero, err := as.Choose(stochastic.ZeroArgID)
	if err != nil {
		t.Errorf("Expected no error for ZeroArgID got nil")
	}
	if zero != 0 {
		t.Errorf("Expected value 0 for ZeroArgID, but got %d", zero)
	}
}

// TestReusableChooseRandArg tests random argument kind in the Choose function of an argument set
func TestReusableChooseRandArg(t *testing.T) {
	mockCtl := gomock.NewController(t)
	defer mockCtl.Finish()
	mockRandomizer := NewMockRandomizer(mockCtl)
	n := int64(1000)
	// needed to fill the queue
	mockRandomizer.EXPECT().SampleArg(n - 1).Return(int64(0)).Times(stochastic.QueueLen)
	as := NewReusable(n, mockRandomizer)
	mockRandomizer.EXPECT().SampleArg(n - 1).Return(int64(4711)).Times(1)
	val, err := as.Choose(stochastic.RandArgID)
	if err != nil {
		t.Errorf("Unexpected error for RandArgID: %v", err)
	}
	if val != 4712 {
		t.Errorf("Expected value 4711 for RandArgID, but got %d", val)
	}
}

// TestReusableChoosePrevArg tests previous argument kind in the Choose function of an argument set
func TestReusableChoosePrevArg(t *testing.T) {
	mockCtl := gomock.NewController(t)
	defer mockCtl.Finish()
	mockRandomizer := NewMockRandomizer(mockCtl)
	n := int64(1000)
	mockRandomizer.EXPECT().SampleArg(n - 1).Return(int64(0)).Times(stochastic.QueueLen)
	as := NewReusable(n, mockRandomizer)
	mockRandomizer.EXPECT().SampleArg(n - 1).Return(int64(4711)).Times(1)
	val, err := as.Choose(stochastic.RandArgID)
	if err != nil {
		t.Errorf("Unexpected error for RandArgID: %v", err)
	}
	if val != 4712 {
		t.Errorf("Expected value 4711 for RandArgID, but got %d", val)
	}
	prev_val, err := as.Choose(stochastic.PrevArgID)
	if err != nil {
		t.Errorf("Unexpected error for PrevArgID: %v", err)
	}
	if prev_val != 4712 {
		t.Errorf("Expected value 501 for PrevArgID, but got %d", prev_val)
	}
}

// TestReusableChooseRecentArg tests recent argument kind in the Choose function of an argument set
func TestReusableChooseRecentArg(t *testing.T) {
	mockCtl := gomock.NewController(t)
	defer mockCtl.Finish()
	mockRandomizer := NewMockRandomizer(mockCtl)
	n := int64(1000)
	var calls []any
	calls = append(calls, mockRandomizer.EXPECT().SampleArg(n-1).Return(int64(4711)).Times(stochastic.QueueLen))
	for i := range stochastic.QueueLen - 1 {
		calls = append(calls, mockRandomizer.EXPECT().SampleQueue().Return(int(i+1)).Times(1))
	}
	gomock.InOrder(calls...)
	as := NewReusable(n, mockRandomizer)
	for range stochastic.QueueLen - 1 {
		val, err := as.Choose(stochastic.RecentArgID)
		if err != nil {
			t.Errorf("Unexpected error for RandArgID: %v", err)
		}
		expected_val := int64(4712)
		if val != expected_val {
			t.Errorf("Expected value %d for RandArgID, but got %d", expected_val, val)
		}
	}
}

// TestReusableRemove tests the Remove function of an argument set
func TestReusableRemove(t *testing.T) {
	mockCtl := gomock.NewController(t)
	defer mockCtl.Finish()
	mockRandomizer := NewMockRandomizer(mockCtl)
	n := int64(minCardinality + 10)
	mockRandomizer.EXPECT().SampleArg(n - 1).Return(int64(48)).Times(stochastic.QueueLen)
	as := NewReusable(n, mockRandomizer)
	mockRandomizer.EXPECT().SampleArg(n - 2).Return(int64(48)).Times(1)
	err := as.Remove(5)
	if err != nil {
		t.Errorf("Unexpected error when removing a valid argument: %v", err)
	}
	if as.n != minCardinality+9 {
		t.Errorf("Expected cardinality to be %d after removing an argument, but got %d", minCardinality+9, as.n)
	}
	err = as.Remove(minCardinality + 10)
	if err == nil {
		t.Errorf("Expected an error when removing an argument that is out of range, but got nil")
	}
	as.n = minCardinality
	err = as.Remove(5)
	if err == nil {
		t.Errorf("Expected an error when removing an argument that would make the cardinality too low, but got nil")
	}
}

// TestReusableChooseNewArgExceedsRange tests the error path in Choose function when new argument exceeds range
func TestReusableChooseNewArgExceedsRange(t *testing.T) {
	mockCtl := gomock.NewController(t)
	defer mockCtl.Finish()
	mockRandomizer := NewMockRandomizer(mockCtl)

	n := int64(minCardinality + 1)
	mockRandomizer.EXPECT().SampleArg(n - 1).Return(int64(0)).Times(stochastic.QueueLen)
	as := NewReusable(n, mockRandomizer)

	as.n = math.MaxInt64
	if _, err := as.Choose(stochastic.NewArgID); err == nil {
		t.Errorf("expected error when new value exceeds cardinality range")
	}
}

// TestReusableChooseUnknownKind tests the error path in Choose function when an unknown argument kind is provided
func TestReusableChooseUnknownKind(t *testing.T) {
	mockCtl := gomock.NewController(t)
	defer mockCtl.Finish()
	mockRandomizer := NewMockRandomizer(mockCtl)

	n := int64(minCardinality + 1)
	mockRandomizer.EXPECT().SampleArg(n - 1).Return(int64(0)).Times(stochastic.QueueLen)
	as := NewReusable(n, mockRandomizer)

	if _, err := as.Choose(9999); err == nil {
		t.Errorf("expected error for unknown argument kind")
	}
}

// TestReusableFindQElemFalse tests the findQElem function of an argument set for a non-existing element
func TestReusableFindQElemTrue(t *testing.T) {
	mockCtl := gomock.NewController(t)
	defer mockCtl.Finish()
	mockRandomizer := NewMockRandomizer(mockCtl)

	n := int64(minCardinality + 1)
	mockRandomizer.EXPECT().SampleArg(n - 1).Return(int64(0)).Times(stochastic.QueueLen)
	as := NewReusable(n, mockRandomizer)

	elem := as.queue[0]
	if !as.findQElem(elem) {
		t.Errorf("expected to find existing element in queue")
	}
}

// TestReusableRemoveQueueReplacement tests that removing an argument replaces it in the queue
func TestReusableRemoveQueueReplacement(t *testing.T) {
	mockCtl := gomock.NewController(t)
	defer mockCtl.Finish()
	mockRandomizer := NewMockRandomizer(mockCtl)

	n := int64(minCardinality + 10)
	mockRandomizer.EXPECT().SampleArg(n - 1).Return(int64(n - 2)).Times(stochastic.QueueLen)
	as := NewReusable(n, mockRandomizer)
	mockRandomizer.EXPECT().SampleArg(n - 2).Return(int64(0)).Times(1)

	if err := as.Remove(1); err != nil {
		t.Fatalf("unexpected error removing valid argument: %v", err)
	}

	found := false
	for i := range stochastic.QueueLen {
		if as.queue[i] == 1 {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected at least one queue element to be replaced with j==1")
	}
}

// TestReusableChooseRecentArgError covers error path in recentQ via invalid queue index.
func TestReusableChooseRecentArgError(t *testing.T) {
	mockCtl := gomock.NewController(t)
	defer mockCtl.Finish()
	mockRandomizer := NewMockRandomizer(mockCtl)

	n := int64(minCardinality + 1)
	mockRandomizer.EXPECT().SampleArg(n - 1).Return(int64(0)).Times(stochastic.QueueLen)
	as := NewReusable(n, mockRandomizer)

	mockRandomizer.EXPECT().SampleQueue().Return(0).Times(1)
	if _, err := as.Choose(stochastic.RecentArgID); err == nil {
		t.Errorf("expected error for invalid recent queue index")
	}
}
