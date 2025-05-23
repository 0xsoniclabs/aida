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
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"testing"
)

// checkStrictPartialOrder checks whether ordinal numbers are also a topological ordering.
func checkStrictPartialOrder(por StrictPartialOrder) bool {
	n := len(por)
	for i := 0; i < n; i++ {
		for j := range por[i] {
			if i <= j {
				return false
			}
		}
	}
	return true
}

// TestEmptyMatching tests whether an empty strict partial order returns an empty maximum matching.
func TestEmptyMatching(t *testing.T) {
	por := StrictPartialOrder{}
	if !checkStrictPartialOrder(por) {
		t.Errorf("Ordinal numbers in strict partial order are not topological orderings")
	}

	matches, _ := maxMatching(por)
	if len(matches) != 0 {
		t.Errorf("Expected empty matching, got %d", len(matches))
	}
}

// TestSingletonMatching tests whether a singleton strict partial order returns an empty maximum matching.
func TestSingletonMatching(t *testing.T) {
	por := StrictPartialOrder{
		OrdinalSet{},
	}
	if !checkStrictPartialOrder(por) {
		t.Errorf("Ordinal numbers in strict partial order are not topological orderings")
	}
	matches, err := maxMatching(por)
	if err != nil {
		t.Errorf("Successful maxMatching expected, got %v", err)
	}
	if len(matches) != 0 {
		t.Errorf("Expected empty matching, got %d", len(matches))
	}
}

// TestSimple1Matching tests whether a strict order {0 ~ 1, 0 ~ 2, 1 ~ 2 }
// represented as function {0 |-> {}, 1 |-> {0}, 2 |-> {0, 1}} returns the
// match {1 -> 0, 2 -> 1}.
func TestSimple1Matching(t *testing.T) {
	por := StrictPartialOrder{
		OrdinalSet{},                             // 0 |-> {}
		OrdinalSet{0: struct{}{}},                // 1 |-> {0}
		OrdinalSet{0: struct{}{}, 1: struct{}{}}, // 2 |-> {0, 1}
	}
	if !checkStrictPartialOrder(por) {
		t.Errorf("Ordinal numbers in strict partial order are not topological orderings")
	}
	matches, err := maxMatching(por)
	if err != nil {
		t.Errorf("Error during matching: %v", err)
	}
	if len(matches) != 2 {
		t.Errorf("Wrong number of matches")
	}
	firstMatch := false  // 1 -> 0
	secondMatch := false // 2 -> 1
	for i := 0; i < len(matches); i++ {
		if matches[i][0] == 1 && matches[i][1] == 0 {
			firstMatch = true
		}
		if matches[i][0] == 2 && matches[i][1] == 1 {
			secondMatch = true
		}
	}
	if !firstMatch || !secondMatch {
		t.Errorf("Cannot find either first or second match")
	}
}

// TestSimple2Matching tests whether a strict order {0 ~ 1, 0 ~ 2, 0 ~ 3, 1 ~ 2, 1 ~ 3}
// represented as function {0 |-> {}, 1 |-> {0}, 2 |-> {0, 1}, 3 |-> {0 1}} returns
// match {1 -> 0, 2 -> 1} or the match {1 -> 0, 3 -> 1} indeterministically.
func TestSimple2Matching(t *testing.T) {
	por := StrictPartialOrder{
		OrdinalSet{},                             // 0 |-> {}
		OrdinalSet{0: struct{}{}},                // 1 |-> {0}
		OrdinalSet{0: struct{}{}, 1: struct{}{}}, // 2 |-> {0, 1}
		OrdinalSet{0: struct{}{}, 1: struct{}{}}, // 3 |-> {0, 1}
	}
	if !checkStrictPartialOrder(por) {
		t.Errorf("Ordinal numbers in strict partial order are not topological orderings")
	}
	matches, err := maxMatching(por)
	if err != nil {
		t.Errorf("Error during matching: %v", err)
	}
	if len(matches) != 2 {
		t.Errorf("Wrong number of matches")
	}
	firstMatch := false  // 1 -> 0
	secondMatch := false // 2 -> 1
	thirdMatch := false  // 3 -> 1
	for i := 0; i < len(matches); i++ {
		if matches[i][0] == 1 && matches[i][1] == 0 {
			firstMatch = true
		}
		if matches[i][0] == 2 && matches[i][1] == 1 {
			secondMatch = true
		}
		if matches[i][0] == 3 && matches[i][1] == 1 {
			secondMatch = true
		}
	}
	// Either the edges {1 -> 0, 2 ->1 } or edges {1 -> 0, 3 -> 1} must be found
	if !((firstMatch && secondMatch) || (firstMatch && thirdMatch)) {
		t.Errorf("Cannot find correct matches")
	}
}

// TestEmptyChainCover tests whether an empty strict partial order returns an empty minimum chain cover.
func TestEmptyChainCover(t *testing.T) {
	por := StrictPartialOrder{}
	chains, _ := MinChainCover(por)
	if len(chains) != 0 {
		t.Errorf("Empty matches expected, got %d", len(chains))
	}
}

// TestSimple1MinCover tests whether a strict order {0 ~ 1, 0 ~ 2, 1 ~ 2 } returns the chain cover {[0,1,2]}.
func TestSimple1MinCover(t *testing.T) {
	por := StrictPartialOrder{
		OrdinalSet{},
		OrdinalSet{0: struct{}{}},
		OrdinalSet{0: struct{}{}, 1: struct{}{}},
	}
	chains, err := MinChainCover(por)
	if err != nil {
		t.Errorf("Error during MinChainCover: %v", err)
	}
	if len(chains) != 1 {
		t.Errorf("Wrong number of chains")
	}
	if chains[0][0] != 0 || chains[0][1] != 1 || chains[0][2] != 2 {
		t.Errorf("Chain was not found")
	}
}

// TestSimple2MinCover tests whether a strict order {0 ~ 1, 0 ~ 2, 0 ~ 3, 1 ~ 2, 1 ~ 3}
// represented as function {0 |-> {}, 1 |-> {0}, 2 |-> {0, 1}, 3 |-> {0 1}} returns the
// chains {[0,1,2], [3]} or chains {[0,1,3], [2]}.
func TestSimple2MinCover(t *testing.T) {
	por := StrictPartialOrder{
		OrdinalSet{},                             // 0 |-> {}
		OrdinalSet{0: struct{}{}},                // 1 |-> {0}
		OrdinalSet{0: struct{}{}, 1: struct{}{}}, // 2 |-> {0, 1}
		OrdinalSet{0: struct{}{}, 1: struct{}{}}, // 3 |-> {0, 1}
	}
	chains, err := MinChainCover(por)
	if err != nil {
		t.Errorf("Error during MinChainCover: %v", err)
	}
	if len(chains) != 2 {
		t.Errorf("Wrong number of chains")
	}
	firstChain := false  // 0->1->2
	secondChain := false // 0->1->3
	thirdChain := false  // 2
	forthChain := false  // 3
	for i := 0; i < len(chains); i++ {
		if len(chains[i]) == 3 {
			if chains[i][0] == 0 && chains[i][1] == 1 && chains[i][2] == 2 {
				firstChain = true
			} else if chains[i][0] == 0 && chains[i][1] == 1 && chains[i][3] == 3 {
				secondChain = true
			} else {
				t.Errorf("Wrong chain %v", chains[i])
			}
		} else if len(chains[i]) == 1 {
			if chains[i][0] == 2 {
				thirdChain = true
			} else if chains[i][0] == 3 {
				forthChain = true
			} else {
				t.Errorf("Wrong chain %v", chains[i])
			}
		}
	}
	if !((firstChain && forthChain) || (secondChain && thirdChain)) {
		t.Errorf("Chain was not found")
	}
}

// TestComplexMatching tests whether a strict order {0 ~ 2, 0 ~ 3, 1 ~ 3,
// 2 ~ 4, 3 ~ 5, 4 ~ 6, 5 ~ 6, 5 ~ 7} returns the chains
// { [0, 2, 4, 6], [1, 3, 5, 7] }.
func TestComplexMinCover(t *testing.T) {
	por := StrictPartialOrder{
		OrdinalSet{},                                            // 0 |-> {}
		OrdinalSet{},                                            // 1 |-> {}
		OrdinalSet{0: struct{}{}},                               // 2 |-> {0}
		OrdinalSet{0: struct{}{}, 1: struct{}{}},                // 3 |-> {0, 1}
		OrdinalSet{0: struct{}{}, 2: struct{}{}},                // 4 |-> {2, 0}
		OrdinalSet{0: struct{}{}, 1: struct{}{}, 3: struct{}{}}, // 5 |-> {0, 1, 3}
		OrdinalSet{0: struct{}{}, 1: struct{}{}, 2: struct{}{}, 3: struct{}{}, 4: struct{}{}, 5: struct{}{}}, // 6 |-> {0, 1, 2, 3, 4, 5}
		OrdinalSet{0: struct{}{}, 1: struct{}{}, 3: struct{}{}, 5: struct{}{}},                               // 7 |-> {0, 1, 3, 5}
	}
	chains, err := MinChainCover(por)
	if err != nil {
		t.Errorf("Error during MinChainCover: %v", err)
	}
	if len(chains) != 2 {
		t.Errorf("Wrong number of chains")
	}
	firstChain := false  // 0->1->2
	secondChain := false // 0->1->3
	for i := 0; i < len(chains); i++ {
		if len(chains[i]) == 4 {
			if chains[i][0] == 0 && chains[i][1] == 2 && chains[i][2] == 4 && chains[i][3] == 6 {
				firstChain = true
			} else if chains[i][0] == 1 && chains[i][1] == 3 && chains[i][2] == 5 && chains[i][3] == 7 {
				secondChain = true
			} else {
				t.Errorf("Wrong chain %v", chains[i])
			}
		} else {
			t.Errorf("Wrong chain %v", chains[i])
		}
	}
	if !firstChain || !secondChain {
		t.Errorf("Chain was not found")
	}
}

// newPartialOrderFromAdjMatrix creates a partial order from adjacency matrix string
// intended only as a helper function for testing purpose
func newPartialOrderFromAdjMatrix(mat string) (StrictPartialOrder, error) {
	mat = strings.TrimSpace(mat)
	if !strings.HasPrefix(mat, "[") || !strings.HasSuffix(mat, "]") {
		return nil, fmt.Errorf("Not a matrix")
	}
	mat = mat[1 : len(mat)-1]
	rows := strings.Split(mat, "] [") // split by "] ["

	por := StrictPartialOrder{}
	for _, row := range rows {
		row = strings.Trim(row, "[] ") // Clean up brackets and spaces
		fields := strings.Fields(row)  // Split by whitespace

		ordset := OrdinalSet{}
		for _, f := range fields {
			num, err := strconv.Atoi(f)
			if err != nil {
				return nil, fmt.Errorf("Adj Matrix Conversion failed; %v", err)
			}

			ordset[num] = struct{}{}
		}
		por = append(por, ordset)
	}

	if !checkStrictPartialOrder(por) {
		return nil, fmt.Errorf("Ordinal numbers in strict partial order are not topological orderings")
	}

	return por, nil
}

func TestMinCover_newPartialOrderFromAdjMatrix(t *testing.T) {
	_, err := newPartialOrderFromAdjMatrix("")
	if err == nil || !strings.Contains(err.Error(), "Not a matrix") {
		t.Errorf("Expected failure when not passed a matrix, got %v", err)
	}

	_, err = newPartialOrderFromAdjMatrix("banana")
	if err == nil || !strings.Contains(err.Error(), "Not a matrix") {
		t.Errorf("Expected failure when not passed a matrix, got %v", err)
	}

	_, err = newPartialOrderFromAdjMatrix("0 1 2 3")
	if err == nil || !strings.Contains(err.Error(), "Not a matrix") {
		t.Errorf("Expected failure when not passed a matrix, got %v", err)
	}

	_, err = newPartialOrderFromAdjMatrix("0 1 2 3")
	if err == nil || !strings.Contains(err.Error(), "Not a matrix") {
		t.Errorf("Expected failure when not passed a matrix, got %v", err)
	}

	_, err = newPartialOrderFromAdjMatrix("[0 1 2 3]")
	if err == nil || !strings.Contains(err.Error(), "not topological orderings") {
		t.Errorf("Expected failure when not passed a matrix, got %v", err)
	}

	_, err = newPartialOrderFromAdjMatrix("[[0 1 2 3]]")
	if err == nil || !strings.Contains(err.Error(), "not topological orderings") {
		t.Errorf("Expected failure when not passed a matrix, got %v", err)
	}

	_, err = newPartialOrderFromAdjMatrix("[[0] [1] [2] [banana]")
	if err == nil || !strings.Contains(err.Error(), "Adj Matrix Conversion failed") {
		t.Errorf("Expected failure when parsing banana, got %v", err)
	}

	_, err = newPartialOrderFromAdjMatrix("[[0] [1] [2] [4]")
	if err == nil || !strings.Contains(err.Error(), "not topological orderings") {
		t.Errorf("Expected failure when parsing edge that is out of bound, got %v", err)
	}

	s := "[[] [] [] []]"
	por, err := newPartialOrderFromAdjMatrix(s)
	if err != nil {
		t.Errorf("Expected success when parsing %s, got %v", s, err)
	}
	actual := StrictPartialOrder{
		OrdinalSet{}, // 0 |-> {}
		OrdinalSet{}, // 1 |-> {}
		OrdinalSet{}, // 2 |-> {}
		OrdinalSet{}, // 3 |-> {}
	}
	equal := reflect.DeepEqual(por, actual)
	if !equal {
		t.Errorf("Expected equality when parsing %s, got %+v", s, actual)
	}

	s8 := "[[] [] [0] [0 1] [0 2] [0 1 3] [0 1 2 3 4 5] [0 1 3 5]]"
	por8, err := newPartialOrderFromAdjMatrix(s8)
	if err != nil {
		t.Errorf("Expected success when parsing %s, got %v", s, err)
	}
	actual8 := StrictPartialOrder{
		OrdinalSet{},                                            // 0 |-> {}
		OrdinalSet{},                                            // 1 |-> {}
		OrdinalSet{0: struct{}{}},                               // 2 |-> {0}
		OrdinalSet{0: struct{}{}, 1: struct{}{}},                // 3 |-> {0, 1}
		OrdinalSet{0: struct{}{}, 2: struct{}{}},                // 4 |-> {2, 0}
		OrdinalSet{0: struct{}{}, 1: struct{}{}, 3: struct{}{}}, // 5 |-> {0, 1, 3}
		OrdinalSet{0: struct{}{}, 1: struct{}{}, 2: struct{}{}, 3: struct{}{}, 4: struct{}{}, 5: struct{}{}}, // 6 |-> {0, 1, 2, 3, 4, 5}
		OrdinalSet{0: struct{}{}, 1: struct{}{}, 3: struct{}{}, 5: struct{}{}},                               // 7 |-> {0, 1, 3, 5}
	}
	equal = reflect.DeepEqual(por8, actual8)
	if !equal {
		t.Errorf("Expected equality when parsing %s, got %+v", s, actual8)
	}

}

func TestMinCover_TestGeneratedCases(t *testing.T) {
	filepath := "./adj.dat"
	tests, err := readAdjDat(filepath)
	if err != nil {
		t.Errorf("Failed to read from %s", filepath)
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("[%s|%d]", tt.name, tt.expected), func(t *testing.T) {
			t.Parallel()

			// create StrictPartialOrder
			por, err := newPartialOrderFromAdjMatrix(tt.adj)
			if err != nil {
				t.Errorf("Block %s failed to create por; %v", tt.name, err)
			}

			matching, err := maxMatching(por)
			if err != nil {
				t.Errorf("Block %s failed to maxMatching; %v", tt.name, err)
			}

			chains, err := MinChainCover(por)
			if err != nil {
				t.Errorf("Block %s failed to MinChainCover; %v", tt.name, err)
			}

			// create equivalent graph
			g, err := fromAdjacencyMatrix(tt.adj)
			if err != nil {
				t.Errorf("Block %s failed to create graph; %v", tt.name, err)
			}

			size, err := g.MaxMatching()
			if err != nil {
				t.Errorf("Block %s failed during MaxMatching; %v", tt.name, err)
			}

			// sizes of por and graph must be the same
			if len(por) != int(g.n) {
				t.Errorf("Block %s failed - graph and por size mismatched; graph: %d, len(por): %d", tt.name, g.n, len(por))
			}

			// both construct of matching problem must result in the same size
			if len(matching) != size {
				t.Errorf("Block %s failed - matching size mismatched; graph: %d, mincover: %d", tt.name, size, len(matching))
			}

			// min chain = size of graph - max matching
			if len(chains) != int(g.n)-size {
				t.Errorf("Block %s failed - min chain cover size incorrect; expected: %d, got: %d", tt.name, int(g.n)-size, len(chains))
			}

		})
	}
}
