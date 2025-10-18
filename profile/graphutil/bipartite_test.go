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
	"bufio"
	"errors"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBipartite_NewBipartiteGraph(t *testing.T) {
	g0 := NewBipartiteGraph(0)
	assert.NotNil(t, g0)

	g1 := NewBipartiteGraph(1)
	assert.NotNil(t, g1)
	assert.Equal(t, 1, g1.n)
}

func TestBipartite_AddEdge(t *testing.T) {
	g4 := NewBipartiteGraph(4)
	if g4 == nil {
		t.Errorf("Expected a graph of size 4, got a nil")
	}

	// confirm empty adjacents
	for u := uint32(0); u < g4.n; u++ {
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
	for u := uint32(1); u < g4.n; u++ {
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
	for u := uint32(1); u < g4.n; u++ {
		if len(g4.adj[u]) != 0 {
			t.Errorf("Expected 0 adjacent node for node %d, got %d", u, len(g4.adj[u]))
		}
	}
}

func TestBipartite_MaxMatchingEdgeless(t *testing.T) {
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

func TestBipartite_MaxMatchingSimple(t *testing.T) {
	g := NewBipartiteGraph(2)

	err := g.AddEdge(0, 0)
	assert.NoError(t, err)
	err = g.AddEdge(0, 1)
	assert.NoError(t, err)

	size, err := g.MaxMatching()
	assert.NoError(t, err)
	assert.Equal(t, uint32(1), size)

	err = CheckConsistentPairing(g.MatchU, g.MatchV)
	assert.NoError(t, err)
}

func TestBipartite_MaxMatchingSimple2(t *testing.T) {
	g := NewBipartiteGraph(4)

	err := g.AddEdge(0, 0)
	assert.NoError(t, err)
	err = g.AddEdge(1, 0)
	assert.NoError(t, err)
	err = g.AddEdge(2, 2)
	assert.NoError(t, err)
	err = g.AddEdge(3, 3)
	assert.NoError(t, err)

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

func TestBipartite_MaxMatchingSimple3(t *testing.T) {
	g := NewBipartiteGraph(4)

	err := g.AddEdge(0, 0)
	assert.NoError(t, err)
	err = g.AddEdge(0, 1)
	assert.NoError(t, err)
	err = g.AddEdge(1, 0)
	assert.NoError(t, err)
	err = g.AddEdge(2, 2)
	assert.NoError(t, err)
	err = g.AddEdge(3, 2)
	assert.NoError(t, err)
	err = g.AddEdge(3, 3)
	assert.NoError(t, err)

	size, err := g.MaxMatching()
	if err != nil {
		t.Errorf("Expected success when MaxMatching, got %v", err)
	}
	if size != 4 {
		t.Errorf("Expected matching of size 4, got %d", size)
	}
	err = CheckConsistentPairing(g.MatchU, g.MatchV)
	if err != nil {
		t.Errorf("Expected success when CheckConsistentPairing, got %v", err)
	}
}

func TestBipartite_MaxMatchingBlock4776(t *testing.T) {
	g := NewBipartiteGraph(5)

	var err error
	for _, p := range [][2]uint32{{0, 1}, {0, 2}, {4, 1}, {4, 0}, {4, 3}} {
		err = g.AddEdge(p[0], p[1])
		if err != nil {
			t.Errorf("Expected success when AddEdge, got %v", err)
		}
	}

	ac := 0
	for _, a := range g.adj {
		ac += len(a)
	}
	if ac != 5 {
		t.Errorf("Expected adjacency lists to show 5 edges, got %d", ac)
	}

	size, err := g.MaxMatching()
	if err != nil {
		t.Errorf("Expected success when MaxMatching, got %v", err)
	}
	if size != 2 {
		t.Errorf("Expected matching of size 2, got %d", size)
	}
	err = CheckConsistentPairing(g.MatchU, g.MatchV)
	if err != nil {
		t.Errorf("Expected success when CheckConsistentPairing, got %v", err)
	}
}

func TestBipartite_MaxMatchingBlock4775_Failed(t *testing.T) {
	g := NewBipartiteGraph(5)

	var errs error = nil
	for _, p := range [][2]uint32{{1, 2}, {1, 3}, {5, 2}, {5, 1}, {5, 4}} {
		err := g.AddEdge(p[0], p[1])
		errs = errors.Join(errs, err)
	}

	if errs == nil {
		t.Errorf("Expected failure when adding edges using (1,n) rather than (0,n-1) as index, but found none")
	}
}

func TestBipartite_CheckConsistentPairing(t *testing.T) {
	g1 := NewBipartiteGraph(4)
	err := g1.AddEdge(0, 0)
	assert.NoError(t, err)
	err = g1.AddEdge(0, 1)
	assert.NoError(t, err)
	err = g1.AddEdge(1, 0)
	assert.NoError(t, err)
	err = g1.AddEdge(2, 2)
	assert.NoError(t, err)
	err = g1.AddEdge(3, 2)
	assert.NoError(t, err)
	err = g1.AddEdge(3, 3)
	assert.NoError(t, err)
	_, err = g1.MaxMatching()
	assert.NoError(t, err)

	g2 := NewBipartiteGraph(2)
	err = g2.AddEdge(0, 0)
	assert.NoError(t, err)
	err = g2.AddEdge(0, 1)
	assert.NoError(t, err)
	_, err = g2.MaxMatching()
	assert.NoError(t, err)

	err = CheckConsistentPairing(g1.MatchU, g1.MatchV)
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

func TestBipartite_MaxMatchingTwice(t *testing.T) {
	g := NewBipartiteGraph(2)

	err := g.AddEdge(0, 0)
	assert.NoError(t, err)
	err = g.AddEdge(0, 1)
	assert.NoError(t, err)

	_, err = g.MaxMatching()
	assert.NoError(t, err)
	_, err = g.MaxMatching()
	if err == nil || !strings.Contains(err.Error(), "Matching has already been performed") {
		t.Errorf("Expected Matching has already been performed, got %v", err)
	}
}

// fromAdjacencyMatrix creates a BipartiteGraph from adjacency matrix string
// intended only as a helper function for testing purpose
func fromAdjacencyMatrix(mat string) (*BipartiteGraph, error) {
	mat = strings.TrimSpace(mat)

	if !strings.HasPrefix(mat, "[") || !strings.HasSuffix(mat, "]") {
		return nil, fmt.Errorf("Not a matrix")
	}

	mat = mat[1 : len(mat)-1]
	rows := strings.Split(mat, "] [") // split by "] ["
	n := len(rows)

	g := NewBipartiteGraph(uint32(n))
	if g == nil {
		return nil, fmt.Errorf("Adj Matrix has length 0")
	}

	var checker [][]uint32

	for u, row := range rows {
		row = strings.Trim(row, "[] ") // Clean up brackets and spaces
		fields := strings.Fields(row)  // Split by whitespace

		var rowSlice []uint32
		for _, f := range fields {
			num, err := strconv.Atoi(f)
			if err != nil {
				return nil, fmt.Errorf("Adj Matrix Conversion failed; %v", err)
			}

			num32 := uint32(num)
			rowSlice = append(rowSlice, num32)
			err = g.AddEdge(uint32(u), num32)
			if err != nil {
				return nil, fmt.Errorf("Adj Matrix Add Edge failed; %v", err)
			}
		}
		checker = append(checker, rowSlice)
	}

	equal := reflect.DeepEqual(g.adj, checker)
	if !equal {
		return nil, fmt.Errorf("Adj Matrix not equal")
	}

	return g, nil
}

func TestBipartite_fromAdjacencyMatrix(t *testing.T) {
	_, err := fromAdjacencyMatrix("")
	if err == nil || !strings.Contains(err.Error(), "Not a matrix") {
		t.Errorf("Expected failure when not passed a matrix, got %v", err)
	}

	_, err = fromAdjacencyMatrix("banana")
	if err == nil || !strings.Contains(err.Error(), "Not a matrix") {
		t.Errorf("Expected failure when not passed a matrix, got %v", err)
	}

	_, err = fromAdjacencyMatrix("0 1 2 3")
	if err == nil || !strings.Contains(err.Error(), "Not a matrix") {
		t.Errorf("Expected failure when not passed a matrix, got %v", err)
	}

	_, err = fromAdjacencyMatrix("[0 1 2 3]")
	if err == nil || !strings.Contains(err.Error(), "Adj Matrix Add Edge failed") {
		t.Errorf("Expected failure when not passed a matrix, got %v", err)
	}

	_, err = fromAdjacencyMatrix("[[0 1 2 3]]")
	if err == nil || !strings.Contains(err.Error(), "Adj Matrix Add Edge failed") {
		t.Errorf("Expected failure when not passed a matrix, got %v", err)
	}

	_, err = fromAdjacencyMatrix("[[0] [1] [2] [banana]")
	if err == nil || !strings.Contains(err.Error(), "Adj Matrix Conversion failed") {
		t.Errorf("Expected failure when parsing banana, got %v", err)
	}

	_, err = fromAdjacencyMatrix("[[0] [1] [2] [4]")
	if err == nil || !strings.Contains(err.Error(), "u or v are out of range") {
		t.Errorf("Expected failure when parsing edge that is out of bound, got %v", err)
	}

	s := "[[] [] [] []]"
	g, err := fromAdjacencyMatrix(s)
	if err != nil {
		t.Errorf("Expected success when parsing %s, got %v", s, err)
	}
	actual := NewBipartiteGraph(4)
	equal := reflect.DeepEqual(g.adj, actual.adj)
	if !equal {
		t.Errorf("Expected equality when parsing %s, got %+v", s, actual.adj)
	}

	s4 := "[[0 1] [0] [2] [2 3]]"
	g4, err := fromAdjacencyMatrix(s4)
	if err != nil {
		t.Errorf("Expected success when parsing %s, got %v", s4, err)
	}
	actual4 := NewBipartiteGraph(4)
	err = actual4.AddEdge(0, 0)
	assert.NoError(t, err)
	err = actual4.AddEdge(0, 1)
	assert.NoError(t, err)
	err = actual4.AddEdge(1, 0)
	assert.NoError(t, err)
	err = actual4.AddEdge(2, 2)
	assert.NoError(t, err)
	err = actual4.AddEdge(3, 2)
	assert.NoError(t, err)
	err = actual4.AddEdge(3, 3)
	assert.NoError(t, err)
	equal = reflect.DeepEqual(g4.adj, actual4.adj)
	if !equal {
		t.Errorf("Expected equality when parsing %s, got %+v", s4, actual4.adj)
	}
}

type maxMatchingTestcase struct {
	name     string
	expected int
	adj      string
}

// readAdjData generate max matching test cases from serialized text file.
// The input text file must have the following format:
// 1. one line per test case
// 2. <name>;<expected size>;<adj matrix> // ;-separated
func readAdjDat(filepath string) ([]maxMatchingTestcase, error) {
	var testcases []maxMatchingTestcase

	file, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}
	defer func() {
		if closeErr := file.Close(); closeErr != nil {
			err = errors.Join(err, closeErr)
		}
	}()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		tokens := strings.Split(scanner.Text(), ";")
		expected, err := strconv.Atoi(tokens[1])
		if err != nil {
			return nil, fmt.Errorf("failed to generate max-matching test case; %v", err)
		}

		testcases = append(testcases, maxMatchingTestcase{
			name:     tokens[0],
			expected: expected,
			adj:      tokens[2],
		})
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return testcases, nil
}

func TestBipartite_TestGeneratedCases(t *testing.T) {
	filepath := "./adj.dat"
	tests, err := readAdjDat(filepath)
	if err != nil {
		t.Errorf("Failed to read from %s", filepath)
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("[%s|%d]", tt.name, tt.expected), func(t *testing.T) {
			t.Parallel()

			g, err := fromAdjacencyMatrix(tt.adj)
			if err != nil {
				t.Errorf("Block %s failed to create graph; %v", tt.name, err)
			}

			size, err := g.MaxMatching()
			if err != nil {
				t.Errorf("Block %s failed during MaxMatching; %v", tt.name, err)
			}

			if size != tt.expected {
				t.Errorf("expected %d, got %d", tt.expected, size)
			}
		})
	}
}
