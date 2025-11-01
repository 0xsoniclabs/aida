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
	"encoding/binary"
	"fmt"
	"io"
	"strconv"
)

type byteReader struct {
	data []byte
	off  int
}

func newByteReader(data []byte) *byteReader {
	return &byteReader{data: data}
}

func (r *byteReader) read(n int) ([]byte, error) {
	if r.off+n > len(r.data) {
		return nil, io.ErrUnexpectedEOF
	}
	b := r.data[r.off : r.off+n]
	r.off += n
	return b, nil
}

func (r *byteReader) readULEB128() (uint64, error) {
	var (
		shift uint
		value uint64
	)
	for {
		if r.off >= len(r.data) {
			return 0, io.ErrUnexpectedEOF
		}
		b := r.data[r.off]
		r.off++
		value |= (uint64(b&0x7f) << shift)
		if b&0x80 == 0 {
			break
		}
		shift += 7
	}
	return value, nil
}

func decodeStringTable(data []byte) ([]string, error) {
	br := newByteReader(data)
	numEntries64, err := br.readULEB128()
	if err != nil {
		return nil, fmt.Errorf("coverage: read string table entry count: %w", err)
	}
	numEntries := int(numEntries64)
	strings := make([]string, 0, numEntries)
	for idx := 0; idx < numEntries; idx++ {
		sz, err := br.readULEB128()
		if err != nil {
			return nil, fmt.Errorf("coverage: read string length: %w", err)
		}
		if sz == 0 {
			strings = append(strings, "")
			continue
		}
		b, err := br.read(int(sz))
		if err != nil {
			return nil, fmt.Errorf("coverage: read string payload: %w", err)
		}
		strings = append(strings, string(b))
	}
	return strings, nil
}

func buildLineKeys(file string, start, end uint32) []string {
	if start == 0 || end == 0 || end < start {
		return nil
	}
	count := int(end-start) + 1
	keys := make([]string, 0, count)
	for line := start; line <= end; line++ {
		keys = append(keys, file+":"+strconv.FormatUint(uint64(line), 10))
	}
	return keys
}

func (r *byteReader) readUint32() (uint32, error) {
	b, err := r.read(4)
	if err != nil {
		return 0, err
	}
	return binary.LittleEndian.Uint32(b), nil
}
