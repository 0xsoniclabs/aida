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

package coverage

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIsValidCarmenPackage(t *testing.T) {
	tests := []struct {
		name string
		pkg  string
		want bool
	}{
		{
			name: "exact module",
			pkg:  carmenModulePrefix,
			want: true,
		},
		{
			name: "sub package",
			pkg:  carmenModulePrefix + "/backend",
			want: true,
		},
		{
			name: "wrong module",
			pkg:  "github.com/0xsoniclabs/aida/stochastic",
			want: false,
		},
		{
			name: "similar prefix without go suffix",
			pkg:  "github.com/0xsoniclabs/carmen/backend",
			want: false,
		},
		{
			name: "empty",
			pkg:  "",
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, isValidCarmenPackage(tt.pkg))
		})
	}
}

func TestFilterCarmenUnits(t *testing.T) {
	units := map[CounterKey]unitInfo{
		{Pkg: 0, Func: 0, Unit: 0}: {PackagePath: carmenModulePrefix + "/backend"},
		{Pkg: 1, Func: 0, Unit: 0}: {PackagePath: "github.com/0xsoniclabs/aida/stochastic"},
		{Pkg: 2, Func: 0, Unit: 0}: {PackagePath: "github.com/0xsoniclabs/carmen"},
	}

	filtered := filterCarmenUnits(units)

	require.Len(t, filtered, 1)
	_, ok := filtered[CounterKey{Pkg: 0, Func: 0, Unit: 0}]
	require.True(t, ok, "expected Carmen package to remain after filtering")
}

func TestFilterCountersForUnits(t *testing.T) {
	allowed := map[CounterKey]unitInfo{
		{Pkg: 0, Func: 0, Unit: 0}: {PackagePath: carmenModulePrefix + "/backend"},
		{Pkg: 0, Func: 1, Unit: 0}: {PackagePath: carmenModulePrefix + "/backend"},
	}
	counts := map[CounterKey]uint32{
		{Pkg: 0, Func: 0, Unit: 0}:    5,
		{Pkg: 99, Func: 99, Unit: 99}: 10,
	}

	filtered := filterCountersForUnits(counts, allowed)

	require.Len(t, filtered, len(allowed))
	require.Equal(t, uint32(5), filtered[CounterKey{Pkg: 0, Func: 0, Unit: 0}])
	require.Equal(t, uint32(0), filtered[CounterKey{Pkg: 0, Func: 1, Unit: 0}], "missing counters should default to zero")
	_, ok := filtered[CounterKey{Pkg: 99, Func: 99, Unit: 99}]
	require.False(t, ok, "non-Carmen counters must be discarded")
}
