package rpc

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRpc_CanRecord(t *testing.T) {
	tests := []struct {
		name      string
		namespace string
		method    string
		expected  bool
	}{
		{
			name:      "valid eth namespace with call method",
			namespace: "eth",
			method:    "call",
			expected:  true,
		},
		{
			name:      "valid ftm namespace with getBalance method",
			namespace: "ftm",
			method:    "getBalance",
			expected:  true,
		},
		{
			name:      "invalid namespace",
			namespace: "invalid",
			method:    "call",
			expected:  false,
		},
		{
			name:      "valid namespace with invalid method",
			namespace: "eth",
			method:    "invalidMethod",
			expected:  false,
		},
		{
			name:      "empty namespace",
			namespace: "",
			method:    "call",
			expected:  false,
		},
		{
			name:      "empty method",
			namespace: "eth",
			method:    "",
			expected:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out := CanRecord(tt.namespace, tt.method)
			assert.Equal(t, tt.expected, out)
		})
	}
}

func TestHeader_SetMethod(t *testing.T) {
	tests := []struct {
		name      string
		namespace string
		method    string
		wantErr   bool
		errMsg    string
	}{
		{
			name:      "valid eth namespace and call method",
			namespace: "eth",
			method:    "call",
			wantErr:   false,
		},
		{
			name:      "valid ftm namespace and getBalance method",
			namespace: "ftm",
			method:    "getBalance",
			wantErr:   false,
		},
		{
			name:      "invalid namespace",
			namespace: "invalid",
			method:    "call",
			wantErr:   true,
			errMsg:    "namespace 'invalid' not recorded",
		},
		{
			name:      "valid namespace but invalid method",
			namespace: "eth",
			method:    "invalidMethod",
			wantErr:   true,
			errMsg:    "method 'invalidMethod' of namespace 'eth' not recorded",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &Header{}
			err := h.SetMethod(tt.namespace, tt.method)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
				decodeNamespace := namespaceDictionary[tt.namespace]
				decodeMethod := methodDictionary[decodeNamespace][tt.method]
				assert.Equal(t, decodeNamespace, h.namespace)
				assert.Equal(t, decodeMethod, h.method)
			}
		})
	}
}

func TestHeader_Namespace(t *testing.T) {

	t.Run("valid namespace eth", func(t *testing.T) {
		h := &Header{
			namespace: namespaceDictionary["eth"],
		}
		ns, err := h.Namespace()
		assert.NoError(t, err)
		assert.Equal(t, "eth", ns)
	})

	t.Run("valid namespace ftm", func(t *testing.T) {
		h := &Header{
			namespace: namespaceDictionary["ftm"],
		}
		ns, err := h.Namespace()
		assert.NoError(t, err)
		assert.Equal(t, "eth", ns)
	})

	t.Run("namespace not initialized", func(t *testing.T) {
		h := &Header{}
		ns, err := h.Namespace()
		assert.NotNil(t, err)
		assert.Contains(t, err.Error(), "namespace not initialized")
		assert.Empty(t, ns)
	})

	t.Run("invalid namespace", func(t *testing.T) {
		h := &Header{
			namespace: 99,
		}
		ns, err := h.Namespace()
		assert.NotNil(t, err)
		assert.Contains(t, err.Error(), "unknown namespace set")
		assert.Empty(t, ns)
	})

}

func TestHeader_Method(t *testing.T) {
	tests := []struct {
		name      string
		setupFunc func(*Header)
		expected  string
		wantErr   bool
		errMsg    string
	}{
		{
			name: "valid method call",
			setupFunc: func(h *Header) {
				h.namespace = namespaceDictionary["eth"]
				h.method = methodDictionary[h.namespace]["call"]
			},
			expected: "call",
			wantErr:  false,
		},
		{
			name: "valid method getBalance",
			setupFunc: func(h *Header) {
				h.namespace = namespaceDictionary["eth"]
				h.method = methodDictionary[h.namespace]["getBalance"]
			},
			expected: "getBalance",
			wantErr:  false,
		},
		{
			name: "uninitialized namespace and method",
			setupFunc: func(h *Header) {
				// Don't set anything
			},
			expected: "",
			wantErr:  true,
			errMsg:   "namespace or method not initialized",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &Header{}
			tt.setupFunc(h)

			out, err := h.Method()

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, out)
			}
		})
	}
}

func TestHeader_SetBlockID(t *testing.T) {
	tests := []struct {
		name     string
		blockID  uint64
		expected uint64
	}{
		{
			name:     "zero block ID",
			blockID:  0,
			expected: 0,
		},
		{
			name:     "small block ID",
			blockID:  123,
			expected: 123,
		},
		{
			name:     "large block ID",
			blockID:  18446744073709551615, // max uint64
			expected: 18446744073709551615,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &Header{}
			h.SetBlockID(tt.blockID)
			assert.Equal(t, tt.expected, h.BlockID())
		})
	}
}

func TestHeader_BlockID(t *testing.T) {
	h := &Header{}

	// Test default value
	assert.Equal(t, uint64(0), h.BlockID())

	// Test after setting
	h.SetBlockID(12345)
	assert.Equal(t, uint64(12345), h.BlockID())
}

func TestHeader_SetBlockTimestamp(t *testing.T) {
	tests := []struct {
		name      string
		timestamp uint64
		expected  uint64
	}{
		{
			name:      "zero timestamp",
			timestamp: 0,
			expected:  0,
		},
		{
			name:      "unix timestamp",
			timestamp: 1640995200, // 2022-01-01 00:00:00 UTC
			expected:  1640995200,
		},
		{
			name:      "large timestamp",
			timestamp: 18446744073709551615, // max uint64
			expected:  18446744073709551615,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &Header{}
			h.SetBlockTimestamp(tt.timestamp)
			assert.Equal(t, tt.expected, h.BlockTimestamp())
		})
	}
}

func TestHeader_BlockTimestamp(t *testing.T) {
	h := &Header{}

	// Test default value
	assert.Equal(t, uint64(0), h.BlockTimestamp())

	// Test after setting
	h.SetBlockTimestamp(1640995200)
	assert.Equal(t, uint64(1640995200), h.BlockTimestamp())
}

func TestHeader_SetQueryLength(t *testing.T) {
	tests := []struct {
		name        string
		queryLength int
		wantErr     bool
		errMsg      string
	}{
		{
			name:        "zero length",
			queryLength: 0,
			wantErr:     false,
		},
		{
			name:        "small query length",
			queryLength: 100,
			wantErr:     false,
		},
		{
			name:        "max short query length",
			queryLength: maxShortQuery,
			wantErr:     false,
		},
		{
			name:        "long query length",
			queryLength: maxShortQuery + 1,
			wantErr:     false,
		},
		{
			name:        "max allowed query length",
			queryLength: maxQuerySizeAllowed,
			wantErr:     false,
		},
		{
			name:        "query too large",
			queryLength: maxQuerySizeAllowed + 1,
			wantErr:     true,
			errMsg:      "query too big",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &Header{}
			err := h.SetQueryLength(tt.queryLength)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, int32(tt.queryLength), h.QueryLength())

				expectedLongQuery := tt.queryLength > maxShortQuery
				assert.Equal(t, expectedLongQuery, h.isLongQuery)
			}
		})
	}
}

func TestHeader_QueryLength(t *testing.T) {
	h := &Header{}

	// Test default value
	assert.Equal(t, int32(0), h.QueryLength())

	// Test after setting
	err := h.SetQueryLength(500)
	assert.NoError(t, err)
	assert.Equal(t, int32(500), h.QueryLength())
}

func TestHeader_SetError(t *testing.T) {
	tests := []struct {
		name     string
		errCode  int
		expected int
	}{
		{
			name:     "error code zero",
			errCode:  0,
			expected: 0,
		},
		{
			name:     "positive error code",
			errCode:  404,
			expected: 404,
		},
		{
			name:     "negative error code",
			errCode:  -1,
			expected: -1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &Header{}
			h.SetError(tt.errCode)

			assert.True(t, h.IsError())
			assert.Equal(t, tt.expected, h.ErrorCode())
			assert.Equal(t, int32(0), h.ResponseLength())
			assert.False(t, h.isLongResult)
		})
	}
}

func TestHeader_ErrorCode(t *testing.T) {
	h := &Header{}

	// Test default value (no error)
	assert.Equal(t, 0, h.ErrorCode())
	assert.False(t, h.IsError())

	// Test after setting error
	h.SetError(500)
	assert.Equal(t, 500, h.ErrorCode())
	assert.True(t, h.IsError())

	// Test after setting response (clears error)
	h.SetResponseLength(100)
	assert.Equal(t, 0, h.ErrorCode())
	assert.False(t, h.IsError())
}

func TestHeader_IsError(t *testing.T) {
	h := &Header{}

	// Test default value
	assert.False(t, h.IsError())

	// Test after setting error
	h.SetError(404)
	assert.True(t, h.IsError())

	// Test after setting response (should clear error)
	h.SetResponseLength(200)
	assert.False(t, h.IsError())
}

func TestHeader_SetResponseLength(t *testing.T) {
	tests := []struct {
		name           string
		responseLength int
		expectedLong   bool
	}{
		{
			name:           "zero length",
			responseLength: 0,
			expectedLong:   false,
		},
		{
			name:           "small response",
			responseLength: 100,
			expectedLong:   false,
		},
		{
			name:           "max short response",
			responseLength: maxShortResponse,
			expectedLong:   false,
		},
		{
			name:           "long response",
			responseLength: maxShortResponse + 1,
			expectedLong:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &Header{}
			h.SetResponseLength(tt.responseLength)

			assert.False(t, h.IsError()) // Should clear error flag
			assert.Equal(t, int32(tt.responseLength), h.ResponseLength())
			assert.Equal(t, tt.expectedLong, h.isLongResult)
		})
	}
}

func TestHeader_ResponseLength(t *testing.T) {
	h := &Header{}

	// Test default value
	assert.Equal(t, int32(0), h.ResponseLength())

	// Test after setting response
	h.SetResponseLength(1024)
	assert.Equal(t, int32(1024), h.ResponseLength())

	// Test when error is set (should return 0)
	h.SetError(500)
	assert.Equal(t, int32(0), h.ResponseLength())
}

func TestHeader_WriteTo(t *testing.T) {
	tests := []struct {
		name      string
		setupFunc func(*Header)
		minBytes  int64
	}{
		{
			name: "basic header with short query and response",
			setupFunc: func(h *Header) {
				h.version = 1
				h.namespace = namespaceDictionary["eth"]
				h.method = methodDictionary[h.namespace]["call"]
				h.isError = false
				h.isLongQuery = false
				h.isLongResult = false
				h.querySize = 100
				h.resultCodeSize = 200
				h.blockID = 12345
				h.blockTimestamp = 1640995200
			},
			minBytes: 18,
		},
		{
			name: "header with long query",
			setupFunc: func(h *Header) {
				h.version = 1
				h.namespace = namespaceDictionary["eth"]
				h.method = methodDictionary[h.namespace]["getBalance"]
				h.isError = false
				h.isLongQuery = true
				h.isLongResult = false
				h.querySize = maxShortQuery + 100
				h.resultCodeSize = 50
				h.blockID = 67890
				h.blockTimestamp = 1640995300
			},
			minBytes: 19,
		},
		{
			name: "header with error",
			setupFunc: func(h *Header) {
				h.version = 1
				h.namespace = namespaceDictionary["ftm"]
				h.method = methodDictionary[h.namespace]["call"]
				h.isError = true
				h.isLongQuery = false
				h.isLongResult = false
				h.querySize = 50
				h.resultCodeSize = 404
				h.blockID = 11111
				h.blockTimestamp = 1640995400
			},
			minBytes: 18,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &Header{}
			tt.setupFunc(h)

			var buf bytes.Buffer
			n, err := h.WriteTo(&buf)

			assert.NoError(t, err)
			assert.GreaterOrEqual(t, n, tt.minBytes)
			assert.Equal(t, n, int64(buf.Len()))
			assert.Greater(t, buf.Len(), 0)
		})
	}
}

func TestHeader_ReadFrom(t *testing.T) {
	tests := []struct {
		name      string
		setupFunc func(*Header)
	}{
		{
			name: "round trip with basic header",
			setupFunc: func(h *Header) {
				h.version = 1
				h.namespace = namespaceDictionary["eth"]
				h.method = methodDictionary[h.namespace]["call"]
				h.isError = false
				h.isLongQuery = false
				h.isLongResult = false
				h.querySize = 100
				h.resultCodeSize = 200
				h.blockID = 12345
				h.blockTimestamp = 1640995200
			},
		},
		{
			name: "round trip with long query",
			setupFunc: func(h *Header) {
				h.version = 1
				h.namespace = namespaceDictionary["ftm"]
				h.method = methodDictionary[h.namespace]["getBalance"]
				h.isError = false
				h.isLongQuery = true
				h.isLongResult = false
				h.querySize = maxShortQuery + 100
				h.resultCodeSize = 50
				h.blockID = 67890
				h.blockTimestamp = 1640995300
			},
		},
		{
			name: "round trip with error",
			setupFunc: func(h *Header) {
				h.version = 1
				h.namespace = namespaceDictionary["eth"]
				h.method = methodDictionary[h.namespace]["estimateGas"]
				h.isError = true
				h.isLongQuery = false
				h.isLongResult = false
				h.querySize = 75
				h.resultCodeSize = 500
				h.blockID = 98765
				h.blockTimestamp = 1640995500
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create original header
			original := &Header{}
			tt.setupFunc(original)

			// Write to buffer
			var buf bytes.Buffer
			_, err := original.WriteTo(&buf)
			assert.NoError(t, err)

			// Read back from buffer
			restored := &Header{}
			n, err := restored.ReadFrom(&buf)
			assert.NoError(t, err)
			assert.Greater(t, n, int64(0))

			// Compare values
			assert.Equal(t, original.BlockID(), restored.BlockID())
			assert.Equal(t, original.BlockTimestamp(), restored.BlockTimestamp())
			assert.Equal(t, original.QueryLength(), restored.QueryLength())
			assert.Equal(t, original.IsError(), restored.IsError())

			if original.IsError() {
				assert.Equal(t, original.ErrorCode(), restored.ErrorCode())
			} else {
				assert.Equal(t, original.ResponseLength(), restored.ResponseLength())
			}

			// Check namespace and method
			origNS, _ := original.Namespace()
			restoredNS, _ := restored.Namespace()
			assert.Equal(t, origNS, restoredNS)

			origMethod, _ := original.Method()
			restoredMethod, _ := restored.Method()
			assert.Equal(t, origMethod, restoredMethod)
		})
	}
}

func TestHeader_codeQuery(t *testing.T) {
	tests := []struct {
		name        string
		setupFunc   func(*Header)
		expectedLen int
	}{
		{
			name: "short query",
			setupFunc: func(h *Header) {
				h.version = 1
				h.namespace = namespaceDictionary["eth"]
				h.method = methodDictionary[h.namespace]["call"]
				h.isLongQuery = false
				h.querySize = 100
			},
			expectedLen: 3,
		},
		{
			name: "long query",
			setupFunc: func(h *Header) {
				h.version = 1
				h.namespace = namespaceDictionary["ftm"]
				h.method = methodDictionary[h.namespace]["getBalance"]
				h.isLongQuery = true
				h.querySize = maxShortQuery + 100
			},
			expectedLen: 4,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &Header{}
			tt.setupFunc(h)

			hdr := make([]byte, headerSize)
			length := h.codeQuery(hdr)

			assert.Equal(t, tt.expectedLen, length)
			assert.NotEqual(t, byte(0), hdr[0]) // Should have written something
		})
	}
}

func TestHeader_codeError(t *testing.T) {
	tests := []struct {
		name      string
		setupFunc func(*Header)
		errCode   int
		offset    int
		expected  int
	}{
		{
			name: "positive error code",
			setupFunc: func(h *Header) {
				h.isError = true
				h.resultCodeSize = 404
			},
			errCode:  404,
			offset:   3,
			expected: 2,
		},
		{
			name: "negative error code",
			setupFunc: func(h *Header) {
				h.isError = true
				h.resultCodeSize = -1
			},
			errCode:  -1,
			offset:   4,
			expected: 2,
		},
		{
			name: "zero error code",
			setupFunc: func(h *Header) {
				h.isError = true
				h.resultCodeSize = 0
			},
			errCode:  0,
			offset:   3,
			expected: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &Header{}
			tt.setupFunc(h)

			hdr := make([]byte, headerSize)
			length := h.codeError(hdr, tt.offset)

			assert.Equal(t, tt.expected, length)
			// Verify error flag is set
			assert.True(t, hdr[0]&(1<<7) > 0)
		})
	}
}

func TestHeader_codeResponse(t *testing.T) {
	tests := []struct {
		name           string
		setupFunc      func(*Header)
		responseLength int
		offset         int
		expectedLen    int
	}{
		{
			name: "short response",
			setupFunc: func(h *Header) {
				h.isError = false
				h.isLongResult = false
				h.resultCodeSize = 100
			},
			responseLength: 100,
			offset:         3,
			expectedLen:    2,
		},
		{
			name: "long response",
			setupFunc: func(h *Header) {
				h.isError = false
				h.isLongResult = true
				h.resultCodeSize = maxShortResponse + 100
			},
			responseLength: maxShortResponse + 100,
			offset:         4,
			expectedLen:    4,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &Header{}
			tt.setupFunc(h)

			hdr := make([]byte, headerSize)
			length := h.codeResponse(hdr, tt.offset)

			assert.Equal(t, tt.expectedLen, length)

			if tt.responseLength > maxShortResponse {
				assert.True(t, hdr[0]&(1<<5) > 0)
			}
		})
	}
}

func TestHeader_readFrom(t *testing.T) {
	tests := []struct {
		name      string
		setupFunc func(*Header) []byte
		wantErr   bool
	}{
		{
			name: "valid header data",
			setupFunc: func(h *Header) []byte {
				h.version = 1
				h.namespace = namespaceDictionary["eth"]
				h.method = methodDictionary[h.namespace]["call"]
				h.isError = false
				h.isLongQuery = false
				h.isLongResult = false
				h.querySize = 100
				h.resultCodeSize = 200
				h.blockID = 12345
				h.blockTimestamp = 1640995200

				var buf bytes.Buffer
				_, err := h.WriteTo(&buf)
				if err != nil {
					t.Fatalf("WriteTo failed: %v", err)
				}
				return buf.Bytes()
			},
			wantErr: false,
		},
		{
			name: "corrupted checksum",
			setupFunc: func(h *Header) []byte {
				h.version = 1
				h.namespace = namespaceDictionary["eth"]
				h.method = methodDictionary[h.namespace]["call"]
				h.isError = false
				h.isLongQuery = false
				h.isLongResult = false
				h.querySize = 100
				h.resultCodeSize = 200
				h.blockID = 12345
				h.blockTimestamp = 1640995200

				var buf bytes.Buffer
				_, err := h.WriteTo(&buf)
				if err != nil {
					t.Fatalf("WriteTo failed: %v", err)
				}
				data := buf.Bytes()
				// Corrupt the checksum
				data[len(data)-1] = 0xFF
				return data
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &Header{}
			data := tt.setupFunc(h)

			reader := bytes.NewReader(data)
			resultData, err := h.readFrom(reader)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, resultData)
			}
		})
	}
}

func TestHeader_decodeFields(t *testing.T) {
	tests := []struct {
		name      string
		setupFunc func(*Header)
	}{
		{
			name: "decode short query and response",
			setupFunc: func(h *Header) {
				h.version = 1
				h.namespace = namespaceDictionary["eth"]
				h.method = methodDictionary[h.namespace]["call"]
				h.isError = false
				h.isLongQuery = false
				h.isLongResult = false
				h.querySize = 100
				h.resultCodeSize = 200
				h.blockID = 12345
				h.blockTimestamp = 1640995200
			},
		},
		{
			name: "decode long query",
			setupFunc: func(h *Header) {
				h.version = 1
				h.namespace = namespaceDictionary["ftm"]
				h.method = methodDictionary[h.namespace]["getBalance"]
				h.isError = false
				h.isLongQuery = true
				h.isLongResult = false
				h.querySize = maxShortQuery + 100
				h.resultCodeSize = 50
				h.blockID = 67890
				h.blockTimestamp = 1640995300
			},
		},
		{
			name: "decode error response",
			setupFunc: func(h *Header) {
				h.version = 1
				h.namespace = namespaceDictionary["eth"]
				h.method = methodDictionary[h.namespace]["estimateGas"]
				h.isError = true
				h.isLongQuery = false
				h.isLongResult = false
				h.querySize = 75
				h.resultCodeSize = 404
				h.blockID = 98765
				h.blockTimestamp = 1640995500
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create and encode original header
			original := &Header{}
			tt.setupFunc(original)

			var buf bytes.Buffer
			_, err := original.WriteTo(&buf)
			if err != nil {
				t.Fatalf("WriteTo failed: %v", err)
			}

			// Create new header and decode
			decoded := &Header{}
			hdrData, err := decoded.readFrom(&buf)
			assert.NoError(t, err)

			// Call decodeFields
			decoded.decodeFields(hdrData)

			// Verify decoded values match original
			assert.Equal(t, original.BlockID(), decoded.BlockID())
			assert.Equal(t, original.BlockTimestamp(), decoded.BlockTimestamp())
			assert.Equal(t, original.QueryLength(), decoded.QueryLength())
			assert.Equal(t, original.IsError(), decoded.IsError())

			if original.IsError() {
				assert.Equal(t, original.ErrorCode(), decoded.ErrorCode())
			} else {
				assert.Equal(t, original.ResponseLength(), decoded.ResponseLength())
			}
		})
	}
}
