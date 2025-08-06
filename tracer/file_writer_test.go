package tracer

import (
	"errors"
	"os"
	"testing"

	"github.com/Fantom-foundation/lachesis-base/common/bigendian"
	"github.com/klauspost/compress/gzip"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestNewFileWriter(t *testing.T) {
	fp := t.TempDir() + "test_record.gz"
	fw, err := NewFileWriter(fp)
	assert.NoError(t, err)
	assert.NotNil(t, fw)
	_, ok := fw.(*fileWriter)
	assert.True(t, ok)
	require.NoError(t, fw.Close())
	// file exists - factory func should fail
	_, err = NewFileWriter(fp)
	assert.ErrorContains(t, err, "already exists")
}

func TestFileWriter_WritesDataIntoFile(t *testing.T) {
	fp := t.TempDir() + "test_record.gz"
	fw, err := NewFileWriter(fp)
	assert.NoError(t, err)
	err = fw.WriteData([]byte("hello world"))
	assert.NoError(t, err)
	err = fw.Close()
	assert.NoError(t, err)
	file, err := os.Open(fp)
	assert.NoError(t, err)
	stats, err := file.Stat()
	assert.NoError(t, err)
	assert.NotZero(t, stats.Size())
}

func createNewFileWriter(t *testing.T, buffer *MockWriteBuffer, filepath string) *fileWriter {
	file, err := os.Create(filepath)
	assert.NoError(t, err)

	return &fileWriter{
		buffer: buffer,
		closer: gzip.NewWriter(file),
	}
}

func TestFileWriter_WriteData(t *testing.T) {
	fp := t.TempDir() + "test_record.gz"
	data := []byte("hello world")
	mockErr := errors.New("mock error")
	tests := []struct {
		name    string
		setup   func(*MockWriteBuffer)
		wantErr error
	}{
		{
			name: "Success",
			setup: func(m *MockWriteBuffer) {
				m.EXPECT().Write(data).Return(len(data), nil)
			},
			wantErr: nil,
		},
		{
			name: "WriteError",
			setup: func(m *MockWriteBuffer) {
				m.EXPECT().Write(data).Return(0, mockErr)
			},
			wantErr: mockErr,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			buffer := NewMockWriteBuffer(ctrl)
			test.setup(buffer)

			fw := createNewFileWriter(t, buffer, fp)
			err := fw.WriteData(data)
			if test.wantErr != nil {
				assert.ErrorIs(t, err, test.wantErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestFileWriter_WriteUint16(t *testing.T) {
	fp := t.TempDir() + "test_record.gz"
	data := uint16(10)
	mockErr := errors.New("mock error")
	tests := []struct {
		name    string
		setup   func(*MockWriteBuffer)
		wantErr error
	}{
		{
			name: "Success",
			setup: func(m *MockWriteBuffer) {
				m.EXPECT().Write(bigendian.Uint16ToBytes(data)).Return(len(bigendian.Uint16ToBytes(data)), nil)
			},
			wantErr: nil,
		},
		{
			name: "WriteError",
			setup: func(m *MockWriteBuffer) {
				m.EXPECT().Write(bigendian.Uint16ToBytes(data)).Return(0, mockErr)
			},
			wantErr: mockErr,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			buffer := NewMockWriteBuffer(ctrl)
			test.setup(buffer)

			fw := createNewFileWriter(t, buffer, fp)
			err := fw.WriteUint16(data)
			if test.wantErr != nil {
				assert.ErrorIs(t, err, test.wantErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestFileWriter_WriteUint8(t *testing.T) {
	fp := t.TempDir() + "test_record.gz"
	data := uint8(11)
	mockErr := errors.New("mock error")
	tests := []struct {
		name    string
		setup   func(*MockWriteBuffer)
		wantErr error
	}{
		{
			name: "Success",
			setup: func(m *MockWriteBuffer) {
				m.EXPECT().WriteByte(data).Return(nil)
			},
			wantErr: nil,
		},
		{
			name: "WriteError",
			setup: func(m *MockWriteBuffer) {
				m.EXPECT().WriteByte(data).Return(mockErr)
			},
			wantErr: mockErr,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			buffer := NewMockWriteBuffer(ctrl)
			test.setup(buffer)

			fw := createNewFileWriter(t, buffer, fp)
			err := fw.WriteUint8(data)
			if test.wantErr != nil {
				assert.ErrorIs(t, err, test.wantErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
