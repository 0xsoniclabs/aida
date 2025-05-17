package utils

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMin(t *testing.T) {
	tests := []struct {
		a, b, expected int
	}{
		{1, 2, 1},
		{2, 1, 1},
		{3, 3, 3},
	}

	for _, test := range tests {
		result := Min(test.a, test.b)
		assert.Equal(t, test.expected, result)
	}
}

func TestMax(t *testing.T) {
	tests := []struct {
		a, b, expected int
	}{
		{1, 2, 2},
		{2, 1, 2},
		{3, 3, 3},
	}

	for _, test := range tests {
		result := Max(test.a, test.b)
		assert.Equal(t, test.expected, result)
	}
}
