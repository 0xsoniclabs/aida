package tracer

import (
	"compress/gzip"
	"github.com/Fantom-foundation/lachesis-base/common/bigendian"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"os"
	"testing"
)

func TestNewFileHandler(t *testing.T) {
	fh, err := NewFileHandler(t.TempDir())
	assert.NoError(t, err)
	assert.NotNil(t, fh)
	_, ok := fh.(*fileHandler)
	assert.True(t, ok)
}

func TestFileHandler_WritesDataIntoFile(t *testing.T) {
	fp := t.TempDir()
	fh, err := NewFileHandler(fp)
	assert.NoError(t, err)
	fh.WriteData([]byte("hello world"))
	err = fh.Close()
	assert.NoError(t, err)
	file, err := os.Open(fp)
	assert.NoError(t, err)
	stats, err := file.Stat()
	assert.NoError(t, err)
	assert.NotZero(t, stats.Size())
}

func createNewFileHandler(t *testing.T, buffer *MockBuffer, filepath string) *fileHandler {
	file, err := os.Create(filepath)
	assert.NoError(t, err)

	return &fileHandler{
		file:   gzip.NewWriter(file),
		buffer: buffer,
	}
}

func TestFileHandler_WriteData(t *testing.T) {
	ctrl := gomock.NewController(t)
	buffer := NewMockBuffer(ctrl)
	fp := t.TempDir()
	fh := createNewFileHandler(t, buffer, fp)
	data := []byte("hello world")
	buffer.EXPECT().Write(data)
	fh.WriteData(data)
}

func TestFileHandler_WriteUint16(t *testing.T) {
	ctrl := gomock.NewController(t)
	buffer := NewMockBuffer(ctrl)
	fp := t.TempDir()
	fh := createNewFileHandler(t, buffer, fp)
	data := uint16(1234)
	buffer.EXPECT().WriteByte(bigendian.Uint16ToBytes(data))
	fh.WriteUint16(data)
}

func TestFileHandler_WriteUint8(t *testing.T) {
	ctrl := gomock.NewController(t)
	buffer := NewMockBuffer(ctrl)
	fp := t.TempDir()
	fh := createNewFileHandler(t, buffer, fp)
	data := uint8(11)
	buffer.EXPECT().WriteByte(data)
	fh.WriteUint8(data)
}
