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

func TestParseMetaFile_ValidFile(t *testing.T) {
	hash := [16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}

	// Build a complete meta file
	buf := &bytes.Buffer{}

	// Write header
	require.NoError(t, binary.Write(buf, binary.LittleEndian, metaFileHeader{
		Magic:        metaFileMagic,
		Version:      1,
		TotalLength:  0, // will be filled later
		Entries:      1, // one package
		MetaFileHash: hash,
		StrTabOffset: 0,
		StrTabLength: 0,
		CounterMode:  0,
		CounterGran:  0,
	}))

	// Write package offsets and lengths
	packageOffset := uint64(buf.Len() + 16) // after offsets and lengths
	require.NoError(t, binary.Write(buf, binary.LittleEndian, packageOffset))

	packageData := buildTestPackage(t)
	packageLength := uint64(len(packageData))
	require.NoError(t, binary.Write(buf, binary.LittleEndian, packageLength))

	// Write package data
	buf.Write(packageData)

	meta, err := parseMetaFile(buf.Bytes())
	require.NoError(t, err)
	require.Equal(t, hash, meta.hash)
	require.Greater(t, meta.totalUnits, 0)
	require.NotEmpty(t, meta.unitDetails)
}

func TestParseMetaFile_WithStringTableSkip(t *testing.T) {
	hash := [16]byte{9, 8, 7, 6, 5, 4, 3, 2, 1, 0, 1, 2, 3, 4, 5, 6}
	pkg := buildTestPackage(t)

	buf := &bytes.Buffer{}
	require.NoError(t, binary.Write(buf, binary.LittleEndian, metaFileHeader{
		Magic:        metaFileMagic,
		Entries:      1,
		MetaFileHash: hash,
		StrTabLength: 4, // exercise string table skip path
	}))

	packageOffset := uint64(buf.Len() + 16 + 4) // offsets+lengths (16) + strtab (4)
	require.NoError(t, binary.Write(buf, binary.LittleEndian, packageOffset))
	require.NoError(t, binary.Write(buf, binary.LittleEndian, uint64(len(pkg))))

	buf.Write(make([]byte, 4)) // dummy string table
	buf.Write(pkg)

	meta, err := parseMetaFile(buf.Bytes())
	require.NoError(t, err)
	require.Equal(t, hash, meta.hash)
	require.NotEmpty(t, meta.unitDetails)
}

func TestParseMetaFile_Errors(t *testing.T) {
	tests := []struct {
		name   string
		data   []byte
		errMsg string
	}{
		{
			name:   "empty file",
			data:   []byte{},
			errMsg: "read meta header",
		},
		{
			name: "invalid magic",
			data: func() []byte {
				buf := &bytes.Buffer{}
				_ = binary.Write(buf, binary.LittleEndian, metaFileHeader{
					Magic: [4]byte{0x99, 0x99, 0x99, 0x99},
				})
				return buf.Bytes()
			}(),
			errMsg: "invalid meta header magic",
		},
		{
			name: "truncated offsets",
			data: func() []byte {
				buf := &bytes.Buffer{}
				_ = binary.Write(buf, binary.LittleEndian, metaFileHeader{
					Magic:   metaFileMagic,
					Entries: 5,
				})
				// Not enough offsets
				return buf.Bytes()
			}(),
			errMsg: "read package offset",
		},
		{
			name: "lengths read error",
			data: func() []byte {
				buf := &bytes.Buffer{}
				_ = binary.Write(buf, binary.LittleEndian, metaFileHeader{
					Magic:   metaFileMagic,
					Entries: 1,
				})
				// Write only offsets, omit lengths to force error
				_ = binary.Write(buf, binary.LittleEndian, uint64(0))
				return buf.Bytes()
			}(),
			errMsg: "read package length",
		},
		{
			name: "package out of range",
			data: func() []byte {
				buf := &bytes.Buffer{}
				_ = binary.Write(buf, binary.LittleEndian, metaFileHeader{
					Magic:   metaFileMagic,
					Entries: 1,
				})
				_ = binary.Write(buf, binary.LittleEndian, uint64(10000)) // offset beyond file
				_ = binary.Write(buf, binary.LittleEndian, uint64(100))   // length
				return buf.Bytes()
			}(),
			errMsg: "malformed package entry",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := parseMetaFile(tt.data)
			require.Error(t, err)
			require.Contains(t, err.Error(), tt.errMsg)
		})
	}
}

func TestParsePackage_Valid(t *testing.T) {
	packageData := buildTestPackage(t)

	out := &metaFile{
		unitDetails: make(map[CounterKey]unitInfo),
	}

	err := parsePackage(0, packageData, out)
	require.NoError(t, err)
	require.Greater(t, out.totalUnits, 0)
	require.NotEmpty(t, out.unitDetails)

	// Check that we have some unit details
	for key, info := range out.unitDetails {
		require.Equal(t, uint32(0), key.Pkg)
		require.NotEmpty(t, info.PackagePath)
		require.NotEmpty(t, info.FuncName)
		require.NotEmpty(t, info.File)
	}
}

func TestParsePackage_Errors(t *testing.T) {
	tests := []struct {
		name   string
		data   []byte
		errMsg string
	}{
		{
			name:   "empty data",
			data:   []byte{},
			errMsg: "read package meta header",
		},
		{
			name: "truncated function offsets",
			data: func() []byte {
				buf := &bytes.Buffer{}
				_ = binary.Write(buf, binary.LittleEndian, metaSymbolHeader{
					NumFuncs: 10, // claim 10 functions
				})
				// But don't write any offsets
				return buf.Bytes()
			}(),
			errMsg: "read function offset",
		},
		{
			name: "invalid package path index",
			data: func() []byte {
				buf := &bytes.Buffer{}
				_ = binary.Write(buf, binary.LittleEndian, metaSymbolHeader{
					PkgPath:  999, // invalid index
					NumFuncs: 0,
				})
				// Write empty string table
				writeULEB128ToBuffer(buf, 0) // 0 strings
				return buf.Bytes()
			}(),
			errMsg: "invalid package path index",
		},
		{
			name: "function offset out of range",
			data: func() []byte {
				buf := &bytes.Buffer{}
				_ = binary.Write(buf, binary.LittleEndian, metaSymbolHeader{
					PkgPath:  0,
					NumFuncs: 1,
				})
				_ = binary.Write(buf, binary.LittleEndian, uint32(99999)) // offset beyond data
				// Write string table with one entry
				writeULEB128ToBuffer(buf, 1)
				writeULEB128ToBuffer(buf, 3)
				buf.Write([]byte("pkg"))
				return buf.Bytes()
			}(),
			errMsg: "function offset out of range",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out := &metaFile{
				unitDetails: make(map[CounterKey]unitInfo),
			}
			err := parsePackage(0, tt.data, out)
			require.Error(t, err)
			require.Contains(t, err.Error(), tt.errMsg)
		})
	}
}

func TestParseFunction_Valid(t *testing.T) {
	strings := []string{"myFunc", "file.go", "github.com/test/pkg"}

	// Build function data
	buf := &bytes.Buffer{}
	writeULEB128ToBuffer(buf, 2) // 2 units
	writeULEB128ToBuffer(buf, 0) // func name index
	writeULEB128ToBuffer(buf, 1) // file index

	// Unit 0
	writeULEB128ToBuffer(buf, 10) // start line
	writeULEB128ToBuffer(buf, 5)  // start col
	writeULEB128ToBuffer(buf, 15) // end line
	writeULEB128ToBuffer(buf, 10) // end col
	writeULEB128ToBuffer(buf, 3)  // num stmts

	// Unit 1
	writeULEB128ToBuffer(buf, 20) // start line
	writeULEB128ToBuffer(buf, 1)  // start col
	writeULEB128ToBuffer(buf, 25) // end line
	writeULEB128ToBuffer(buf, 8)  // end col
	writeULEB128ToBuffer(buf, 5)  // num stmts

	// Literal flag
	writeULEB128ToBuffer(buf, 0)

	out := &metaFile{
		unitDetails: make(map[CounterKey]unitInfo),
	}

	err := parseFunction(0, 0, "github.com/test/pkg", strings, buf.Bytes(), out)
	require.NoError(t, err)
	require.Equal(t, 2, out.totalUnits)

	// Check unit 0
	info0 := out.unitDetails[CounterKey{Pkg: 0, Func: 0, Unit: 0}]
	require.Equal(t, "github.com/test/pkg", info0.PackagePath)
	require.Equal(t, "myFunc", info0.FuncName)
	require.Equal(t, "file.go", info0.File)
	require.Equal(t, uint32(10), info0.StartLine)
	require.Equal(t, uint32(15), info0.EndLine)
	require.Len(t, info0.lineKeys, 6) // lines 10-15 inclusive

	// Check unit 1
	info1 := out.unitDetails[CounterKey{Pkg: 0, Func: 0, Unit: 1}]
	require.Equal(t, uint32(20), info1.StartLine)
	require.Equal(t, uint32(25), info1.EndLine)
}

func TestParseFunction_Errors(t *testing.T) {
	strings := []string{"func", "file.go"}

	tests := []struct {
		name   string
		data   []byte
		errMsg string
	}{
		{
			name: "start line read error",
			data: func() []byte {
				buf := &bytes.Buffer{}
				writeULEB128ToBuffer(buf, 1) // 1 unit
				writeULEB128ToBuffer(buf, 0) // name index ok
				writeULEB128ToBuffer(buf, 0) // file index ok
				// missing unit data triggers start line read error
				return buf.Bytes()
			}(),
			errMsg: "read unit start line",
		},
		{
			name:   "truncated unit count",
			data:   []byte{},
			errMsg: "read func unit count",
		},
		{
			name: "truncated name index",
			data: func() []byte {
				buf := &bytes.Buffer{}
				writeULEB128ToBuffer(buf, 1) // 1 unit
				// missing name index
				return buf.Bytes()
			}(),
			errMsg: "read func name index",
		},
		{
			name: "invalid string index",
			data: func() []byte {
				buf := &bytes.Buffer{}
				writeULEB128ToBuffer(buf, 1)   // 1 unit
				writeULEB128ToBuffer(buf, 999) // invalid name index
				writeULEB128ToBuffer(buf, 0)   // file index
				return buf.Bytes()
			}(),
			errMsg: "string table index out of range",
		},
		{
			name: "truncated unit data",
			data: func() []byte {
				buf := &bytes.Buffer{}
				writeULEB128ToBuffer(buf, 1)  // 1 unit
				writeULEB128ToBuffer(buf, 0)  // name index
				writeULEB128ToBuffer(buf, 1)  // file index
				writeULEB128ToBuffer(buf, 10) // start line
				// missing rest of unit data
				return buf.Bytes()
			}(),
			errMsg: "read unit start column",
		},
		{
			name: "literal flag error",
			data: func() []byte {
				buf := &bytes.Buffer{}
				writeULEB128ToBuffer(buf, 0) // 0 units
				writeULEB128ToBuffer(buf, 0) // name index
				writeULEB128ToBuffer(buf, 0) // file index
				// literal flag missing
				return buf.Bytes()
			}(),
			errMsg: "read literal flag",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out := &metaFile{
				unitDetails: make(map[CounterKey]unitInfo),
			}
			err := parseFunction(0, 0, "pkg", strings, tt.data, out)
			require.Error(t, err)
			require.Contains(t, err.Error(), tt.errMsg)
		})
	}
}

// Helper functions for building test data

func buildTestPackage(t *testing.T) []byte {
	buf := &bytes.Buffer{}

	// Package header
	require.NoError(t, binary.Write(buf, binary.LittleEndian, metaSymbolHeader{
		PkgPath:  0,
		NumFuncs: 1,
	}))

	// Function offset (points just after this offset array)
	funcOffset := uint32(4) // one uint32 for the offset itself
	require.NoError(t, binary.Write(buf, binary.LittleEndian, funcOffset))

	// String table (placed before function data in this simplified version)
	stringTableStart := buf.Len()

	// Build string table
	strBuf := &bytes.Buffer{}
	writeULEB128ToBuffer(strBuf, 3) // 3 strings
	writeULEB128ToBuffer(strBuf, 7)
	strBuf.Write([]byte("testPkg"))
	writeULEB128ToBuffer(strBuf, 8)
	strBuf.Write([]byte("testFunc"))
	writeULEB128ToBuffer(strBuf, 7)
	strBuf.Write([]byte("test.go"))

	// Function data
	funcBuf := &bytes.Buffer{}
	writeULEB128ToBuffer(funcBuf, 1) // 1 unit
	writeULEB128ToBuffer(funcBuf, 1) // func name = "testFunc"
	writeULEB128ToBuffer(funcBuf, 2) // file = "test.go"
	writeULEB128ToBuffer(funcBuf, 5) // start line
	writeULEB128ToBuffer(funcBuf, 1) // start col
	writeULEB128ToBuffer(funcBuf, 8) // end line
	writeULEB128ToBuffer(funcBuf, 1) // end col
	writeULEB128ToBuffer(funcBuf, 2) // num stmts
	writeULEB128ToBuffer(funcBuf, 0) // literal flag

	// Calculate actual function offset from start of package
	actualFuncOffset := uint32(stringTableStart) + uint32(strBuf.Len())

	// Rebuild with correct offset
	finalBuf := &bytes.Buffer{}
	require.NoError(t, binary.Write(finalBuf, binary.LittleEndian, metaSymbolHeader{
		PkgPath:  0,
		NumFuncs: 1,
	}))
	require.NoError(t, binary.Write(finalBuf, binary.LittleEndian, actualFuncOffset))
	finalBuf.Write(strBuf.Bytes())
	finalBuf.Write(funcBuf.Bytes())

	return finalBuf.Bytes()
}

func writeULEB128ToBuffer(buf *bytes.Buffer, val uint64) {
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
