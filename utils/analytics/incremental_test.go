package analytics

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIncrementalStats_String(t *testing.T) {
	obj := IncrementalStats{
		count: 10,
		min:   0,
		max:   0,
		ksum:  0,
		c:     0,
		m1:    0,
		m2:    0,
		m3:    0,
		m4:    0,
	}

	str, err := json.Marshal(obj) //nolint:staticcheck // SA9005: ignore for test comparison
	assert.NoError(t, err)
	assert.Equal(t, string(str), obj.String())
}
