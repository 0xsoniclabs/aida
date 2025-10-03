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

import (
	"errors"
	"fmt"
	"math"
)

const (
	NoMatch     uint32 = math.MaxUint32 // used to indicate NoMatch in MatchU and MatchV
	InfDistance uint32 = math.MaxUint32 // used to indicate largest distance in distance
)

// BipartiteGraph is used to create a bipartite graph of size n x n.
// Hopcroft-Karp algorithm can be performed on it to return maximum matching provided the edges added.
// MaxMatching can be performed once, after which MatchU and MatchV can be accessed directly.
type BipartiteGraph struct {
	n              uint32     // Size of U and V
	adj            [][]uint32 // Adjacency node list
	distance       []uint32   // Distance for BFS
	MatchU, MatchV []uint32   // Matching pair for a node in U and V
}

// NewBipartiteGraph returns a BipartiteGraph of size n x n without any edge.
func NewBipartiteGraph(n uint32) *BipartiteGraph {
	if n == 0 {
		return nil
	}

	return &BipartiteGraph{
		n:   n,
		adj: make([][]uint32, n),
	}
}

// AddEdge adds to the adjancency node list the edge u->v if u, v are valid and u->v hasn't been added before
func (g *BipartiteGraph) AddEdge(u, v uint32) error {
	if u >= g.n || v >= g.n {
		return errors.New("u or v are out of range")
	}

	// do nothing if edge has been added before
	for _, w := range g.adj[u] {
		if v == w {
			return nil
		}
	}

	g.adj[u] = append(g.adj[u], v)
	return nil
}

// BFS for Hopcroft-Karp
func (g *BipartiteGraph) BFS() bool {
	queue := make([]uint32, 0)

	for u := uint32(0); u < g.n; u++ {
		if g.MatchU[u] == NoMatch {
			g.distance[u] = 0
			queue = append(queue, u)
		} else {
			g.distance[u] = InfDistance
		}
	}

	found := false
	for len(queue) > 0 {
		u := queue[0]
		queue = queue[1:]
		for _, v := range g.adj[u] {
			if g.MatchV[v] == NoMatch {
				found = true
			} else if g.distance[g.MatchV[v]] == InfDistance {
				g.distance[g.MatchV[v]] = g.distance[u] + 1
				queue = append(queue, g.MatchV[v])
			}
		}
	}
	return found
}

// DFS for Hopcroft-Karp
func (g *BipartiteGraph) DFS(u uint32) bool {
	for _, v := range g.adj[u] {
		if g.MatchV[v] == NoMatch ||
			(g.distance[g.MatchV[v]] == g.distance[u]+1 && g.DFS(g.MatchV[v])) {
			g.MatchU[u] = v
			g.MatchV[v] = u
			return true
		}
	}

	g.distance[u] = InfDistance
	return false
}

// MaxMatching prepares and executes Hopcroft-Karp algorithm
func (g *BipartiteGraph) MaxMatching() (int, error) {
	if g.distance != nil || g.MatchU != nil || g.MatchV != nil {
		return 0, errors.New("Matching has already been performed")
	}

	// initialize distance, matchU, matchV
	g.distance = make([]uint32, g.n)

	g.MatchU = make([]uint32, g.n)
	for i := range g.MatchU {
		g.MatchU[i] = NoMatch
	}

	g.MatchV = make([]uint32, g.n)
	for i := range g.MatchV {
		g.MatchV[i] = NoMatch
	}

	// Hopcroft-Karp function
	matchingSize := 0
	for g.BFS() {
		for u := uint32(0); u < g.n; u++ {
			if g.MatchU[u] == NoMatch && g.DFS(u) {
				matchingSize++
			}
		}
	}

	return matchingSize, nil
}

// CheckConsistentPairing checks if the provided matching is consistent e.g. agrees with one another.
// Not part of MaxMatching to facilitate testing without going through mocks.
func CheckConsistentPairing(matchU, matchV []uint32) error {
	err := checkConsistentPairing(matchU, matchV)
	if err != nil {
		return err
	}

	err = checkConsistentPairing(matchV, matchU)
	if err != nil {
		return err
	}

	return nil
}

func checkConsistentPairing(matchU, matchV []uint32) error {
	for u := 0; u < len(matchU); u++ {
		v := matchU[u]
		if v != NoMatch && matchV[v] != uint32(u) {
			return fmt.Errorf("inconsistent pairing: u=%d->v=%d but v=%d->u=%d", u, v, v, matchV[v])
		}
	}
	return nil
}

// matching is a list of ordinal number pairs that represents matches in the bipartite graph.
// There can be at most n pairs in the matching, and the numbers range from 0 to n-1.
// matching is moved here for testing purposes.
type matching [][2]uint32

// getMatching returns the matching in [][2]uint32 format
// This is only done to get intermediate matching for testing purposes.
func (g *BipartiteGraph) getMatching() (matching, error) {
	if g == nil || g.n == 0 {
		return matching{}, nil
	}

	if g.distance == nil || g.MatchU == nil || g.MatchV == nil {
		return nil, errors.New("Matching has not been performed")
	}

	matches := matching{}
	for u := uint32(0); u < g.n; u++ {
		v := g.MatchU[u]
		if v != NoMatch {
			matches = append(matches, [2]uint32{u, v})
		}
	}
	return matches, nil
}
