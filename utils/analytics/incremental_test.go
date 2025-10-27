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
