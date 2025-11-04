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

import "fmt"

type PanicError struct {
	message string
	stack   []byte
}

func NewPanicError(message string, stack []byte) *PanicError {
	return &PanicError{
		message: message,
		stack:   stack,
	}
}

// Implement the error interface for MyCustomError
func (e *PanicError) Error() string {
	return fmt.Sprintf("PanicError: %s\nStack Trace:\n%s", e.message, string(e.stack))
}
