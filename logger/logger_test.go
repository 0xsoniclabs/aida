package logger

import (
	"testing"
	"time"

	"github.com/op/go-logging"
	"github.com/stretchr/testify/assert"
)

func TestLogger_NewLogger(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		logger := NewLogger("DEBUG", "testModule")
		assert.NotNil(t, logger)
		assert.True(t, logger.IsEnabledFor(logging.DEBUG))
	})

	t.Run("invalid log level", func(t *testing.T) {
		logger := NewLogger("INVALID", "testModule")
		assert.NotNil(t, logger)
		assert.True(t, logger.IsEnabledFor(logging.INFO))
	})
}

func TestLogger_ParseTime(t *testing.T) {
	elapsed := 3661 * time.Second // 1 hour, 1 minute, and 1 second
	hours, minutes, seconds := ParseTime(elapsed)

	assert.Equal(t, uint32(1), hours)
	assert.Equal(t, uint32(1), minutes)
	assert.Equal(t, uint32(1), seconds)
}
