// Copyright 2024 Fantom Foundation
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

package tracer

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"testing"
)

func TestArgumentContext_WriteOp(t *testing.T) {
	ctrl := gomock.NewController(t)
	fh := NewMockFileHandler(ctrl)
	ctx := NewArgumentContext(fh)
	uintData := uint16(123)
	byteData := []byte("test data")

	fh.EXPECT().WriteUint16(uintData)
	fh.EXPECT().WriteData(byteData)

	ctx.WriteOp(uintData, byteData)
}

func TestArgumentContext_WriteAddressOp(t *testing.T) {
	ctrl := gomock.NewController(t)
	fh := NewMockFileHandler(ctrl)
	ctx := NewArgumentContext(fh)
	uintData := uint16(123)
	addrData := common.Address{0x1, 0x2, 0x3}
	byteData := []byte("test data")

	// The address is unknown, so we expect it to be written as a byte slice
	fh.EXPECT().WriteData(addrData.Bytes())
	fh.EXPECT().WriteData(byteData)
	ctx.WriteAddressOp(uintData, &addrData, byteData)

	// The address is known, so we expect only the idx is written as uint8
	fh.EXPECT().WriteUint8(uint8(0))
	fh.EXPECT().WriteData(byteData)
	ctx.WriteAddressOp(uintData, &addrData, byteData)
}

func TestArgumentContext_WriteKeyOp(t *testing.T) {
	ctrl := gomock.NewController(t)
	fh := NewMockFileHandler(ctrl)
	ctx := NewArgumentContext(fh)
	uintData := uint16(123)
	addrData1 := common.Address{0x1, 0x2, 0x3}
	addrData2 := common.Address{0x2, 0x3, 0x4}
	keyData1 := common.Hash{0x4, 0x5, 0x6}
	keyData2 := common.Hash{0x5, 0x6, 0x7}
	byteData := []byte("test data")

	// The address and the are unknown, so we expect it to be written as a byte slice
	fh.EXPECT().WriteData(addrData1.Bytes())
	fh.EXPECT().WriteData(keyData1.Bytes())
	fh.EXPECT().WriteData(byteData)
	ctx.WriteKeyOp(uintData, &addrData1, &keyData1, byteData)

	// The address is known, so we expect only the idx is written as uint8, key is unknown hence written as a byte slice
	fh.EXPECT().WriteUint8(uint8(0))
	fh.EXPECT().WriteData(keyData2.Bytes())
	fh.EXPECT().WriteData(byteData)
	ctx.WriteKeyOp(uintData, &addrData1, &keyData2, byteData)

	// The key is known, so we expect only the idx is written as uint8, address is unknown hence written as a byte slice
	fh.EXPECT().WriteData(addrData2.Bytes())
	fh.EXPECT().WriteUint8(uint8(0))
	fh.EXPECT().WriteData(byteData)
	ctx.WriteKeyOp(uintData, &addrData2, &keyData2, byteData)

	// Both is known, so we expect only the idx is written as uint8
	fh.EXPECT().WriteUint8(uint8(1))
	fh.EXPECT().WriteUint8(uint8(1))
	fh.EXPECT().WriteData(byteData)
	ctx.WriteKeyOp(uintData, &addrData2, &keyData2, byteData)
}

func TestArgumentContext_WriteValueOp(t *testing.T) {
	ctrl := gomock.NewController(t)
	fh := NewMockFileHandler(ctrl)
	ctx := NewArgumentContext(fh)
	uintData := uint16(123)
	addrData1 := common.Address{0x1, 0x2, 0x3}
	addrData2 := common.Address{0x2, 0x3, 0x4}
	keyData1 := common.Hash{0x4, 0x5, 0x6}
	keyData2 := common.Hash{0x5, 0x6, 0x7}
	valueData1 := common.Hash{0x6, 0x7, 0x8}
	valueData2 := common.Hash{0x7, 0x8, 0x9}
	valueData3 := common.Hash{0x8, 0x9, 0x10}

	// Everything is unknown, so we expect it to be written as a byte slice
	fh.EXPECT().WriteData(addrData1.Bytes())
	fh.EXPECT().WriteData(keyData1.Bytes())
	fh.EXPECT().WriteData(valueData1.Bytes())
	ctx.WriteValueOp(uintData, &addrData1, &keyData1, &valueData1)

	// The address is known, so we expect only the idx is written as uint8,
	// key and value are unknown hence written as a byte slice
	fh.EXPECT().WriteUint8(uint8(0))
	fh.EXPECT().WriteData(keyData2.Bytes())
	fh.EXPECT().WriteData(valueData2.Bytes())
	ctx.WriteValueOp(uintData, &addrData1, &keyData2, &valueData1)

	// The key is known, so we expect only the idx is written as uint8,
	// address and value are unknown hence written as a byte slice
	fh.EXPECT().WriteData(addrData2.Bytes())
	fh.EXPECT().WriteUint8(uint8(0))
	fh.EXPECT().WriteData(valueData3.Bytes())
	ctx.WriteValueOp(uintData, &addrData2, &keyData2, &valueData3)

	// Everything is known, so we expect only the idx is written as uint8
	fh.EXPECT().WriteUint8(uint8(1))
	fh.EXPECT().WriteUint8(uint8(1))
	fh.EXPECT().WriteData(uint8(1))
	ctx.WriteValueOp(uintData, &addrData1, &keyData1, &valueData1)
}

func TestArgumentContext_Close_ClosesFileHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	fh := NewMockFileHandler(ctrl)
	ctx := NewArgumentContext(fh)
	fh.EXPECT().Close().Return(nil)

	err := ctx.Close()
	assert.NoError(t, err)
}
