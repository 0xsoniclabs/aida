package rpc

import (
	"bytes"
	"context"
	"io"
	"testing"

	"go.uber.org/mock/gomock"

	"github.com/stretchr/testify/assert"
)

// Package-level function tests
func TestRpc_newIterator(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockContext := context.TODO()
	mockReadCloser := NewMockProxyIoReadCloser(ctrl)
	mockReadCloser.EXPECT().Read(gomock.Any()).Return(0, io.EOF).AnyTimes()
	out := newIterator(mockContext, mockReadCloser, 10)
	assert.NotNil(t, out)
	assert.Equal(t, mockContext, out.ctx)
	assert.Equal(t, mockReadCloser, out.in)
	assert.NotNil(t, out.closed)
	assert.NotNil(t, out.out)
}

// Iterator struct method tests
func TestIterator_Next(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockContext := context.TODO()
	mockReadCloser := NewMockProxyIoReadCloser(ctrl)
	mockReadCloser.EXPECT().Read(gomock.Any()).Return(0, io.EOF).AnyTimes()
	iter := newIterator(mockContext, mockReadCloser, 1)
	out := iter.Next()
	assert.False(t, out) // Expecting false since Read will return io.EOF
}

func TestIterator_Close(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockContext := context.TODO()
	mockReadCloser := NewMockProxyIoReadCloser(ctrl)
	mockReadCloser.EXPECT().Read(gomock.Any()).Return(0, io.EOF)
	mockReadCloser.EXPECT().Close().Return(nil)
	iter := newIterator(mockContext, mockReadCloser, 1)
	iter.Close()
}

func TestIterator_Value(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockItem := &RequestAndResults{}
	iter := &iterator{
		item: mockItem,
	}
	out := iter.Value()
	assert.Equal(t, mockItem, out)
}

func TestIterator_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockErr := io.ErrUnexpectedEOF
	iter := &iterator{
		err: mockErr,
	}
	out := iter.Error()
	assert.Equal(t, mockErr, out)
}

func TestIterator_read(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	t.Run("success", func(t *testing.T) {
		// Create original header
		h := &Header{}
		paramBytes := []byte(`["0x1234567890abcdef", "latest"]`)
		h.version = 1
		h.namespace = namespaceDictionary["eth"]
		h.method = methodDictionary[h.namespace]["call"]
		h.isError = false
		h.isLongQuery = false
		h.isLongResult = false
		h.querySize = int32(len(paramBytes))
		h.resultCodeSize = 0
		h.blockID = 12345
		h.blockTimestamp = 1640995200

		var buf bytes.Buffer
		_, err := h.WriteTo(&buf)
		assert.NoError(t, err)

		payloadBytes := append(buf.Bytes(), paramBytes...)
		reader := io.NopCloser(bytes.NewReader(payloadBytes))
		iter := &iterator{
			in: reader,
		}

		out, err := iter.read()
		assert.NotNil(t, out)
		assert.NoError(t, err)
	})

	t.Run("error method", func(t *testing.T) {
		// Create original header
		h := &Header{}
		paramBytes := []byte(`["0x1234567890abcdef", "latest"]`)
		h.version = 1
		h.namespace = namespaceDictionary["eth"]
		h.method = methodDictionary[h.namespace]["xyz"]
		h.isError = false
		h.isLongQuery = false
		h.isLongResult = false
		h.querySize = int32(len(paramBytes))
		h.resultCodeSize = 0
		h.blockID = 12345
		h.blockTimestamp = 1640995200

		var buf bytes.Buffer
		_, err := h.WriteTo(&buf)
		assert.NoError(t, err)

		payloadBytes := append(buf.Bytes(), paramBytes...)
		reader := io.NopCloser(bytes.NewReader(payloadBytes))
		iter := &iterator{
			in: reader,
		}

		out, err := iter.read()
		assert.Nil(t, out)
		assert.Error(t, err)
	})

	t.Run("error namespace", func(t *testing.T) {
		// Create original header
		h := &Header{}
		paramBytes := []byte(`["0x1234567890abcdef", "latest"]`)
		h.version = 1
		h.namespace = 0
		h.method = methodDictionary[namespaceDictionary["eth"]]["call"]
		h.isError = false
		h.isLongQuery = false
		h.isLongResult = false
		h.querySize = int32(len(paramBytes))
		h.resultCodeSize = 0
		h.blockID = 12345
		h.blockTimestamp = 1640995200

		var buf bytes.Buffer
		_, err := h.WriteTo(&buf)
		assert.NoError(t, err)

		payloadBytes := append(buf.Bytes(), paramBytes...)
		reader := io.NopCloser(bytes.NewReader(payloadBytes))
		iter := &iterator{
			in: reader,
		}

		out, err := iter.read()
		assert.Nil(t, out)
		assert.Error(t, err)
	})

}

func TestIterator_decode(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	t.Run("success with response", func(t *testing.T) {
		// Create original header
		h := &Header{}
		paramBytes := []byte(`["0x1234567890abcdef", "latest"]`)
		responseBytes := []byte(`["0x1234567890abcdef", "latest"]`)
		h.version = 1
		h.namespace = namespaceDictionary["eth"]
		h.method = methodDictionary[h.namespace]["call"]
		h.isError = false
		h.isLongQuery = false
		h.isLongResult = false
		h.querySize = int32(len(paramBytes))
		h.resultCodeSize = int32(len(responseBytes))
		h.blockID = 12345
		h.blockTimestamp = 1640995200

		payloadBytes := append(paramBytes, responseBytes...)
		reader := io.NopCloser(bytes.NewReader(payloadBytes))
		iter := &iterator{
			in: reader,
		}
		out, err := iter.decode(h, "eth", "call")
		assert.NotNil(t, out)
		assert.NoError(t, err)
	})

	t.Run("success with error response", func(t *testing.T) {
		// Create original header
		h := &Header{}
		paramBytes := []byte(`["0x1234567890abcdef", "latest"]`)
		errorBytes := []byte(`["0x1234567890abcdef", "latest"]`)
		h.version = 1
		h.namespace = namespaceDictionary["eth"]
		h.method = methodDictionary[h.namespace]["call"]
		h.isError = true
		h.isLongQuery = false
		h.isLongResult = false
		h.querySize = int32(len(paramBytes))
		h.blockID = 12345
		h.blockTimestamp = 1640995200

		payloadBytes := append(paramBytes, errorBytes...)
		reader := io.NopCloser(bytes.NewReader(payloadBytes))
		iter := &iterator{
			in: reader,
		}
		out, err := iter.decode(h, "eth", "call")
		assert.NotNil(t, out)
		assert.NoError(t, err)
	})

	t.Run("error read param", func(t *testing.T) {
		// Create original header
		h := &Header{}
		h.querySize = 1000
		reader := io.NopCloser(bytes.NewReader([]byte("test payload")))
		iter := &iterator{
			in: reader,
		}
		out, err := iter.decode(h, "eth", "call")
		assert.Nil(t, out)
		assert.Error(t, err)
	})

	t.Run("error parse param", func(t *testing.T) {
		// Create original header
		h := &Header{}
		reader := io.NopCloser(bytes.NewReader([]byte("test payload")))
		iter := &iterator{
			in: reader,
		}
		out, err := iter.decode(h, "eth", "call")
		assert.Nil(t, out)
		assert.Error(t, err)
	})

	t.Run("error read response", func(t *testing.T) {
		// Create original header
		h := &Header{}
		paramBytes := []byte(`["0x1234567890abcdef", "latest"]`)
		h.version = 1
		h.namespace = namespaceDictionary["eth"]
		h.method = methodDictionary[h.namespace]["call"]
		h.isError = false
		h.isLongQuery = false
		h.isLongResult = false
		h.querySize = int32(len(paramBytes))
		h.resultCodeSize = 1000
		h.blockID = 12345
		h.blockTimestamp = 1640995200

		reader := io.NopCloser(bytes.NewReader(paramBytes))
		iter := &iterator{
			in: reader,
		}
		out, err := iter.decode(h, "eth", "call")
		assert.Nil(t, out)
		assert.Error(t, err)
	})

}

func TestIterator_loadPayload(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	t.Run("success", func(t *testing.T) {
		mockReadCloser := NewMockProxyIoReadCloser(ctrl)
		mockReadCloser.EXPECT().Read(gomock.Any()).Return(2, nil)
		iter := &iterator{
			in: mockReadCloser,
		}

		err := iter.loadPayload([]byte{0x1, 0x2})
		assert.NoError(t, err)
	})

	t.Run("error", func(t *testing.T) {
		mockReadCloser := NewMockProxyIoReadCloser(ctrl)
		mockReadCloser.EXPECT().Read(gomock.Any()).Return(0, io.ErrUnexpectedEOF)
		iter := &iterator{
			in: mockReadCloser,
		}

		err := iter.loadPayload([]byte{0x1, 0x2})
		assert.Error(t, err)
		assert.Equal(t, io.ErrUnexpectedEOF, err)
	})

}
