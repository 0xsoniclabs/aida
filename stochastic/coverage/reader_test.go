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
	"io"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestByteReader_Read(t *testing.T) {
	data := []byte{1, 2, 3, 4, 5}
	br := newByteReader(data)

	// Read first 2 bytes
	b, err := br.read(2)
	require.NoError(t, err)
	require.Equal(t, []byte{1, 2}, b)

	// Read next 3 bytes
	b, err = br.read(3)
	require.NoError(t, err)
	require.Equal(t, []byte{3, 4, 5}, b)

	// Try to read beyond end
	_, err = br.read(1)
	require.Error(t, err)
	require.Equal(t, io.ErrUnexpectedEOF, err)
}

func TestByteReader_ReadULEB128(t *testing.T) {
	tests := []struct {
		name     string
		data     []byte
		expected uint64
		wantErr  bool
	}{
		{
			name:     "single byte value",
			data:     []byte{0x05},
			expected: 5,
		},
		{
			name:     "two byte value",
			data:     []byte{0x80, 0x01},
			expected: 128,
		},
		{
			name:     "zero value",
			data:     []byte{0x00},
			expected: 0,
		},
		{
			name:     "max single byte",
			data:     []byte{0x7f},
			expected: 127,
		},
		{
			name:     "three byte value",
			data:     []byte{0xff, 0xff, 0x03},
			expected: 65535,
		},
		{
			name:    "truncated data",
			data:    []byte{0x80},
			wantErr: true,
		},
		{
			name:    "empty data",
			data:    []byte{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			br := newByteReader(tt.data)
			val, err := br.readULEB128()
			if tt.wantErr {
				require.Error(t, err)
				require.Equal(t, io.ErrUnexpectedEOF, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.expected, val)
			}
		})
	}
}

func TestByteReader_ReadUint32(t *testing.T) {
	// Little endian encoding of 0x12345678
	data := []byte{0x78, 0x56, 0x34, 0x12}
	br := newByteReader(data)

	val, err := br.readUint32()
	require.NoError(t, err)
	require.Equal(t, uint32(0x12345678), val)

	// Try to read beyond end
	_, err = br.readUint32()
	require.Error(t, err)
	require.Equal(t, io.ErrUnexpectedEOF, err)
}

func TestDecodeStringTable(t *testing.T) {
	tests := []struct {
		name     string
		data     []byte
		expected []string
		wantErr  bool
	}{
		{
			name:     "three strings",
			data:     encodeStringTable([]string{"hello", "world", "test"}),
			expected: []string{"hello", "world", "test"},
		},
		{
			name:     "empty table",
			data:     encodeStringTable([]string{}),
			expected: []string{},
		},
		{
			name:     "with empty string",
			data:     encodeStringTable([]string{"foo", "", "bar"}),
			expected: []string{"foo", "", "bar"},
		},
		{
			name:    "truncated count",
			data:    []byte{},
			wantErr: true,
		},
		{
			name:    "truncated length",
			data:    []byte{0x01}, // 1 entry but no length
			wantErr: true,
		},
		{
			name:    "truncated string data",
			data:    []byte{0x01, 0x05}, // 1 entry, length 5, but no data
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			strings, err := decodeStringTable(tt.data)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.expected, strings)
			}
		})
	}
}

func TestBuildLineKeys(t *testing.T) {
	tests := []struct {
		name     string
		file     string
		start    uint32
		end      uint32
		expected []string
	}{
		{
			name:     "single line",
			file:     "foo.go",
			start:    10,
			end:      10,
			expected: []string{"foo.go:10"},
		},
		{
			name:     "multiple lines",
			file:     "bar.go",
			start:    5,
			end:      8,
			expected: []string{"bar.go:5", "bar.go:6", "bar.go:7", "bar.go:8"},
		},
		{
			name:     "zero start",
			file:     "baz.go",
			start:    0,
			end:      5,
			expected: nil,
		},
		{
			name:     "zero end",
			file:     "qux.go",
			start:    5,
			end:      0,
			expected: nil,
		},
		{
			name:     "end before start",
			file:     "bad.go",
			start:    10,
			end:      5,
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			keys := buildLineKeys(tt.file, tt.start, tt.end)
			require.Equal(t, tt.expected, keys)
		})
	}
}

// Helper function to encode a string table for testing
func encodeStringTable(strings []string) []byte {
	br := &byteWriter{}
	br.writeULEB128(uint64(len(strings)))
	for _, s := range strings {
		br.writeULEB128(uint64(len(s)))
		br.write([]byte(s))
	}
	return br.data
}

type byteWriter struct {
	data []byte
}

func (w *byteWriter) write(b []byte) {
	w.data = append(w.data, b...)
}

func (w *byteWriter) writeULEB128(val uint64) {
	for {
		b := byte(val & 0x7f)
		val >>= 7
		if val != 0 {
			b |= 0x80
		}
		w.data = append(w.data, b)
		if val == 0 {
			break
		}
	}
}
