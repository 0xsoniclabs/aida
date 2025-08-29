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

package dbcomponent

import "testing"

func TestDbComponent_Parse(t *testing.T) {
	tests := []struct {
		input    string
		expected DbComponent
		err      bool
	}{
		{"all", All, false},
		{"substate", Substate, false},
		{"delete", Delete, false},
		{"update", Update, false},
		{"state-hash", StateHash, false},
		{"block-hash", BlockHash, false},
		{"invalid", "", true},
	}

	for _, test := range tests {
		result, err := ParseDbComponent(test.input)
		if test.err && err == nil {
			t.Errorf("expected error for input %s, got none", test.input)
			continue
		}
		if !test.err && err != nil {
			t.Errorf("unexpected error for input %s: %v", test.input, err)
			continue
		}
		if result != test.expected {
			t.Errorf("for input %s: expected %s, got %s", test.input, test.expected, result)
		}
	}
}
