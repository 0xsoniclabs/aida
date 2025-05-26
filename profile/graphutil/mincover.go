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

package graphutil

// Computes the minimum chain cover of a strict partial order
// using Koenig's bipartite construction and graph matching. The carrier set
// of the strict partial order is represented by ordinal numbers
// from zero to n-1 where n is the cardinality of the carrier
// set. The ordinal numbers are also the topological numbers of the
// strict partial order.

// OrdinalSet represents a subset of the carrier set
type OrdinalSet map[int]struct{}

// StrictPartialOrder stores a strict partial order as a representative function pre: A -> 2^A.
// Iff a ~ b is a relating pair of elements in the strict partial order, then
// element a is in pre(b).  The ordinal numbers coincide with a topological sort
// of the partial order, i.e., for all i: for all j in pre[i]:  i < j.
type StrictPartialOrder []OrdinalSet

// Chain is a list of ordinal numbers which are pairwise-comparable, and the elements are ordered in ascending order.
// The length of the chain is limited by n and the numbers range from 0 to n-1.
type Chain []int

// ChainSet is a set of chains. The number of sets is limited by n.
type ChainSet []Chain

// MinChainCover computes the minimum chain cover of a strict partial order.
func MinChainCover(order StrictPartialOrder) (ChainSet, error) {
	chains, _, err := minChainCover(order)
	return chains, err
}

// minChainCover is MinChainCover but with exposed matching for testing purposes
func minChainCover(order StrictPartialOrder) (ChainSet, *BipartiteGraph, error) {
	graph := NewBipartiteGraph(uint32(len(order)))
	if graph == nil || graph.n == 0 {
		return ChainSet{}, graph, nil
	}

	for i, set := range order {
		for j := range set {
			err := graph.AddEdge(uint32(i), uint32(j))
			if err != nil {
				return nil, nil, err
			}
		}
	}

	// hopcroft-karp
	_, err := graph.MaxMatching()
	if err != nil {
		return nil, nil, err
	}

	err = CheckConsistentPairing(graph.MatchU, graph.MatchV)
	if err != nil {
		return nil, nil, err
	}

	cover := ChainSet{}
	for u := uint32(0); u < graph.n; u++ {
		if graph.MatchU[u] == NoMatch { // first element in chain
			newChain := Chain{int(u)}
			// iterate until NoMatch is reached
			for ix := u; graph.MatchV[ix] != NoMatch; ix = graph.MatchV[ix] {
				newChain = append(newChain, int(graph.MatchV[ix]))
			}
			cover = append(cover, newChain)
		}
	}

	return cover, graph, nil

}
