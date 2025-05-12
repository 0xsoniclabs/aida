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

package graphutil

import (
	"strings"
	"testing"
)

func Bipartite_TestNewBipartiteGraph(t *testing.T) {
	g0 := NewBipartiteGraph(0)
	if g0 != nil {
		t.Errorf("Expected nil, got a graph of size %d", g0.n)
	}

	g1 := NewBipartiteGraph(1)
	if g1 == nil {
		t.Errorf("Expected a graph of size 1, got a nil")
	}
	if g1.n != 1 {
		t.Errorf("Expected a graph of size 1, got a graph of size %d", g1.n)
	}
}

func Bipartite_TestAddEdge(t *testing.T) {
	g4 := NewBipartiteGraph(4)
	if g4 == nil {
		t.Errorf("Expected a graph of size 4, got a nil")
	}

	// confirm empty adjacents
	for u := uint32(0); u <= g4.n; u++ {
		if len(g4.adj[u]) != 0 {
			t.Errorf("Expected 0 adjacent node for node %d, got %d", u, len(g4.adj[u]))
		}
	}

	err := g4.AddEdge(0, 4)
	if err == nil || !strings.Contains(err.Error(), "u or v are out of range") {
		t.Errorf("Expected error where u or v are out of range, got %v", err)
	}

	err = g4.AddEdge(0, 10000)
	if err == nil || !strings.Contains(err.Error(), "u or v are out of range") {
		t.Errorf("Expected error where u or v are out of range, got %v", err)
	}

	err = g4.AddEdge(10000, 0)
	if err == nil || !strings.Contains(err.Error(), "u or v are out of range") {
		t.Errorf("Expected error where u or v are out of range, got %v", err)
	}

	err = g4.AddEdge(10000, 20000)
	if err == nil || !strings.Contains(err.Error(), "u or v are out of range") {
		t.Errorf("Expected error where u or v are out of range, got %v", err)
	}

	err = g4.AddEdge(0, 1)
	if err != nil {
		t.Errorf("Expected success when adding edge 0->1, got %v", err)
	}
	if len(g4.adj[0]) != 1 {
		t.Errorf("Expected 1 adjacent node for node 0, got %d", len(g4.adj[0]))
	}
	for u := uint32(1); u <= g4.n; u++ {
		if len(g4.adj[u]) != 0 {
			t.Errorf("Expected 0 adjacent node for node %d, got %d", u, len(g4.adj[u]))
		}
	}

	//repeat this and making sure that nothing changes
	err = g4.AddEdge(0, 1)
	if err != nil {
		t.Errorf("Expected success when adding edge 0->1, got %v", err)
	}
	if len(g4.adj[0]) != 1 {
		t.Errorf("Expected 1 adjacent node for node 0, got %d", len(g4.adj[0]))
	}
	for u := uint32(1); u <= g4.n; u++ {
		if len(g4.adj[u]) != 0 {
			t.Errorf("Expected 0 adjacent node for node %d, got %d", u, len(g4.adj[u]))
		}
	}
}

func Bipartite_TestMaxMatchingEdgeless(t *testing.T) {
	g := NewBipartiteGraph(20)
	size, err := g.MaxMatching()
	if err != nil {
		t.Errorf("Expected success when MaxMatching, got %v", err)
	}
	if size != 0 {
		t.Errorf("Expected matching of size 0, got %d", size)
	}
	err = CheckConsistentPairing(g.MatchU, g.MatchV)
	if err != nil {
		t.Errorf("Expected success when CheckConsistentPairing, got %v", err)
	}
}

func Bipartite_TestMaxMatchingSimple(t *testing.T) {
	g := NewBipartiteGraph(2)

	g.AddEdge(0, 0)
	g.AddEdge(0, 1)

	size, err := g.MaxMatching()
	if err != nil {
		t.Errorf("Expected success when MaxMatching, got %v", err)
	}
	if size != 1 {
		t.Errorf("Expected matching of size 1, got %d", size)
	}
	err = CheckConsistentPairing(g.MatchU, g.MatchV)
	if err != nil {
		t.Errorf("Expected success when CheckConsistentPairing, got %v", err)
	}
}

func Bipartite_TestMaxMatchingSimple2(t *testing.T) {
	g := NewBipartiteGraph(4)

	g.AddEdge(0, 0)
	g.AddEdge(1, 0)
	g.AddEdge(2, 2)
	g.AddEdge(3, 3)

	size, err := g.MaxMatching()
	if err != nil {
		t.Errorf("Expected success when MaxMatching, got %v", err)
	}
	if size != 3 {
		t.Errorf("Expected matching of size 3, got %d", size)
	}
	err = CheckConsistentPairing(g.MatchU, g.MatchV)
	if err != nil {
		t.Errorf("Expected success when CheckConsistentPairing, got %v", err)
	}
}

func Bipartite_TestMaxMatchingSimple3(t *testing.T) {
	g := NewBipartiteGraph(4)

	g.AddEdge(0, 0)
	g.AddEdge(0, 1)
	g.AddEdge(1, 0)
	g.AddEdge(2, 2)
	g.AddEdge(3, 2)
	g.AddEdge(3, 3)

	size, err := g.MaxMatching()
	if err != nil {
		t.Errorf("Expected success when MaxMatching, got %v", err)
	}
	if size != 3 {
		t.Errorf("Expected matching of size 3, got %d", size)
	}
	err = CheckConsistentPairing(g.MatchU, g.MatchV)
	if err != nil {
		t.Errorf("Expected success when CheckConsistentPairing, got %v", err)
	}
}

func Bipartite_TestCheckConsistentPairing(t *testing.T) {
	g1 := NewBipartiteGraph(4)
	g1.AddEdge(0, 0)
	g1.AddEdge(0, 1)
	g1.AddEdge(1, 0)
	g1.AddEdge(2, 2)
	g1.AddEdge(3, 2)
	g1.AddEdge(3, 3)
	g1.MaxMatching()

	g2 := NewBipartiteGraph(2)
	g2.AddEdge(0, 0)
	g2.AddEdge(0, 1)
	g2.MaxMatching()

	err := CheckConsistentPairing(g1.MatchU, g1.MatchV)
	if err != nil {
		t.Errorf("Expected success when CheckConsistentPairing, got %v", err)
	}
	err = CheckConsistentPairing(g2.MatchU, g2.MatchV)
	if err != nil {
		t.Errorf("Expected success when CheckConsistentPairing, got %v", err)
	}
	err = CheckConsistentPairing(g1.MatchU, g2.MatchV)
	if err == nil || !strings.Contains(err.Error(), "inconsistent pairing") {
		t.Errorf("Expected inconsistent pairing, got %v", err)
	}
}

func Bipartite_TestMaxMatchingTwice(t *testing.T) {
	g := NewBipartiteGraph(2)

	g.AddEdge(0, 0)
	g.AddEdge(0, 1)

	g.MaxMatching()
	_, err := g.MaxMatching()
	if err == nil || !strings.Contains(err.Error(), "Matching has already been performed") {
		t.Errorf("Expected Matching has already been performed, got %v", err)
	}

}
