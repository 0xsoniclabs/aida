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

package executor

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPanicError_Error(t *testing.T) {
	err := &PanicError{
		message: "something went wrong",
		stack:   []byte("stack trace here"),
	}
	expected := "PanicError: something went wrong\nStack Trace:\nstack trace here"
	assert.Equal(t, expected, err.Error())
}

func TestNewPanicError(t *testing.T) {
	err := NewPanicError("error message", []byte("stack trace"))
	assert.Equal(t, "error message", err.message)
	assert.Equal(t, []byte("stack trace"), err.stack)
}
