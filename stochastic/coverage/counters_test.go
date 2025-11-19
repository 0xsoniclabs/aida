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
	"bytes"
	"encoding/binary"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseCounterFile_ValidFile(t *testing.T) {
	hash := [16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}

	buf := &bytes.Buffer{}

	// Write header
	require.NoError(t, binary.Write(buf, binary.LittleEndian, counterFileHeader{
		Magic:     counterFileMagic,
		Version:   1,
		MetaHash:  hash,
		Flavor:    1, // CtrRaw
		BigEndian: false,
	}))

	// Write segment header
	require.NoError(t, binary.Write(buf, binary.LittleEndian, counterSegmentHeader{
		FcnEntries: 2,
		StrTabLen:  0,
		ArgsLen:    0,
	}))

	// Write function 1: pkg=0, func=0, 3 counters
	require.NoError(t, binary.Write(buf, binary.LittleEndian, uint32(3)))  // numCounters
	require.NoError(t, binary.Write(buf, binary.LittleEndian, uint32(0)))  // pkgIdx
	require.NoError(t, binary.Write(buf, binary.LittleEndian, uint32(0)))  // funcIdx
	require.NoError(t, binary.Write(buf, binary.LittleEndian, uint32(10))) // counter 0
	require.NoError(t, binary.Write(buf, binary.LittleEndian, uint32(20))) // counter 1
	require.NoError(t, binary.Write(buf, binary.LittleEndian, uint32(0)))  // counter 2

	// Write function 2: pkg=0, func=1, 2 counters
	require.NoError(t, binary.Write(buf, binary.LittleEndian, uint32(2)))  // numCounters
	require.NoError(t, binary.Write(buf, binary.LittleEndian, uint32(0)))  // pkgIdx
	require.NoError(t, binary.Write(buf, binary.LittleEndian, uint32(1)))  // funcIdx
	require.NoError(t, binary.Write(buf, binary.LittleEndian, uint32(5)))  // counter 0
	require.NoError(t, binary.Write(buf, binary.LittleEndian, uint32(15))) // counter 1

	// Write footer
	require.NoError(t, binary.Write(buf, binary.LittleEndian, counterFileFooter{
		Magic:       counterFileMagic,
		NumSegments: 1,
	}))

	result, err := parseCounterFile(hash, buf.Bytes())
	require.NoError(t, err)
	require.Len(t, result, 5)

	require.Equal(t, uint32(10), result[CounterKey{Pkg: 0, Func: 0, Unit: 0}])
	require.Equal(t, uint32(20), result[CounterKey{Pkg: 0, Func: 0, Unit: 1}])
	require.Equal(t, uint32(0), result[CounterKey{Pkg: 0, Func: 0, Unit: 2}])
	require.Equal(t, uint32(5), result[CounterKey{Pkg: 0, Func: 1, Unit: 0}])
	require.Equal(t, uint32(15), result[CounterKey{Pkg: 0, Func: 1, Unit: 1}])
}

func TestParseCounterFile_WithAlignment(t *testing.T) {
	hash := [16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}

	buf := &bytes.Buffer{}

	// Write header
	require.NoError(t, binary.Write(buf, binary.LittleEndian, counterFileHeader{
		Magic:     counterFileMagic,
		Version:   1,
		MetaHash:  hash,
		Flavor:    1,
		BigEndian: false,
	}))

	// Write segment with string table (requires alignment)
	require.NoError(t, binary.Write(buf, binary.LittleEndian, counterSegmentHeader{
		FcnEntries: 1,
		StrTabLen:  7, // Odd number to test alignment
		ArgsLen:    0,
	}))

	// Write dummy string table
	buf.Write(make([]byte, 7))

	// Padding byte for 4-byte alignment (7 bytes -> need 1 padding byte)
	buf.Write([]byte{0})

	// Write function data
	require.NoError(t, binary.Write(buf, binary.LittleEndian, uint32(1)))  // numCounters
	require.NoError(t, binary.Write(buf, binary.LittleEndian, uint32(0)))  // pkgIdx
	require.NoError(t, binary.Write(buf, binary.LittleEndian, uint32(0)))  // funcIdx
	require.NoError(t, binary.Write(buf, binary.LittleEndian, uint32(42))) // counter value

	// Write footer
	require.NoError(t, binary.Write(buf, binary.LittleEndian, counterFileFooter{
		Magic:       counterFileMagic,
		NumSegments: 1,
	}))

	result, err := parseCounterFile(hash, buf.Bytes())
	require.NoError(t, err)
	require.Equal(t, uint32(42), result[CounterKey{Pkg: 0, Func: 0, Unit: 0}])
}

func TestParseCounterFile_ULEB128Encoding(t *testing.T) {
	hash := [16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}

	buf := &bytes.Buffer{}

	// Write header with ULEB128 flavor
	require.NoError(t, binary.Write(buf, binary.LittleEndian, counterFileHeader{
		Magic:     counterFileMagic,
		Version:   1,
		MetaHash:  hash,
		Flavor:    2, // CtrULeb128
		BigEndian: false,
	}))

	// Write segment header
	require.NoError(t, binary.Write(buf, binary.LittleEndian, counterSegmentHeader{
		FcnEntries: 1,
		StrTabLen:  0,
		ArgsLen:    0,
	}))

	// Write function data with ULEB128 encoding
	writeULEB128(buf, 2)   // numCounters
	writeULEB128(buf, 5)   // pkgIdx
	writeULEB128(buf, 3)   // funcIdx
	writeULEB128(buf, 100) // counter 0
	writeULEB128(buf, 200) // counter 1

	// Write footer
	require.NoError(t, binary.Write(buf, binary.LittleEndian, counterFileFooter{
		Magic:       counterFileMagic,
		NumSegments: 1,
	}))

	result, err := parseCounterFile(hash, buf.Bytes())
	require.NoError(t, err)
	require.Equal(t, uint32(100), result[CounterKey{Pkg: 5, Func: 3, Unit: 0}])
	require.Equal(t, uint32(200), result[CounterKey{Pkg: 5, Func: 3, Unit: 1}])
}

func TestParseCounterFile_BigEndian(t *testing.T) {
	hash := [16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}

	buf := &bytes.Buffer{}

	// Write header (header is always little endian)
	require.NoError(t, binary.Write(buf, binary.LittleEndian, counterFileHeader{
		Magic:     counterFileMagic,
		Version:   1,
		MetaHash:  hash,
		Flavor:    1,
		BigEndian: true,
	}))

	// Write segment header
	require.NoError(t, binary.Write(buf, binary.LittleEndian, counterSegmentHeader{
		FcnEntries: 1,
		StrTabLen:  0,
		ArgsLen:    0,
	}))

	// Write function data in big endian
	require.NoError(t, binary.Write(buf, binary.BigEndian, uint32(1)))   // numCounters
	require.NoError(t, binary.Write(buf, binary.BigEndian, uint32(0)))   // pkgIdx
	require.NoError(t, binary.Write(buf, binary.BigEndian, uint32(0)))   // funcIdx
	require.NoError(t, binary.Write(buf, binary.BigEndian, uint32(123))) // counter value

	// Write footer
	require.NoError(t, binary.Write(buf, binary.LittleEndian, counterFileFooter{
		Magic:       counterFileMagic,
		NumSegments: 1,
	}))

	result, err := parseCounterFile(hash, buf.Bytes())
	require.NoError(t, err)
	require.Equal(t, uint32(123), result[CounterKey{Pkg: 0, Func: 0, Unit: 0}])
}

func TestParseCounterFile_Errors(t *testing.T) {
	hash := [16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}
	wrongHash := [16]byte{9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9}

	tests := []struct {
		name   string
		data   []byte
		hash   [16]byte
		errMsg string
	}{
		{
			name:   "empty file",
			data:   []byte{},
			hash:   hash,
			errMsg: "read counter header",
		},
		{
			name: "invalid magic",
			data: func() []byte {
				buf := &bytes.Buffer{}
				_ = binary.Write(buf, binary.LittleEndian, counterFileHeader{
					Magic:    [4]byte{0x99, 0x99, 0x99, 0x99},
					MetaHash: hash,
				})
				return buf.Bytes()
			}(),
			hash:   hash,
			errMsg: "invalid counter header magic",
		},
		{
			name: "hash mismatch",
			data: func() []byte {
				buf := &bytes.Buffer{}
				_ = binary.Write(buf, binary.LittleEndian, counterFileHeader{
					Magic:    counterFileMagic,
					MetaHash: wrongHash,
				})
				return buf.Bytes()
			}(),
			hash:   hash,
			errMsg: "counter/meta hash mismatch",
		},
		{
			name: "truncated file",
			data: func() []byte {
				buf := &bytes.Buffer{}
				_ = binary.Write(buf, binary.LittleEndian, counterFileHeader{
					Magic:    counterFileMagic,
					MetaHash: hash,
				})
				// No footer
				return buf.Bytes()
			}(),
			hash:   hash,
			errMsg: "truncated counter file",
		},
		{
			name: "invalid footer magic",
			data: func() []byte {
				buf := &bytes.Buffer{}
				_ = binary.Write(buf, binary.LittleEndian, counterFileHeader{
					Magic:    counterFileMagic,
					MetaHash: hash,
				})
				_ = binary.Write(buf, binary.LittleEndian, counterFileFooter{
					Magic: [4]byte{0x88, 0x88, 0x88, 0x88},
				})
				return buf.Bytes()
			}(),
			hash:   hash,
			errMsg: "invalid counter footer magic",
		},
		{
			name: "unknown flavor",
			data: func() []byte {
				buf := &bytes.Buffer{}
				_ = binary.Write(buf, binary.LittleEndian, counterFileHeader{
					Magic:    counterFileMagic,
					MetaHash: hash,
					Flavor:   99, // unknown flavor
				})
				_ = binary.Write(buf, binary.LittleEndian, counterSegmentHeader{
					FcnEntries: 1,
				})
				_ = binary.Write(buf, binary.LittleEndian, uint32(1)) // will try to read with unknown flavor
				return buf.Bytes()
			}(),
			hash:   hash,
			errMsg: "unknown counter flavor",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := parseCounterFile(tt.hash, tt.data)
			require.Error(t, err)
			require.Contains(t, err.Error(), tt.errMsg)
		})
	}
}

func TestAlignReader(t *testing.T) {
	data := []byte{1, 2, 3, 4, 5, 6, 7, 8}
	reader := bytes.NewReader(data)

	// Read 1 byte, position is now 1
	_, err := reader.ReadByte()
	require.NoError(t, err)

	// Align to 4 bytes should skip to position 4
	err = alignReader(reader, 4)
	require.NoError(t, err)

	pos, _ := reader.Seek(0, 1) // Get current position
	require.Equal(t, int64(4), pos)

	// Already aligned, should not move
	err = alignReader(reader, 4)
	require.NoError(t, err)

	pos, _ = reader.Seek(0, 1)
	require.Equal(t, int64(4), pos)
}

func TestReadUint32Raw(t *testing.T) {
	// Little endian
	data := []byte{0x78, 0x56, 0x34, 0x12}
	reader := bytes.NewReader(data)
	val, err := readUint32Raw(reader, false)
	require.NoError(t, err)
	require.Equal(t, uint32(0x12345678), val)

	// Big endian
	data = []byte{0x12, 0x34, 0x56, 0x78}
	reader = bytes.NewReader(data)
	val, err = readUint32Raw(reader, true)
	require.NoError(t, err)
	require.Equal(t, uint32(0x12345678), val)

	// Truncated
	data = []byte{0x12, 0x34}
	reader = bytes.NewReader(data)
	_, err = readUint32Raw(reader, false)
	require.Error(t, err)
}

func TestReadUint32ULEB(t *testing.T) {
	tests := []struct {
		name     string
		data     []byte
		expected uint32
		wantErr  bool
	}{
		{
			name:     "small value",
			data:     []byte{0x05},
			expected: 5,
		},
		{
			name:     "two byte value",
			data:     []byte{0x80, 0x02},
			expected: 256,
		},
		{
			name:     "zero",
			data:     []byte{0x00},
			expected: 0,
		},
		{
			name:    "truncated",
			data:    []byte{0x80},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := bytes.NewReader(tt.data)
			val, err := readUint32ULEB(reader)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.expected, val)
			}
		})
	}
}

// Helper to write ULEB128 for testing
func writeULEB128(buf *bytes.Buffer, val uint32) {
	for {
		b := byte(val & 0x7f)
		val >>= 7
		if val != 0 {
			b |= 0x80
		}
		buf.WriteByte(b)
		if val == 0 {
			break
		}
	}
}
