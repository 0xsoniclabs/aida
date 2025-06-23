package dbcomponent

import "testing"

func TestParseDbComponent(t *testing.T) {
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
