// // Copyright 2024 Fantom Foundation
// // This file is part of Aida Testing Infrastructure for Sonic
// //
// // Aida is free software: you can redistribute it and/or modify
// // it under the terms of the GNU Lesser General Public License as published by
// // the Free Software Foundation, either version 3 of the License, or
// // (at your option) any later version.
// //
// // Aida is distributed in the hope that it will be useful,
// // but WITHOUT ANY WARRANTY; without even the implied warranty of
// // MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// // GNU Lesser General Public License for more details.
// //
// // You should have received a copy of the GNU Lesser General Public License
// // along with Aida. If not, see <http://www.gnu.org/licenses/>.

package generator

import (
	"testing"

	"github.com/0xsoniclabs/aida/stochastic/statistics"
	"github.com/golang/mock/gomock"
)

// test no argument kind in the Choose function of an argument set
func TestArgSetChooseNoArg(t *testing.T) {
	mockCtl := gomock.NewController(t)
	defer mockCtl.Finish()
	mockRandomizer := NewMockRandomizer(mockCtl)
	n := ArgumentType(1000)
	// needed to fill the queue
	mockRandomizer.EXPECT().SampleDistribution(n - 1).Return(ArgumentType(0)).Times(statistics.QueueLen)
	as := NewArgumentSet(n, mockRandomizer)
	_, err := as.Choose(statistics.NoArgID)
	if err == nil {
		t.Errorf("Expected an error for NoArgID, but got nil")
	}
}

// TestArgSetChooseZeroARg tests zero argument kind in the Choose function of an argument set
func TestArgSetChooseZeroArg(t *testing.T) {
	mockCtl := gomock.NewController(t)
	defer mockCtl.Finish()
	mockRandomizer := NewMockRandomizer(mockCtl)
	n := ArgumentType(1000)
	// needed to fill the queue
	mockRandomizer.EXPECT().SampleDistribution(n - 1).Return(ArgumentType(0)).Times(statistics.QueueLen)
	as := NewArgumentSet(n, mockRandomizer)
	zero, err := as.Choose(statistics.ZeroArgID)
	if err != nil {
		t.Errorf("Expected no error for ZeroArgID got nil")
	}
	if zero != 0 {
		t.Errorf("Expected value 0 for ZeroArgID, but got %d", zero)
	}
}

// TestArgSetChooseRandArg tests random argument kind in the Choose function of an argument set
func TestArgSetChooseRandArg(t *testing.T) {
	mockCtl := gomock.NewController(t)
	defer mockCtl.Finish()
	mockRandomizer := NewMockRandomizer(mockCtl)
	n := ArgumentType(1000)
	// needed to fill the queue
	mockRandomizer.EXPECT().SampleDistribution(n - 1).Return(ArgumentType(0)).Times(statistics.QueueLen)
	as := NewArgumentSet(n, mockRandomizer)
	mockRandomizer.EXPECT().SampleDistribution(n - 1).Return(ArgumentType(4711)).Times(1)
	val, err := as.Choose(statistics.RandArgID)
	if err != nil {
		t.Errorf("Unexpected error for RandArgID: %v", err)
	}
	if val != 4712 {
		t.Errorf("Expected value 4711 for RandArgID, but got %d", val)
	}
}

// TestArgSetChoosePrevArg tests previous argument kind in the Choose function of an argument set
func TestArgSetChoosePrevArg(t *testing.T) {
	mockCtl := gomock.NewController(t)
	defer mockCtl.Finish()
	mockRandomizer := NewMockRandomizer(mockCtl)
	n := ArgumentType(1000)
	// needed to fill the queue
	mockRandomizer.EXPECT().SampleDistribution(n - 1).Return(ArgumentType(0)).Times(statistics.QueueLen)
	as := NewArgumentSet(n, mockRandomizer)

	// load queue with a known value
	mockRandomizer.EXPECT().SampleDistribution(n - 1).Return(ArgumentType(4711)).Times(1)
	val, err := as.Choose(statistics.RandArgID)
	if err != nil {
		t.Errorf("Unexpected error for RandArgID: %v", err)
	}
	if val != 4712 {
		t.Errorf("Expected value 4711 for RandArgID, but got %d", val)
	}

	// check whether the queue returns the expected previous value
	prev_val, err := as.Choose(statistics.PrevArgID)
	if err != nil {
		t.Errorf("Unexpected error for PrevArgID: %v", err)
	}
	if prev_val != 4712 {
		t.Errorf("Expected value 501 for PrevArgID, but got %d", prev_val)
	}
}

// TestArgSetChooseRecentArg tests recent argument kind in the Choose function of an argument set
func TestArgSetChooseRecentArg(t *testing.T) {
	mockCtl := gomock.NewController(t)
	defer mockCtl.Finish()
	mockRandomizer := NewMockRandomizer(mockCtl)
	n := ArgumentType(1000)

	var calls []*gomock.Call
	// prepare mock calls to fill the queue with values from 1 to QueueLen
	calls = append(calls, mockRandomizer.EXPECT().SampleDistribution(n-1).Return(ArgumentType(4711)).Times(statistics.QueueLen))
	// prepare mock calls to select recent values in ascending order from the queue
	for i := range statistics.QueueLen - 1 {
		calls = append(calls, mockRandomizer.EXPECT().SampleQueue().Return(int(i+1)).Times(1))
	}
	gomock.InOrder(calls...)

	// create argument set and query each queue element subsequentially
	as := NewArgumentSet(n, mockRandomizer)
	for range statistics.QueueLen - 1 {
		val, err := as.Choose(statistics.RecentArgID)
		if err != nil {
			t.Errorf("Unexpected error for RandArgID: %v", err)
		}
		expected_val := ArgumentType(4712)
		if val != expected_val {
			t.Errorf("Expected value %d for RandArgID, but got %d", expected_val, val)
		}
	}
}

// / TestArgSetRemove tests the Remove function of an argument set
func TestArgSetRemove(t *testing.T) {
	mockCtl := gomock.NewController(t)
	defer mockCtl.Finish()
	mockRandomizer := NewMockRandomizer(mockCtl)
	n := ArgumentType(minCardinality + 10)
	mockRandomizer.EXPECT().SampleDistribution(n - 1).Return(ArgumentType(48)).Times(statistics.QueueLen)
	as := NewArgumentSet(n, mockRandomizer)

	mockRandomizer.EXPECT().SampleDistribution(n - 2).Return(ArgumentType(48)).Times(1)
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
