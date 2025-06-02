package dbcomponent

import "testing"

func TestDBComponent(t *testing.T) {
	tests := []struct {
		input    string
		expected DbComponent
	}{
		{"all", All},
		{"substate", Substate},
		{"delete", Delete},
		{"update", Update},
		{"state-hash", StateHash},
	}

	for _, test := range tests {
		result, err := ParseDbComponent(test.input)
		if err != nil {
			t.Errorf("ParseDbComponent(%s) returned error: %v", test.input, err)
			continue
		}
		if result != test.expected {
			t.Errorf("ParseDbComponent(%s) = %v; want %v", test.input, result, test.expected)
		}
	}
}
