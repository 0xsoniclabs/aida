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

package classifier

import "github.com/0xsoniclabs/aida/stochastic"

// ArgClassifier struct for account addresses, storage keys, and storage values.
type ArgClassifier[T comparable] struct {
	cstats argCount[T]   // counting statistics for arguments
	qstats countQueue[T] // counting queue statistics for queue accesses
}

// NewArgClassifier creates a new argument classifier.
func NewArgClassifier[T comparable]() ArgClassifier[T] {
	return ArgClassifier[T]{newArgCount[T](), NewCountQueue[T]()}
}

// Update the argument classifier with a new argument and return its kind.
func (a *ArgClassifier[T]) Update(data T) stochastic.ArgKind {
	kind := a.classify(data)
	a.place(data)
	return kind
}

// places the argument into the counting and queuing statistics.
func (a *ArgClassifier[T]) place(data T) {
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

// classify an argument depending on previous placements.
func (a *ArgClassifier[T]) classify(data T) stochastic.ArgKind {
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

// ArgClassifierJSON is the JSON output for the quantitiative description
// of an argument classifier. It contains the ECDF of the counting statistics
// and the distribution of the positions of the recently accessed arguments
// in the queue.
type ArgClassifierJSON struct {
	Counting ArgStatsJSON   `json:"counting"`
	Queuing  QueueStatsJSON `json:"queue"`
}

// NewArgClassifierJSON produces JSON output for an access statistics.
func (a *ArgClassifier[T]) JSON() ArgClassifierJSON {
	return ArgClassifierJSON{a.cstats.json(), a.qstats.json()}
}
