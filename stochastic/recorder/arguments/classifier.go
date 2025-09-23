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

import "github.com/0xsoniclabs/aida/stochastic"

// Classifier struct for account addresses, storage keys, and storage values.
type Classifier[T comparable] struct {
	cstats count[T] // counting statistics for arguments
	qstats queue[T] // counting queue statistics for queue accesses
}

// NewClassifier creates a new argument classifier.
func NewClassifier[T comparable]() Classifier[T] {
	return Classifier[T]{newCount[T](), newQueue[T]()}
}

// Classify the argument classifier with a new argument and return its kind.
func (a *Classifier[T]) Classify(data T) int {
	kind := a.get(data)
	a.place(data)
	return kind
}

// place() places the argument into the counting and queuing statistics.
func (a *Classifier[T]) place(data T) {
	var zeroValue T
	if data == zeroValue {
		return // don't place zero value argument into argument/queue stats
	}
	if a.qstats.findPos(data) == -1 {
		// argument not found in the counting queue; place into counting statistics
		a.cstats.place(data)
	}
	a.qstats.place(data) // place data into queuing statistics
}

// get classification of an argument depending on previous placements.
func (a *Classifier[T]) get(data T) int {
	// check zero value
	var zeroValue T
	if data == zeroValue {
		return stochastic.ZeroArgID
	}
	switch a.qstats.findPos(data) {
	case -1:
		// data not found in the queuing statistics
		// => check argument counting statistics
		if !a.cstats.exists(data) {
			return stochastic.NewArgID
		} else {
			return stochastic.RandArgID
		}
	case 0:
		// previous entry
		return stochastic.PrevArgID
	default:
		// data found in queuing statistics with a queue position > 0
		return stochastic.RecentArgID
	}
}

// ClassifierJSON is the JSON output for the quantitiative description
// of an argument classifier. It contains the ECDF of the counting statistics
// and the distribution of the positions of the recently accessed arguments
// in the queue.
type ClassifierJSON struct {
	Counting ArgStatsJSON   `json:"counting"`
	Queuing  QueueStatsJSON `json:"queue"`
}

// JSON produces output for the classifier statistics.
func (a *Classifier[T]) JSON() ClassifierJSON {
	return ClassifierJSON{a.cstats.json(), a.qstats.json()}
}
