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
	"errors"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/holiman/uint256"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"testing"
)

func TestArgumentContext_WriteOp(t *testing.T) {
	ctrl := gomock.NewController(t)
	fw := NewMockFileWriter(ctrl)
	ctx := NewArgumentContext(fw)

	fw.EXPECT().WriteUint16(uint16(125))
	fw.EXPECT().WriteData([]byte{})
	err := ctx.WriteOp(BeginBlockID, []byte{})
	assert.NoError(t, err)
}

func TestArgumentContext_WriteAddressOp(t *testing.T) {
	ctrl := gomock.NewController(t)
	fw := NewMockFileWriter(ctrl)
	ctx := NewArgumentContext(fw)

	addrData1 := common.Address{0x1, 0x2, 0x3}
	addrData2 := common.Address{0x2, 0x3, 0x4}
	byteData := uint256.NewInt(123).Bytes()
	gomock.InOrder(
		// The address is unknown, so we expect it to be written as a byte slice
		fw.EXPECT().WriteUint16(uint16(50)),
		fw.EXPECT().WriteData(addrData1.Bytes()),
		fw.EXPECT().WriteData(byteData),

		// The address is previous, so we do not write any info about it
		fw.EXPECT().WriteUint16(uint16(75)),
		fw.EXPECT().WriteData(byteData),

		// The address is unknown, so we expect it to be written as a byte slice
		fw.EXPECT().WriteUint16(uint16(50)),
		fw.EXPECT().WriteData(addrData2.Bytes()),
		fw.EXPECT().WriteData(byteData),

		// The address is known although not previous are used - written as idx
		fw.EXPECT().WriteUint16(uint16(100)),
		fw.EXPECT().WriteUint8(uint8(0)),
		fw.EXPECT().WriteData(byteData),
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
	fw := NewMockFileWriter(ctrl)
	ctx := NewArgumentContext(fw)
	addrData1 := common.Address{0x1, 0x2, 0x3}
	addrData2 := common.Address{0x2, 0x3, 0x4}
	keyData1 := common.Hash{0x4, 0x5, 0x6}
	keyData2 := common.Hash{0x5, 0x6, 0x7}
	byteData := []byte("test data")

	gomock.InOrder(
		// Everything is unknown - everything is written as bytes slice
		fw.EXPECT().WriteUint16(uint16(2060)),
		fw.EXPECT().WriteData(addrData1.Bytes()),
		fw.EXPECT().WriteData(keyData1.Bytes()),
		fw.EXPECT().WriteData(byteData),

		// Previous address was used - no writing
		// key is unknown hence written as a byte slice
		fw.EXPECT().WriteUint16(uint16(2085)),
		fw.EXPECT().WriteData(keyData2.Bytes()),
		fw.EXPECT().WriteData(byteData),

		// Previous key was used - no writing
		// address is unknown hence written as a byte slice
		fw.EXPECT().WriteUint16(uint16(2065)),
		fw.EXPECT().WriteData(addrData2.Bytes()),
		fw.EXPECT().WriteData(byteData),

		// Previous key and address were used - no writing
		fw.EXPECT().WriteUint16(uint16(2090)),
		fw.EXPECT().WriteData(byteData),

		// Everything is known although not previous are used - everything is written as idx
		fw.EXPECT().WriteUint16(uint16(2120)),
		fw.EXPECT().WriteUint8(uint8(0)),
		fw.EXPECT().WriteUint8(uint8(0)),
		fw.EXPECT().WriteData(byteData),
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
	fw := NewMockFileWriter(ctrl)
	ctx := NewArgumentContext(fw)
	addrData1 := common.Address{0x1, 0x2, 0x3}
	addrData2 := common.Address{0x2, 0x3, 0x4}
	keyData1 := common.Hash{0x4, 0x5, 0x6}
	keyData2 := common.Hash{0x5, 0x6, 0x7}
	valueData1 := common.Hash{0x6, 0x7, 0x8}
	valueData2 := common.Hash{0x7, 0x8, 0x9}
	valueData3 := common.Hash{0x8, 0x9, 0x10}

	gomock.InOrder(
		// Everything is unknown - everyting is written as bytes slice
		fw.EXPECT().WriteUint16(uint16(2687)),
		fw.EXPECT().WriteData(addrData1.Bytes()),
		fw.EXPECT().WriteData(keyData1.Bytes()),
		fw.EXPECT().WriteData(valueData1.Bytes()),

		// Previous address was used - no writing
		// key and value are unknown hence written as a byte slice
		fw.EXPECT().WriteUint16(uint16(2712)),
		fw.EXPECT().WriteData(keyData2.Bytes()),
		fw.EXPECT().WriteData(valueData2.Bytes()),

		// Previous key was used - no writing
		// address and value are unknown hence written as a byte slice
		fw.EXPECT().WriteUint16(uint16(2692)),
		fw.EXPECT().WriteData(addrData2.Bytes()),
		fw.EXPECT().WriteData(valueData3.Bytes()),

		// Everything is known although not previous are used - everything is written as idx
		fw.EXPECT().WriteUint16(uint16(2749)),
		fw.EXPECT().WriteUint8(uint8(0)),
		fw.EXPECT().WriteUint8(uint8(0)),
		fw.EXPECT().WriteUint8(uint8(1)),
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
	fw := NewMockFileWriter(ctrl)
	ctx := NewArgumentContext(fw)
	fw.EXPECT().Close().Return(nil)

	err := ctx.Close()
	assert.NoError(t, err)
}

func TestArgumentContext_writeClassifiedOp_UnknownOp(t *testing.T) {
	ctrl := gomock.NewController(t)
	fw := NewMockFileWriter(ctrl)
	ctx := NewArgumentContext(fw)
	err := ctx.(*argumentContext).writeClassifiedOp(NumOps+1, 0, nil)
	assert.Error(t, err)
	assert.Equal(t, fmt.Sprintf("unexpected argument classification: %d", NumOps+1), err.Error())
}

func TestArgumentContext_ErrorsAreDistributedCorrectly(t *testing.T) {
	ctrl := gomock.NewController(t)
	fw := NewMockFileWriter(ctrl)
	ctx := &argumentContext{
		contracts: NewQueue[common.Address](),
		keys:      NewQueue[common.Hash](),
		values:    NewQueue[common.Hash](),
		file:      fw,
	}

	mockErr := errors.New("mock err")

	{
		// Overflow the number of operations to ensure it is handled correctly
		err := ctx.WriteOp(NumOps, []byte{})
		require.ErrorContains(t, err, "EncodeArgOp: invalid operation/arguments")

		fw.EXPECT().WriteUint16(gomock.Any()).Return(mockErr)
		err = ctx.WriteOp(BeginBlockID, []byte{})
		assert.ErrorIs(t, err, mockErr)

		fw.EXPECT().WriteUint16(gomock.Any())
		fw.EXPECT().WriteData(gomock.Any()).Return(mockErr)
		err = ctx.WriteOp(BeginBlockID, []byte{})
		assert.ErrorIs(t, err, mockErr)
	}
	{
		// Overflow the number of operations to ensure it is handled correctly
		err := ctx.WriteAddressOp(NumOps, &common.Address{0x20}, []byte{})
		require.ErrorContains(t, err, "EncodeArgOp: invalid operation/arguments")

		fw.EXPECT().WriteUint16(gomock.Any()).Return(mockErr)
		err = ctx.WriteAddressOp(AddBalanceID, &common.Address{0x1}, []byte{})
		assert.ErrorIs(t, err, mockErr)

		gomock.InOrder(
			fw.EXPECT().WriteUint16(gomock.Any()),
			fw.EXPECT().WriteData(gomock.Any()).Return(mockErr),
		)
		err = ctx.WriteAddressOp(AddBalanceID, &common.Address{0x2}, []byte{})

		gomock.InOrder(
			fw.EXPECT().WriteUint16(gomock.Any()),
			fw.EXPECT().WriteData(gomock.Any()),
			fw.EXPECT().WriteData(gomock.Any()).Return(mockErr),
		)
		err = ctx.WriteAddressOp(AddBalanceID, &common.Address{0x3}, []byte{})
		assert.ErrorIs(t, err, mockErr)
	}
	{
		// Overflow the number of operations to ensure it is handled correctly
		err := ctx.WriteKeyOp(NumOps, &common.Address{0x21}, &common.Hash{0x21}, []byte{})
		require.ErrorContains(t, err, "EncodeArgOp: invalid operation/arguments")

		fw.EXPECT().WriteUint16(gomock.Any()).Return(mockErr)
		err = ctx.WriteKeyOp(GetStateID, &common.Address{0x4}, &common.Hash{0x1}, []byte{})
		assert.ErrorIs(t, err, mockErr)

		gomock.InOrder(
			fw.EXPECT().WriteUint16(gomock.Any()),
			fw.EXPECT().WriteData(gomock.Any()).Return(mockErr),
		)
		err = ctx.WriteKeyOp(GetStateID, &common.Address{0x5}, &common.Hash{0x2}, []byte{})
		assert.ErrorIs(t, err, mockErr)

		gomock.InOrder(
			fw.EXPECT().WriteUint16(gomock.Any()),
			fw.EXPECT().WriteData(gomock.Any()),
			fw.EXPECT().WriteData(gomock.Any()).Return(mockErr),
		)
		err = ctx.WriteKeyOp(GetStateID, &common.Address{0x6}, &common.Hash{0x3}, []byte{})
		assert.ErrorIs(t, err, mockErr)

		gomock.InOrder(
			fw.EXPECT().WriteUint16(gomock.Any()),
			fw.EXPECT().WriteData(gomock.Any()),
			fw.EXPECT().WriteData(gomock.Any()),
			fw.EXPECT().WriteData(gomock.Any()).Return(mockErr),
		)
		err = ctx.WriteKeyOp(GetStateID, &common.Address{0x7}, &common.Hash{0x4}, []byte{})
		assert.ErrorIs(t, err, mockErr)
	}
	{
		// Overflow the number of operations to ensure it is handled correctly
		err := ctx.WriteValueOp(NumOps, &common.Address{0x22}, &common.Hash{0x22}, &common.Hash{0x23})
		require.ErrorContains(t, err, "EncodeArgOp: invalid operation/arguments")

		fw.EXPECT().WriteUint16(gomock.Any()).Return(mockErr)
		err = ctx.WriteValueOp(SetStateID, &common.Address{0x8}, &common.Hash{0x6}, &common.Hash{0x5})
		assert.ErrorIs(t, err, mockErr)

		gomock.InOrder(
			fw.EXPECT().WriteUint16(gomock.Any()),
			fw.EXPECT().WriteData(gomock.Any()).Return(mockErr),
		)
		err = ctx.WriteValueOp(SetStateID, &common.Address{0x9}, &common.Hash{0x8}, &common.Hash{0x7})
		assert.ErrorIs(t, err, mockErr)

		gomock.InOrder(
			fw.EXPECT().WriteUint16(gomock.Any()),
			fw.EXPECT().WriteData(gomock.Any()),
			fw.EXPECT().WriteData(gomock.Any()).Return(mockErr),
		)
		err = ctx.WriteValueOp(SetStateID, &common.Address{0x10}, &common.Hash{0x10}, &common.Hash{0x9})
		assert.ErrorIs(t, err, mockErr)
		gomock.InOrder(
			fw.EXPECT().WriteUint16(gomock.Any()),
			fw.EXPECT().WriteData(gomock.Any()),
			fw.EXPECT().WriteData(gomock.Any()),
			fw.EXPECT().WriteData(gomock.Any()).Return(mockErr),
		)
		err = ctx.WriteValueOp(SetStateID, &common.Address{0x11}, &common.Hash{0x12}, &common.Hash{0x11})
		assert.ErrorIs(t, err, mockErr)
	}
}

func Test_writeClassifiedOp(t *testing.T) {
	mockErr := errors.New("mock err")
	data := common.Hash{0x23}
	tests := []struct {
		name    string
		class   uint8
		setup   func(m *MockFileWriter)
		wantErr error
	}{
		{
			name:  "ZeroValueID",
			class: ZeroValueID,
			setup: func(m *MockFileWriter) {
				// nothing happens
			},
			wantErr: nil,
		},
		{
			name:  "PreviousValueID",
			class: PreviousValueID,
			setup: func(m *MockFileWriter) {
				// nothing happens
			},
			wantErr: nil,
		},
		{
			name:  "RecentValueID_Success",
			class: RecentValueID,
			setup: func(m *MockFileWriter) {
				m.EXPECT().WriteUint8(uint8(1)).Return(nil)
			},
			wantErr: nil,
		},
		{
			name:  "RecentValueID_Error",
			class: RecentValueID,
			setup: func(m *MockFileWriter) {
				m.EXPECT().WriteUint8(uint8(1)).Return(mockErr)
			},
			wantErr: mockErr,
		},
		{
			name:  "NewValueID_Success",
			class: NewValueID,
			setup: func(m *MockFileWriter) {
				m.EXPECT().WriteData(data.Bytes()).Return(nil)
			},
			wantErr: nil,
		},
		{
			name:  "NewValueID_Error",
			class: NewValueID,
			setup: func(m *MockFileWriter) {
				m.EXPECT().WriteData(data.Bytes()).Return(mockErr)
			},
			wantErr: mockErr,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			fw := NewMockFileWriter(ctrl)
			ctx := &argumentContext{
				file: fw,
			}
			test.setup(fw)
			err := ctx.writeClassifiedOp(test.class, 1, data)
			if test.wantErr != nil {
				assert.ErrorIs(t, err, test.wantErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
