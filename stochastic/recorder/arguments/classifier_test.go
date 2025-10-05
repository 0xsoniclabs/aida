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
	"encoding/json"
	"reflect"
	"testing"

	"github.com/0xsoniclabs/aida/stochastic"
)

// TestArgClassifierSimple tests for classifying arguments based on previous updates
func TestArgClassifierUpdateSimple(t *testing.T) {
	// create index arg_classifier
	arg_classifier := NewClassifier[int]()

	// check zero argument
	kind := arg_classifier.Classify(0)
	if kind != stochastic.ZeroArgID {
		t.Fatalf("wrong classification %v (zero argument expected)", kind)
	}

	// check new argument
	kind = arg_classifier.Classify(1)
	if kind != stochastic.NewArgID {
		t.Fatalf("wrong classification %v (new argument expected)", kind)
	}

	// check previous argument
	kind = arg_classifier.Classify(1)
	if kind != stochastic.PrevArgID {
		t.Fatalf("wrong classification %v (previous argument)", kind)
	}

	// check new argument
	kind = arg_classifier.Classify(2)
	if kind != stochastic.NewArgID {
		t.Fatalf("wrong classification %v (new argument)", kind)
	}

	// check recent argument
	kind = arg_classifier.Classify(1)
	if kind != stochastic.RecentArgID {
		t.Fatalf("wrong classification %v (new argument)", kind)
	}

	// place arguments into the queue and ensure to displace one and two
	for i := range stochastic.QueueLen {
		arg_classifier.place(i + 3)
	}

	// check random argument
	kind = arg_classifier.Classify(1)
	if kind != stochastic.RandArgID {
		t.Fatalf("wrong classification %v (new argument)", kind)
	}
}

// TestArgClassifierSimpleQueue tests for classifying arguments based on previous updates.
func TestArgClassifierSimpleQueue(t *testing.T) {
	// create index arg_classifier
	arg_classifier := NewClassifier[int]()

	const offset = 10
	// place elements into the queue and ensure it overfills
	for i := range stochastic.QueueLen + offset {
		arg_classifier.place(i)
	}

	// check previous element's class
	class := arg_classifier.get(stochastic.QueueLen + offset - 1)
	if class != stochastic.PrevArgID {
		t.Fatalf("wrong classification (previous item)")
	}

	// check recent element's class (ones that should be in the queue)
	for i := offset; i < stochastic.QueueLen+offset-1; i++ {
		class := arg_classifier.get(i)
		if class != stochastic.RecentArgID {
			t.Fatalf("wrong classification %v (recent data)", i)
		}
	}

	// check class of elements that fell out of the queue
	for i := 1; i < offset; i++ {
		class := arg_classifier.get(i)
		if class != stochastic.RandArgID {
			t.Fatalf("wrong classification %v (random data)", i)
		}
	}

	// check class of new elements
	class = arg_classifier.get(stochastic.QueueLen + offset)
	if class != stochastic.NewArgID {
		t.Fatalf("wrong classification (new data)")
	}

	// check class of new elements
	class = arg_classifier.get(0)
	if class != stochastic.ZeroArgID {
		t.Fatalf("wrong classification (new data)")
	}
}

// TestArgClassifierJSONOutput tests for JSON output of argument classifiers.
func TestArgClassifierJSONOutput(t *testing.T) {
	x := NewClassifier[int]()

	// place arguments in classifier ensure that the queue overfills
	// so that we have some random arguments for the counting statistics.
	for i := range 2 * stochastic.QueueLen {
		x.Classify(i)
		x.get(i - 10)
	}

	// produce stats in JSON format
	jsonX := x.JSON()
	jOut, err := json.Marshal(jsonX)
	if err != nil {
		t.Fatalf("Marshalling failed to produce distribution")
	}

	// read stats back from JSON format
	var jsonY ClassifierJSON
	if err := json.Unmarshal(jOut, &jsonY); err != nil {
		t.Fatalf("Unmarshalling failed to reproduce distribution")
	}
	if !reflect.DeepEqual(jsonX, jsonY) {
		t.Errorf("Unmarshaling mismatch. Expected:\n%+v\nActual:\n%+v", jsonX, jsonY)
	}
}
