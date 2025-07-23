package tracer

import (
	"bytes"
	"encoding/gob"
	"errors"
	"github.com/0xsoniclabs/aida/txcontext"
	"github.com/Fantom-foundation/lachesis-base/common/bigendian"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/tracing"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/params"
	"github.com/holiman/uint256"
	"github.com/klauspost/compress/gzip"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"math"
	"math/big"
	"os"
	"testing"
)

func TestNewFileReader_ErrorCases(t *testing.T) {
	emptyFile := t.TempDir() + "/empty_file"
	create, err := os.Create(emptyFile)
	require.NoError(t, err)
	require.NoError(t, create.Close())
	tests := []struct {
		name     string
		filepath string
		wantErr  string
	}{
		{
			name:     "file does not exist",
			filepath: "non_existent_file",
			wantErr:  "could not stat file: non_existent_file, does it exist?",
		},
		{
			name:     "file is a directory",
			filepath: t.TempDir(),
			wantErr:  "given path to trace file is a directory",
		},
		{
			name:     "file is empty",
			filepath: emptyFile,
			wantErr:  "given trace file is empty",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := NewFileReader(test.filepath)
			require.Error(t, err)
			require.Contains(t, err.Error(), test.wantErr)
		})
	}
}

func TestNewFileReader_Success(t *testing.T) {
	tempFile := t.TempDir() + "/test_file.gz"
	file, err := os.Create(tempFile)
	require.NoError(t, err)
	writer := gzip.NewWriter(file)
	_, err = writer.Write([]byte("test data for file reader"))
	require.NoError(t, err)
	require.NoError(t, writer.Close())

	reader, err := NewFileReader(tempFile)
	require.NoError(t, err)
	require.NotNil(t, reader)
	_, ok := reader.(*fileReader)
	require.True(t, ok)

	// Ensure the reader can be closed without error
	err = reader.Close()
	require.NoError(t, err)
}

func TestFileReader_Read(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockErr := errors.New("mock error")
	tests := []struct {
		name    string
		wantErr error
		Read    func(fr *fileReader) error
		setup   func(m *MockReadBuffer)
	}{
		{
			name:    "ReadData_Success",
			wantErr: nil,
			Read: func(fr *fileReader) error {
				_, err := fr.ReadFixedSizeData(1)
				return err
			},
			setup: func(m *MockReadBuffer) {
				m.EXPECT().Read(gomock.Any()).MinTimes(1).Return(1, nil)
			},
		},
		{
			name:    "ReadData_Error",
			wantErr: mockErr,
			Read: func(fr *fileReader) error {
				_, err := fr.ReadFixedSizeData(1)
				return err
			},
			setup: func(m *MockReadBuffer) {
				m.EXPECT().Read(gomock.Any()).MinTimes(1).Return(0, mockErr)
			},
		},
		{
			name:    "ReadByte_Success",
			wantErr: nil,
			Read: func(fr *fileReader) error {
				_, err := fr.ReadUint8()
				return err
			},
			setup: func(m *MockReadBuffer) {
				m.EXPECT().ReadByte().Return(uint8(3), nil)
			},
		},
		{
			name:    "ReadByte_Error",
			wantErr: mockErr,
			Read: func(fr *fileReader) error {
				_, err := fr.ReadUint8()
				return err
			},
			setup: func(m *MockReadBuffer) {
				m.EXPECT().ReadByte().Return(uint8(0), mockErr)
			},
		},
		{
			name:    "ReadUint16_Success",
			wantErr: nil,
			Read: func(fr *fileReader) error {
				_, err := fr.ReadUint16()
				return err
			},
			setup: func(m *MockReadBuffer) {
				m.EXPECT().Read(gomock.Any()).MinTimes(1).Return(1, nil)
			},
		},
		{
			name:    "ReadUint16_Error",
			wantErr: mockErr,
			Read: func(fr *fileReader) error {
				_, err := fr.ReadUint16()
				return err
			},
			setup: func(m *MockReadBuffer) {
				m.EXPECT().Read(gomock.Any()).MinTimes(1).Return(0, mockErr)
			},
		},
		{
			name:    "ReadUint32_Success",
			wantErr: nil,
			Read: func(fr *fileReader) error {
				_, err := fr.ReadUint32()
				return err
			},
			setup: func(m *MockReadBuffer) {
				m.EXPECT().Read(gomock.Any()).MinTimes(1).Return(1, nil)
			},
		},
		{
			name:    "ReadUint32_Error",
			wantErr: mockErr,
			Read: func(fr *fileReader) error {
				_, err := fr.ReadUint32()
				return err
			},
			setup: func(m *MockReadBuffer) {
				m.EXPECT().Read(gomock.Any()).MinTimes(1).Return(0, mockErr)
			},
		},
		{
			name:    "ReadUint64_Success",
			wantErr: nil,
			Read: func(fr *fileReader) error {
				_, err := fr.ReadUint64()
				return err
			},
			setup: func(m *MockReadBuffer) {
				m.EXPECT().Read(gomock.Any()).MinTimes(1).Return(1, nil)
			},
		},
		{
			name:    "ReadUint64_Error",
			wantErr: mockErr,
			Read: func(fr *fileReader) error {
				_, err := fr.ReadUint64()
				return err
			},
			setup: func(m *MockReadBuffer) {
				m.EXPECT().Read(gomock.Any()).MinTimes(1).Return(0, mockErr)
			},
		},
		{
			name:    "ReadHash_Success",
			wantErr: nil,
			Read: func(fr *fileReader) error {
				_, err := fr.ReadHash()
				return err
			},
			setup: func(m *MockReadBuffer) {
				m.EXPECT().Read(gomock.Any()).MinTimes(1).Return(1, nil)
			},
		},
		{
			name:    "ReadHash_Error",
			wantErr: mockErr,
			Read: func(fr *fileReader) error {
				_, err := fr.ReadHash()
				return err
			},
			setup: func(m *MockReadBuffer) {
				m.EXPECT().Read(gomock.Any()).MinTimes(1).Return(0, mockErr)
			},
		},
		{
			name:    "ReadAddr_Success",
			wantErr: nil,
			Read: func(fr *fileReader) error {
				_, err := fr.ReadAddr()
				return err
			},
			setup: func(m *MockReadBuffer) {
				m.EXPECT().Read(gomock.Any()).MinTimes(1).Return(1, nil)
			},
		},
		{
			name:    "ReadAddr_Error",
			wantErr: mockErr,
			Read: func(fr *fileReader) error {
				_, err := fr.ReadAddr()
				return err
			},
			setup: func(m *MockReadBuffer) {
				m.EXPECT().Read(gomock.Any()).MinTimes(1).Return(0, mockErr)
			},
		},
		{
			name:    "ReadBalanceChange_Success",
			wantErr: nil,
			Read: func(fr *fileReader) error {
				_, _, err := fr.ReadBalanceChange()
				return err
			},
			setup: func(m *MockReadBuffer) {
				m.EXPECT().Read(gomock.Any()).MinTimes(1).Return(1, nil)
				m.EXPECT().ReadByte().Return(byte(1), nil)
			},
		},
		{
			name:    "ReadBalanceChange_Error",
			wantErr: mockErr,
			Read: func(fr *fileReader) error {
				_, _, err := fr.ReadBalanceChange()
				return err
			},
			setup: func(m *MockReadBuffer) {
				m.EXPECT().Read(gomock.Any()).MinTimes(1).Return(0, mockErr)
			},
		},
		{
			name:    "ReadUnknownSizeData_Success",
			wantErr: nil,
			Read: func(fr *fileReader) error {
				_, err := fr.ReadVariableSizeData()
				return err
			},
			setup: func(m *MockReadBuffer) {
				// Firstly the size is read, then the code itself
				m.EXPECT().Read(gomock.Any()).MinTimes(2).Return(1, nil)
			},
		},
		{
			name:    "ReadUnknownSizeData_Error",
			wantErr: mockErr,
			Read: func(fr *fileReader) error {
				_, err := fr.ReadVariableSizeData()
				return err
			},
			setup: func(m *MockReadBuffer) {
				m.EXPECT().Read(gomock.Any()).MinTimes(1).Return(0, mockErr)
			},
		},
		{
			name:    "ReadNonceChange_Success",
			wantErr: nil,
			Read: func(fr *fileReader) error {
				_, _, err := fr.ReadNonceChange()
				return err
			},
			setup: func(m *MockReadBuffer) {
				m.EXPECT().Read(gomock.Any()).MinTimes(1).Return(1, nil)
			},
		},
		{
			name:    "ReadNonceChange_Error",
			wantErr: mockErr,
			Read: func(fr *fileReader) error {
				_, _, err := fr.ReadNonceChange()
				return err
			},
			setup: func(m *MockReadBuffer) {
				m.EXPECT().Read(gomock.Any()).MinTimes(1).Return(0, mockErr)
			},
		},
		{
			name:    "ReadBool_Success",
			wantErr: nil,
			Read: func(fr *fileReader) error {
				_, err := fr.ReadBool()
				return err
			},
			setup: func(m *MockReadBuffer) {
				m.EXPECT().ReadByte().Return(byte(1), nil)
			},
		},
		{
			name:    "ReadBool_Error",
			wantErr: mockErr,
			Read: func(fr *fileReader) error {
				_, err := fr.ReadBool()
				return err
			},
			setup: func(m *MockReadBuffer) {
				m.EXPECT().ReadByte().Return(byte(0), mockErr)
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			buf := NewMockReadBuffer(ctrl)
			test.setup(buf)
			fr := &fileReader{
				reader: buf,
			}
			err := test.Read(fr)
			if test.wantErr != nil {
				require.ErrorIs(t, err, test.wantErr)
			} else {
				require.NoError(t, err)
			}

		})
	}
}

func TestFileReader_ReadBalanceChange(t *testing.T) {
	wantBalance := uint256.NewInt(math.MaxUint64)
	wantReason := tracing.BalanceChangeTransfer
	size := uint32(wantBalance.ByteLen())
	data := append(bigendian.Uint32ToBytes(size), wantBalance.Bytes()...)
	data = append(data, byte(wantReason))
	// reset the buffer and add the data in correct order
	fr := &fileReader{
		reader: bytes.NewBuffer(data),
	}
	gotBalance, gotReason, err := fr.ReadBalanceChange()
	require.NoError(t, err)
	require.Equal(t, wantBalance, gotBalance)
	require.Equal(t, wantReason, gotReason)
}

func TestFileReader_ReadNonce(t *testing.T) {
	wantBalance := uint64(135)
	wantReason := tracing.NonceChangeNewContract
	data := append(bigendian.Uint64ToBytes(wantBalance), byte(wantReason))
	data = append(data, byte(wantReason))
	// reset the buffer and add the data in correct order
	fr := &fileReader{
		reader: bytes.NewBuffer(data),
	}
	gotBalance, gotReason, err := fr.ReadNonceChange()
	require.NoError(t, err)
	require.Equal(t, wantBalance, gotBalance)
	require.Equal(t, wantReason, gotReason)
}

func TestFileReader_ReadRules(t *testing.T) {
	buf := &bytes.Buffer{}
	enc := gob.NewEncoder(buf)
	gob.Register(params.Rules{})
	want := params.Rules{
		ChainID:    big.NewInt(146),
		IsBerlin:   true,
		IsLondon:   true,
		IsMerge:    true,
		IsShanghai: true,
		IsCancun:   true,
		IsPrague:   true,
		IsOsaka:    true,
		IsVerkle:   true,
	}
	err := enc.Encode(want)
	require.NoError(t, err)
	encodedSize := bigendian.Uint32ToBytes(uint32(buf.Len()))
	// Append size of rules and rules
	data := append(encodedSize, buf.Bytes()...)
	// reset the buffer and add the data in correct order
	buf.Reset()
	buf = bytes.NewBuffer(data)
	fr := &fileReader{
		reader: buf,
	}
	got, err := fr.ReadRules()
	require.NoError(t, err)
	require.Equal(t, want, got)
}

func TestFileReader_ReadAccessList(t *testing.T) {
	buf := &bytes.Buffer{}
	enc := gob.NewEncoder(buf)
	gob.Register(types.AccessList{})
	want := types.AccessList{
		{
			Address: common.Address{0x1},
			StorageKeys: []common.Hash{
				{0x2},
			},
		},
		{
			Address: common.Address{0x2},
			StorageKeys: []common.Hash{
				{0x3},
			},
		},
	}
	err := enc.Encode(want)
	require.NoError(t, err)
	encodedSize := bigendian.Uint32ToBytes(uint32(buf.Len()))
	// Append size of rules and rules
	data := append(encodedSize, buf.Bytes()...)
	// reset the buffer and add the data in correct order
	buf.Reset()
	buf = bytes.NewBuffer(data)
	fr := &fileReader{
		reader: buf,
	}
	got, err := fr.ReadAccessList()
	require.NoError(t, err)
	require.Equal(t, want, got)
}

func TestFileReader_ReadWorldState(t *testing.T) {
	buf := &bytes.Buffer{}
	enc := gob.NewEncoder(buf)
	gob.Register(txcontext.NewNilAccount())
	want := txcontext.NewWorldState(map[common.Address]txcontext.Account{
		common.Address{0x1}: txcontext.NewAccount(
			[]byte{0x22},
			map[common.Hash]common.Hash{{0x1}: {0x3}},
			big.NewInt(22),
			12,
		),
	})
	err := enc.Encode(want)
	require.NoError(t, err)
	encodedSize := bigendian.Uint32ToBytes(uint32(buf.Len()))
	// Append size of rules and rules
	data := append(encodedSize, buf.Bytes()...)
	// reset the buffer and add the data in correct order
	buf.Reset()
	buf = bytes.NewBuffer(data)
	fr := &fileReader{
		reader: buf,
	}
	got, err := fr.ReadWorldState()
	require.NoError(t, err)
	require.Equal(t, want, got)
}

func TestFileReader_ReadLog(t *testing.T) {
	buf := &bytes.Buffer{}
	enc := gob.NewEncoder(buf)
	gob.Register(types.Log{})
	want := &types.Log{
		Address:        common.Address{0x2},
		BlockNumber:    12,
		TxHash:         common.Hash{0x3},
		TxIndex:        13,
		BlockHash:      common.Hash{0x4},
		BlockTimestamp: 14,
		Index:          15,
		Removed:        true,
	}
	err := enc.Encode(want)
	require.NoError(t, err)
	encodedSize := bigendian.Uint32ToBytes(uint32(buf.Len()))
	// Append size of rules and rules
	data := append(encodedSize, buf.Bytes()...)
	// reset the buffer and add the data in correct order
	buf.Reset()
	buf = bytes.NewBuffer(data)
	fr := &fileReader{
		reader: buf,
	}
	got, err := fr.ReadLog()
	require.NoError(t, err)
	require.Equal(t, want, got)
}
