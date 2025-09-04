// Copyright 2025 Fantom Foundation
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

// ArgSetRandomizer interface for argument sets
type ArgSetRandomizer interface {
	SampleArg(n ArgumentType) ArgumentType // sample argument distribution
	SampleQueue() int                      // sample queue distribution
}

// SampleArgRandomizer interface for argument sets
type SampleArgRandomizer interface {
	SampleArg(n ArgumentType) ArgumentType // sample argument distribution
}

// SampleQueueRandomizer interface for argument sets
type SampleQueueRandomizer interface {
	SampleQueue() int // sample queue distribution
}

// ProxyRandomizer struct
type ProxyRandomizer struct {
	sampleArg SampleArgRandomizer
	sampleQ   SampleQueueRandomizer
}

// NewProxyRandomizer creates a new ProxyRandomizer
func NewProxyRandomizer(argR SampleArgRandomizer, qR SampleQueueRandomizer) *ProxyRandomizer {
	return &ProxyRandomizer{
		sampleArg: argR,
		sampleQ:   qR,
	}
}

// SampleArg samples an argument from a distribution with n possible arguments
func (r *ProxyRandomizer) SampleArg(n ArgumentType) ArgumentType {
	return r.sampleArg.SampleArg(n)
}

// SampleQueue samples an index for a queue
func (r *ProxyRandomizer) SampleQueue() int {
	return r.sampleQ.SampleQueue()
}
