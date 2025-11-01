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
	"fmt"
)

var (
	metaFileMagic = [4]byte{0x00, 0x63, 0x76, 0x6d}
)

type metaFile struct {
	hash        [16]byte
	totalUnits  int
	unitDetails map[CounterKey]unitInfo
}

type metaFileHeader struct {
	Magic        [4]byte
	Version      uint32
	TotalLength  uint64
	Entries      uint64
	MetaFileHash [16]byte
	StrTabOffset uint32
	StrTabLength uint32
	CounterMode  uint8
	CounterGran  uint8
	_            [6]byte
}

type metaSymbolHeader struct {
	Length     uint32
	PkgName    uint32
	PkgPath    uint32
	ModulePath uint32
	MetaHash   [16]byte
	_          byte
	_          [3]byte
	NumFiles   uint32
	NumFuncs   uint32
}

type unitInfo struct {
	PackagePath string
	FuncName    string
	File        string
	StartLine   uint32
	EndLine     uint32
	StartCol    uint32
	EndCol      uint32
	NumStmts    uint32
	lineKeys    []string
}

func parseMetaFile(data []byte) (*metaFile, error) {
	reader := bytes.NewReader(data)

	var hdr metaFileHeader
	if err := binary.Read(reader, binary.LittleEndian, &hdr); err != nil {
		return nil, fmt.Errorf("coverage: read meta header: %w", err)
	}
	if hdr.Magic != metaFileMagic {
		return nil, fmt.Errorf("coverage: invalid meta header magic")
	}

	pkgOffsets := make([]uint64, hdr.Entries)
	for i := range pkgOffsets {
		if err := binary.Read(reader, binary.LittleEndian, &pkgOffsets[i]); err != nil {
			return nil, fmt.Errorf("coverage: read package offset: %w", err)
		}
	}
	pkgLengths := make([]uint64, hdr.Entries)
	for i := range pkgLengths {
		if err := binary.Read(reader, binary.LittleEndian, &pkgLengths[i]); err != nil {
			return nil, fmt.Errorf("coverage: read package length: %w", err)
		}
	}

	// Skip file-level string table; it currently contains only a handful of entries.
	if hdr.StrTabLength > 0 {
		if _, err := reader.Seek(int64(hdr.StrTabLength), 1); err != nil {
			return nil, fmt.Errorf("coverage: skip meta string table: %w", err)
		}
	}

	result := &metaFile{
		hash:        hdr.MetaFileHash,
		unitDetails: make(map[CounterKey]unitInfo),
	}

	for pkgIdx := uint64(0); pkgIdx < hdr.Entries; pkgIdx++ {
		start := pkgOffsets[pkgIdx]
		length := pkgLengths[pkgIdx]
		if start+length > uint64(len(data)) {
			return nil, fmt.Errorf("coverage: malformed package entry %d (out of range)", pkgIdx)
		}
		payload := data[start : start+length]
		if err := parsePackage(uint32(pkgIdx), payload, result); err != nil {
			return nil, err
		}
	}

	return result, nil
}

func parsePackage(pkgIdx uint32, payload []byte, out *metaFile) error {
	reader := bytes.NewReader(payload)

	var hdr metaSymbolHeader
	if err := binary.Read(reader, binary.LittleEndian, &hdr); err != nil {
		return fmt.Errorf("coverage: read package meta header: %w", err)
	}

	funcOffsets := make([]uint32, hdr.NumFuncs)
	for i := range funcOffsets {
		if err := binary.Read(reader, binary.LittleEndian, &funcOffsets[i]); err != nil {
			return fmt.Errorf("coverage: read function offset: %w", err)
		}
	}

	stringTableOffset := int(reader.Size() - int64(reader.Len()))
	strings, err := decodeStringTable(payload[stringTableOffset:])
	if err != nil {
		return fmt.Errorf("coverage: decode string table: %w", err)
	}

	if int(hdr.PkgPath) >= len(strings) {
		return fmt.Errorf("coverage: invalid package path index %d", hdr.PkgPath)
	}
	packagePath := strings[hdr.PkgPath]

	for funcIdx, funcOffset := range funcOffsets {
		if funcOffset >= uint32(len(payload)) {
			return fmt.Errorf("coverage: function offset out of range (%d)", funcOffset)
		}
		funcData := payload[funcOffset:]
		if err := parseFunction(pkgIdx, uint32(funcIdx), packagePath, strings, funcData, out); err != nil {
			return err
		}
	}

	return nil
}

func parseFunction(pkgIdx, funcIdx uint32, pkgPath string, strings []string, data []byte, out *metaFile) error {
	br := newByteReader(data)

	numUnits64, err := br.readULEB128()
	if err != nil {
		return fmt.Errorf("coverage: read func unit count: %w", err)
	}
	numUnits := uint32(numUnits64)

	nameIdx64, err := br.readULEB128()
	if err != nil {
		return fmt.Errorf("coverage: read func name index: %w", err)
	}
	fileIdx64, err := br.readULEB128()
	if err != nil {
		return fmt.Errorf("coverage: read func file index: %w", err)
	}

	nameIdx := int(nameIdx64)
	fileIdx := int(fileIdx64)
	if nameIdx >= len(strings) || fileIdx >= len(strings) {
		return fmt.Errorf("coverage: string table index out of range")
	}

	funcName := strings[nameIdx]
	fileName := strings[fileIdx]

	for unitIdx := uint32(0); unitIdx < numUnits; unitIdx++ {
		stLine, err := br.readULEB128()
		if err != nil {
			return fmt.Errorf("coverage: read unit start line: %w", err)
		}
		stCol, err := br.readULEB128()
		if err != nil {
			return fmt.Errorf("coverage: read unit start column: %w", err)
		}
		enLine, err := br.readULEB128()
		if err != nil {
			return fmt.Errorf("coverage: read unit end line: %w", err)
		}
		enCol, err := br.readULEB128()
		if err != nil {
			return fmt.Errorf("coverage: read unit end column: %w", err)
		}
		nxStmts, err := br.readULEB128()
		if err != nil {
			return fmt.Errorf("coverage: read unit statement count: %w", err)
		}

		key := CounterKey{Pkg: pkgIdx, Func: funcIdx, Unit: unitIdx}
		info := unitInfo{
			PackagePath: pkgPath,
			FuncName:    funcName,
			File:        fileName,
			StartLine:   uint32(stLine),
			EndLine:     uint32(enLine),
			StartCol:    uint32(stCol),
			EndCol:      uint32(enCol),
			NumStmts:    uint32(nxStmts),
			lineKeys:    buildLineKeys(fileName, uint32(stLine), uint32(enLine)),
		}
		out.unitDetails[key] = info
		out.totalUnits++
	}

	// The final flag denotes whether this function is a literal; ignore for now.
	if _, err := br.readULEB128(); err != nil {
		return fmt.Errorf("coverage: read literal flag: %w", err)
	}

	return nil
}
