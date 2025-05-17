package utils

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestRlp_RlpHash(t *testing.T) {
	input := "Hello, World!"
	hashed := RlpHash(input)
	expected := "0x4fb6316a8b79d5448c1dece3c7a55e2dde3d436aad14dd040f1cf5851cf3b713"
	assert.Equal(t, expected, hashed.Hex())
}
