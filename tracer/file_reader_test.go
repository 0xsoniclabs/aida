package tracer

import (
	"errors"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
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
	_, err = file.Write([]byte("test data for file reader"))
	require.NoError(t, err)
	require.NoError(t, file.Close())

	reader, err := NewFileReader(tempFile)
	require.NoError(t, err)
	require.NotNil(t, reader)
	_, ok := reader.(*fileReader)
	require.True(t, ok)

	// Ensure the reader can be closed without error
	err = reader.Close()
	require.NoError(t, err)
}

func TestFileReader_ReadData(t *testing.T) {
	ctrl := gomock.NewController(t)
	var buf *MockReadBuffer
	mockErr := errors.New("mock error")

	tests := []struct {
		name    string
		wantErr error
		setup   func()
	}{
		{
			name:    "Success",
			wantErr: nil,
			setup: func() {
				buf = NewMockReadBuffer(ctrl)
				buf.EXPECT().Read(gomock.Any()).MinTimes(1).Return(1, nil)
			},
		},
		{
			name:    "Error",
			wantErr: mockErr,
			setup: func() {
				buf = NewMockReadBuffer(ctrl)
				buf.EXPECT().Read(gomock.Any()).MinTimes(1).Return(0, mockErr)
			},
		},
	}
	for _, test := range tests {
		test.setup()
		fr := &fileReader{
			reader: buf,
		}
		data, err := fr.ReadData(1)
		if err != nil {
			require.ErrorIs(t, err, test.wantErr)
		} else {
			require.Len(t, data, 1)
		}
	}
}

func TestFileReader_ReadUint16(t *testing.T) {
	ctrl := gomock.NewController(t)
	var buf *MockReadBuffer
	mockErr := errors.New("mock error")

	tests := []struct {
		name    string
		wantErr error
		setup   func()
	}{
		{
			name:    "Success",
			wantErr: nil,
			setup: func() {
				buf = NewMockReadBuffer(ctrl)
				buf.EXPECT().Read(gomock.Any()).MinTimes(1).Return(1, nil)
			},
		},
		{
			name:    "Error",
			wantErr: mockErr,
			setup: func() {
				buf = NewMockReadBuffer(ctrl)
				buf.EXPECT().Read(gomock.Any()).MinTimes(1).Return(0, mockErr)
			},
		},
	}
	for _, test := range tests {
		test.setup()
		fr := &fileReader{
			reader: buf,
		}
		data, err := fr.ReadUint16()
		if err != nil {
			require.ErrorIs(t, err, test.wantErr)
		} else {
			require.Len(t, data, 2)
		}
	}
}

func TestFileReader_ReadUint8(t *testing.T) {
	ctrl := gomock.NewController(t)
	var buf *MockReadBuffer
	mockErr := errors.New("mock error")

	tests := []struct {
		name    string
		wantErr error
		setup   func()
	}{
		{
			name:    "Success",
			wantErr: nil,
			setup: func() {
				buf = NewMockReadBuffer(ctrl)
				buf.EXPECT().Read(gomock.Any()).MinTimes(1).Return(1, nil)
			},
		},
		{
			name:    "Error",
			wantErr: mockErr,
			setup: func() {
				buf = NewMockReadBuffer(ctrl)
				buf.EXPECT().Read(gomock.Any()).MinTimes(1).Return(0, mockErr)
			},
		},
	}
	for _, test := range tests {
		test.setup()
		fr := &fileReader{
			reader: buf,
		}
		data, err := fr.ReadUint16()
		if err != nil {
			require.ErrorIs(t, err, test.wantErr)
		} else {
			require.Len(t, data, 1)
		}
	}
}
