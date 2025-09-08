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
