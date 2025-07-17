package tracer

import (
	"compress/gzip"
	"github.com/Fantom-foundation/lachesis-base/common/bigendian"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"os"
	"testing"
)

func TestNewFileWriter(t *testing.T) {
	fp := t.TempDir() + "test_record.gz"
	fw, err := NewFileWriter(fp)
	assert.NoError(t, err)
	assert.NotNil(t, fw)
	_, ok := fw.(*fileWriter)
	assert.True(t, ok)
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

func createNewFileWriter(t *testing.T, buffer *MockBuffer, filepath string) *fileWriter {
	file, err := os.Create(filepath)
	assert.NoError(t, err)

	return &fileWriter{
		file:   gzip.NewWriter(file),
		buffer: buffer,
	}
}

func TestFileWriter_WriteData(t *testing.T) {
	ctrl := gomock.NewController(t)
	buffer := NewMockBuffer(ctrl)
	fp := t.TempDir() + "test_record.gz"
	fw := createNewFileWriter(t, buffer, fp)
	data := []byte("hello world")
	buffer.EXPECT().Write(data)
	err := fw.WriteData(data)
	assert.NoError(t, err)
}

func TestFileWriter_WriteUint16(t *testing.T) {
	ctrl := gomock.NewController(t)
	buffer := NewMockBuffer(ctrl)
	fp := t.TempDir() + "test_record.gz"
	fw := createNewFileWriter(t, buffer, fp)
	data := uint16(1234)
	buffer.EXPECT().Write(bigendian.Uint16ToBytes(data))
	err := fw.WriteUint16(data)
	assert.NoError(t, err)
}

func TestFileWriter_WriteUint8(t *testing.T) {
	ctrl := gomock.NewController(t)
	buffer := NewMockBuffer(ctrl)
	fp := t.TempDir() + "test_record.gz"
	fw := createNewFileWriter(t, buffer, fp)
	data := uint8(11)
	buffer.EXPECT().WriteByte(data)
	err := fw.WriteUint8(data)
	assert.NoError(t, err)
}

func TestReplay_EncodeDecode(t *testing.T) {
	data := bigendian.Uint64ToBytes(64)

	argOp, err := EncodeArgOp(AddRefundID, NoArgID, NoArgID, NoArgID)
	assert.NoError(t, err)
	fh, err := NewFileWriter(t.TempDir())
	assert.NoError(t, err)
	err = fh.WriteUint16(argOp)
	assert.NoError(t, err)
	err = fh.WriteData(data)
	assert.NoError(t, err)
	decodedOp, decodedData, err := DecodeArgOp(fh)
}
