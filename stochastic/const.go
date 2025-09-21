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

package stochastic

// Argument classifier for StateDB arguments (account addresses, storage keys,  and storage values).
// The  classifier is based on a combination of counting and queuing statistics to classify arguments
// into the following kinds:
const (
	NoArgID     = iota // no argument
	ZeroArgID          // zero argument
	PrevArgID          // previously seen argument (last access)
	RecentArgID        // recent argument value (found in the counting queue)
	RandArgID          // random access (pick randomly from argument set)
	NewArgID           // new argument (not seen before and increases the argument set)

	NumArgKinds // number of argument kinds
)

// QueueLen sets the length of counting queue (must be greater than one).
const QueueLen = 32

// NumECDFPoints sets the number of points in the empirical cumulative distribution function.
const NumECDFPoints = 300
