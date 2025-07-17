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
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/holiman/uint256"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"testing"
)

func TestArgumentContext_WriteOp(t *testing.T) {
	ctrl := gomock.NewController(t)
	fh := NewMockFileHandler(ctrl)
	ctx := NewArgumentContext(fh)

	fh.EXPECT().WriteUint16(uint16(125))
	fh.EXPECT().WriteData([]byte{})
	err := ctx.WriteOp(BeginBlockID, []byte{})
	assert.NoError(t, err)
}

func TestArgumentContext_WriteAddressOp(t *testing.T) {
	ctrl := gomock.NewController(t)
	fh := NewMockFileHandler(ctrl)
	ctx := NewArgumentContext(fh)
	addrData1 := common.Address{0x1, 0x2, 0x3}
	addrData2 := common.Address{0x2, 0x3, 0x4}
	byteData := uint256.NewInt(123).Bytes()
	gomock.InOrder(
		// The address is unknown, so we expect it to be written as a byte slice
		fh.EXPECT().WriteUint16(uint16(50)),
		fh.EXPECT().WriteData(addrData1.Bytes()),
		fh.EXPECT().WriteData(byteData),

		// The address is previous, so we do not write any info about it
		fh.EXPECT().WriteUint16(uint16(75)),
		fh.EXPECT().WriteData(byteData),

		// The address is unknown, so we expect it to be written as a byte slice
		fh.EXPECT().WriteUint16(uint16(50)),
		fh.EXPECT().WriteData(addrData2.Bytes()),
		fh.EXPECT().WriteData(byteData),

		// The address is known although not previous are used - written as idx
		fh.EXPECT().WriteUint16(uint16(100)),
		fh.EXPECT().WriteUint8(uint8(0)),
		fh.EXPECT().WriteData(byteData),
	)

	err := ctx.WriteAddressOp(AddBalanceID, &addrData1, byteData)
	assert.NoError(t, err)
	err = ctx.WriteAddressOp(AddBalanceID, &addrData1, byteData)
	assert.NoError(t, err)
	err = ctx.WriteAddressOp(AddBalanceID, &addrData2, byteData)
	assert.NoError(t, err)
	err = ctx.WriteAddressOp(AddBalanceID, &addrData1, byteData)
	assert.NoError(t, err)
}

func TestArgumentContext_WriteKeyOp(t *testing.T) {
	ctrl := gomock.NewController(t)
	fh := NewMockFileHandler(ctrl)
	ctx := NewArgumentContext(fh)
	addrData1 := common.Address{0x1, 0x2, 0x3}
	addrData2 := common.Address{0x2, 0x3, 0x4}
	keyData1 := common.Hash{0x4, 0x5, 0x6}
	keyData2 := common.Hash{0x5, 0x6, 0x7}
	byteData := []byte("test data")

	gomock.InOrder(
		// Everything is unknown - everything is written as bytes slice
		fh.EXPECT().WriteUint16(uint16(2060)),
		fh.EXPECT().WriteData(addrData1.Bytes()),
		fh.EXPECT().WriteData(keyData1.Bytes()),
		fh.EXPECT().WriteData(byteData),

		// Previous address was used - no writing
		// key is unknown hence written as a byte slice
		fh.EXPECT().WriteUint16(uint16(2085)),
		fh.EXPECT().WriteData(keyData2.Bytes()),
		fh.EXPECT().WriteData(byteData),

		// Previous key was used - no writing
		// address is unknown hence written as a byte slice
		fh.EXPECT().WriteUint16(uint16(2065)),
		fh.EXPECT().WriteData(addrData2.Bytes()),
		fh.EXPECT().WriteData(byteData),

		// Previous key and address were used - no writing
		fh.EXPECT().WriteUint16(uint16(2090)),
		fh.EXPECT().WriteData(byteData),

		// Everything is known although not previous are used - everything is written as idx
		fh.EXPECT().WriteUint16(uint16(2120)),
		fh.EXPECT().WriteUint8(uint8(0)),
		fh.EXPECT().WriteUint8(uint8(0)),
		fh.EXPECT().WriteData(byteData),
	)

	err := ctx.WriteKeyOp(GetStateID, &addrData1, &keyData1, byteData)
	assert.NoError(t, err)
	err = ctx.WriteKeyOp(GetStateID, &addrData1, &keyData2, byteData)
	assert.NoError(t, err)
	err = ctx.WriteKeyOp(GetStateID, &addrData2, &keyData2, byteData)
	assert.NoError(t, err)
	err = ctx.WriteKeyOp(GetStateID, &addrData2, &keyData2, byteData)
	assert.NoError(t, err)
	err = ctx.WriteKeyOp(GetStateID, &addrData1, &keyData1, byteData)
	assert.NoError(t, err)
}

func TestArgumentContext_WriteValueOp(t *testing.T) {
	ctrl := gomock.NewController(t)
	fh := NewMockFileHandler(ctrl)
	ctx := NewArgumentContext(fh)
	addrData1 := common.Address{0x1, 0x2, 0x3}
	addrData2 := common.Address{0x2, 0x3, 0x4}
	keyData1 := common.Hash{0x4, 0x5, 0x6}
	keyData2 := common.Hash{0x5, 0x6, 0x7}
	valueData1 := common.Hash{0x6, 0x7, 0x8}
	valueData2 := common.Hash{0x7, 0x8, 0x9}
	valueData3 := common.Hash{0x8, 0x9, 0x10}

	gomock.InOrder(
		// Everything is unknown - everyting is written as bytes slice
		fh.EXPECT().WriteUint16(uint16(2687)),
		fh.EXPECT().WriteData(addrData1.Bytes()),
		fh.EXPECT().WriteData(keyData1.Bytes()),
		fh.EXPECT().WriteData(valueData1.Bytes()),

		// Previous address was used - no writing
		// key and value are unknown hence written as a byte slice
		fh.EXPECT().WriteUint16(uint16(2712)),
		fh.EXPECT().WriteData(keyData2.Bytes()),
		fh.EXPECT().WriteData(valueData2.Bytes()),

		// Previous key was used - no writing
		// address and value are unknown hence written as a byte slice
		fh.EXPECT().WriteUint16(uint16(2692)),
		fh.EXPECT().WriteData(addrData2.Bytes()),
		fh.EXPECT().WriteData(valueData3.Bytes()),

		// Everything is known although not previous are used - everything is written as idx
		fh.EXPECT().WriteUint16(uint16(2749)),
		fh.EXPECT().WriteUint8(uint8(0)),
		fh.EXPECT().WriteUint8(uint8(0)),
		fh.EXPECT().WriteUint8(uint8(1)),
	)

	err := ctx.WriteValueOp(SetStateID, &addrData1, &keyData1, &valueData1)
	assert.NoError(t, err)
	err = ctx.WriteValueOp(SetStateID, &addrData1, &keyData2, &valueData2)
	assert.NoError(t, err)
	err = ctx.WriteValueOp(SetStateID, &addrData2, &keyData2, &valueData3)
	assert.NoError(t, err)
	err = ctx.WriteValueOp(SetStateID, &addrData1, &keyData1, &valueData1)
	assert.NoError(t, err)
}

func TestArgumentContext_Close_ClosesFileHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	fh := NewMockFileHandler(ctrl)
	ctx := NewArgumentContext(fh)
	fh.EXPECT().Close().Return(nil)

	err := ctx.Close()
	assert.NoError(t, err)
}

func TestArgumentContext_writeClassifiedOp_UnknownOp(t *testing.T) {
	ctrl := gomock.NewController(t)
	fh := NewMockFileHandler(ctrl)
	ctx := NewArgumentContext(fh)
	err := ctx.(*argumentContext).writeClassifiedOp(NumOps+1, 0, nil)
	assert.Error(t, err)
	assert.Equal(t, fmt.Sprintf("unexpected argument classification: %d", NumOps+1), err.Error())
}
