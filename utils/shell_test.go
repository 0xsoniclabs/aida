package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestShell_Command(t *testing.T) {
	s := NewShell()
	out, err := s.Command("echo", "hello")
	assert.NoError(t, err)
	assert.Equal(t, "hello\n", string(out))
}
