// Copyright 2025 Sonic Labs
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

package rpc

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"testing"
	"time"

	"github.com/0xsoniclabs/sonic/gossip/filters"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

// Package-level function tests
func TestRpc_newIterator(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockContext := context.TODO()
	mockReadCloser := io.NopCloser(bytes.NewReader([]byte{}))
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
	mockReadCloser := io.NopCloser(bytes.NewReader([]byte{}))
	iter := newIterator(mockContext, mockReadCloser, 1)
	out := iter.Next()
	assert.False(t, out) // Expecting false since Read will return io.EOF
}

func TestIterator_Close(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockContext := context.TODO()
	mockReadCloser := io.NopCloser(bytes.NewReader([]byte{}))
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
		mockReadCloser := io.NopCloser(bytes.NewReader([]byte{0x1, 0x2}))
		iter := &iterator{
			in: mockReadCloser,
		}

		err := iter.loadPayload([]byte{0x1, 0x2})
		assert.NoError(t, err)
	})

	t.Run("error", func(t *testing.T) {
		mockReadCloser := io.NopCloser(bytes.NewReader([]byte{}))
		iter := &iterator{
			in: mockReadCloser,
		}

		err := iter.loadPayload([]byte{0x1, 0x2})
		assert.Error(t, err)
		assert.Equal(t, io.EOF, err)
	})

}

func TestIterator_ReadRpcFile(t *testing.T) {
	path := "/media/herbert/WorkData/sonic/rpc_logs/rpc_logs.gz"

	iter, err := NewFileReader(context.Background(), path)
	assert.NoError(t, err)
	defer iter.Close()

	minBlock := uint64(math.MaxUint64)
	maxBlock := uint64(0)
	c := 0
	for iter.Next() {

		req := iter.Value()
		assert.NoError(t, iter.Error())
		assert.NotNil(t, req)

		if req.Query.Method != "eth_getLogs" {
			continue
		}

		if len(req.ResponseRaw) <= 2 {
			continue
		}

		var filterQuery []filters.FilterCriteria
		err = json.Unmarshal(req.ParamsRaw, &filterQuery)
		assert.NoError(t, err)
		require.Equal(t, 1, len(filterQuery))

		// Skip all queries that do not filter by topics or addresses.
		query := filterQuery[0]
		if len(query.Topics) == 0 && len(query.Addresses) == 0 {
			continue
		}

		var results []*types.Log
		err = json.Unmarshal(req.Response.Result, &results)
		require.NoError(t, err)
		for _, log := range results {
			minBlock = min(minBlock, log.BlockNumber)
			maxBlock = max(maxBlock, log.BlockNumber)
		}

		fmt.Printf("Request:\n")
		fmt.Printf("  Time: 	%v\n", time.Unix(int64(req.Timestamp), 0))
		fmt.Printf("  Method:	%s\n", req.Query.Method)
		fmt.Printf("  Params:	%s\n", req.Query.Params)
		fmt.Printf("  ParamsRaw: %s\n", string(req.ParamsRaw))
		fmt.Printf("  Query:    %+v\n", filterQuery[0])
		fmt.Printf("  Response:	%s\n", string(req.Response.Result))

		c++
		if c >= 5 {
			break
		}

		/*
			c++
			if c%1000 == 0 {
				fmt.Printf("Processed %d eth_getLogs requests, minimal block %d, maximal block %d\n", c, minBlock, maxBlock)
			}
		*/
	}

	fmt.Printf("Processed %d eth_getLogs requests, minimal block %d, maximal block %d\n", c, minBlock, maxBlock)

	t.Fail()

	/*
		methods := map[string]int{}
		printStats := func() {
			fmt.Printf("Method, Count\n")
			for _, k := range slices.Sorted(maps.Keys(methods)) {
				t.Logf("%s, %d\n", k, methods[k])
			}
		}

		total := 0
		for iter.Next() {
			req := iter.Value()
			assert.NoError(t, iter.Error())
			assert.NotNil(t, req)
			methods[req.Query.Method]++

			total++
			if total%100 == 0 {
				printStats()
			}

			if req.Query.Method == "eth_getLogs" {

				if string(req.Response.Result) == "[]" {
					continue
				}

				params, ok := req.Query.Params[0].(map[string]interface{})
				if !ok {
					continue
				}
				topics, ok := params["topics"].([]interface{})
				if !ok {
					continue
				}

				if len(topics) == 0 {
					continue
				}

				_, hasSingle := topics[0].(string)
				_, hasMultiple := topics[0].([]interface{})
				if !hasSingle && !hasMultiple {
					continue
				}

				fmt.Printf("Params: %v\n", req.Query.Params)
				parsed, err := parseQuery(req.Query.Params)
				assert.NoError(t, err)
				fmt.Printf("Parsed Query: %+v\n\n", parsed)

				fmt.Printf("Response: %s\n\n", string(req.Response.Result))
			}

		}
		printStats()

		t.Fail()
	*/
}

func parseQuery(params []interface{}) (Query, error) {
	encoded, err := json.Marshal(params[0])
	if err != nil {
		return Query{}, err
	}
	var q Query
	err = json.Unmarshal(encoded, &q)
	return q, err
}

type Query struct {
	FromBlock string `json:"fromBlock"`
	ToBlock   string `json:"toBlock"`
	BlockHash string `json:"blockHash"`
	Address   any    `json:"address"`
	Topics    any    `json:"topics"`
}

/*
[
	{"address":"0x2f6f07cdcf3588944bf4c42ac74ff24bf56e7590","topics":["0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef","0x000000000000000000000000dc6bb9f44ad17d19fd31cefafe0e498e9f971adf","0x0000000000000000000000000000000000000000000000000000000000000000"],"data":"0x00000000000000000000000000000000000000000000000000760569fb704000","blockNumber":"0x3bc2bb2","transactionHash":"0x4259bc4525099ee1919b5f2bc9302f925f6a7109fcab6dfd360455d5850850b2","transactionIndex":"0x5","blockHash":"0x000339ce00000959bc2e6e579f46e43303ac42ef3fd8bb7043955a50dfac4e9a","logIndex":"0x19","removed":false},
	{"address":"0xd6070ae98b8069de6b494332d1a1a81b6179d960","topics":["0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef","0x00000000000000000000000097bbb619fd59e31f96546c68ed78372ba535d2a1","0x0000000000000000000000001656728af3a14e1319f030dc147fabf6f627059e"],"data":"0x00000000000000000000000000000000000000000000000003e2cff4faa842f2","blockNumber":"0x3bc2bb2","transactionHash":"0x9116a52245ea2abc674c0c9c33d71596ebcdf968a66d7d9902ba51948ea2a5a0","transactionIndex":"0x2","blockHash":"0x000339ce00000959bc2e6e579f46e43303ac42ef3fd8bb7043955a50dfac4e9a","logIndex":"0xb","removed":false},
	{"address":"0x21be370d5312f44cb42ce377bc9b8a0cef1a4c83","topics":["0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef","0x0000000000000000000000001656728af3a14e1319f030dc147fabf6f627059e","0x000000000000000000000000ac97153e7ce86fb3e61681b969698af7c22b4b12"],"data":"0x000000000000000000000000000000000000000000000011f4b03cd788675b2b","blockNumber":"0x3bc2bb2","transactionHash":"0x9116a52245ea2abc674c0c9c33d71596ebcdf968a66d7d9902ba51948ea2a5a0","transactionIndex":"0x2","blockHash":"0x000339ce00000959bc2e6e579f46e43303ac42ef3fd8bb7043955a50dfac4e9a","logIndex":"0xc","removed":false},
	{"address":"0x04068da6c83afcfa0e13ba15a6696662335d5b75","topics":["0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef","0x000000000000000000000000ac97153e7ce86fb3e61681b969698af7c22b4b12","0x00000000000000000000000097bbb619fd59e31f96546c68ed78372ba535d2a1"],"data":"0x00000000000000000000000000000000000000000000000000000000074bbc1e","blockNumber":"0x3bc2bb2","transactionHash":"0x9116a52245ea2abc674c0c9c33d71596ebcdf968a66d7d9902ba51948ea2a5a0","transactionIndex":"0x2","blockHash":"0x000339ce00000959bc2e6e579f46e43303ac42ef3fd8bb7043955a50dfac4e9a","logIndex":"0xf","removed":false},
	{"address":"0xd6070ae98b8069de6b494332d1a1a81b6179d960","topics":["0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef","0x00000000000000000000000097bbb619fd59e31f96546c68ed78372ba535d2a1","0x000000000000000000000000c28cf9aebfe1a07a27b3a4d722c841310e504fe3"],"data":"0x00000000000000000000000000000000000000000000000001470b59b1c721af","blockNumber":"0x3bc2bb2","transactionHash":"0x9116a52245ea2abc674c0c9c33d71596ebcdf968a66d7d9902ba51948ea2a5a0","transactionIndex":"0x2","blockHash":"0x000339ce00000959bc2e6e579f46e43303ac42ef3fd8bb7043955a50dfac4e9a","logIndex":"0x12","removed":false},
	{"address":"0x21be370d5312f44cb42ce377bc9b8a0cef1a4c83","topics":["0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef","0x000000000000000000000000c28cf9aebfe1a07a27b3a4d722c841310e504fe3","0x000000000000000000000000ac97153e7ce86fb3e61681b969698af7c22b4b12"],"data":"0x000000000000000000000000000000000000000000000005e9085a100d5abe14","blockNumber":"0x3bc2bb2","transactionHash":"0x9116a52245ea2abc674c0c9c33d71596ebcdf968a66d7d9902ba51948ea2a5a0","transactionIndex":"0x2","blockHash":"0x000339ce00000959bc2e6e579f46e43303ac42ef3fd8bb7043955a50dfac4e9a","logIndex":"0x13","removed":false},
	{"address":"0x04068da6c83afcfa0e13ba15a6696662335d5b75","topics":["0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef","0x000000000000000000000000ac97153e7ce86fb3e61681b969698af7c22b4b12","0x00000000000000000000000097bbb619fd59e31f96546c68ed78372ba535d2a1"],"data":"0x000000000000000000000000000000000000000000000000000000000266a95c","blockNumber":"0x3bc2bb2","transactionHash":"0x9116a52245ea2abc674c0c9c33d71596ebcdf968a66d7d9902ba51948ea2a5a0","transactionIndex":"0x2","blockHash":"0x000339ce00000959bc2e6e579f46e43303ac42ef3fd8bb7043955a50dfac4e9a","logIndex":"0x16","removed":false},
	{"address":"0x04068da6c83afcfa0e13ba15a6696662335d5b75","topics":["0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef","0x000000000000000000000000554ef7d4744d2a326a536bffa60f37e1dcf1de40","0x00000000000000000000000012edea9cd262006cc3c4e77c90d2cd2dd4b1eb97"],"data":"0x0000000000000000000000000000000000000000000000000000000001ca5fc0","blockNumber":"0x3bc2bb2","transactionHash":"0xc52f81e620bdb2e0fd0b1c89ec36f6ad39dba8cdea6d7b4490b2c6d683978255","transactionIndex":"0x6","blockHash":"0x000339ce00000959bc2e6e579f46e43303ac42ef3fd8bb7043955a50dfac4e9a","logIndex":"0x21","removed":false},
	{"address":"0x04068da6c83afcfa0e13ba15a6696662335d5b75","topics":["0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef","0x0000000000000000000000008d98c325da223fc6750a3b79c4e05dc6a66aa69c","0x000000000000000000000000ac97153e7ce86fb3e61681b969698af7c22b4b12"],"data":"0x000000000000000000000000000000000000000000000000000000000b06e040","blockNumber":"0x3bc2bb2","transactionHash":"0xe49deb7ee8cf70813ff398460f7c7eb976b433a9be158d30f81a1653145dfcef","transactionIndex":"0x1","blockHash":"0x000339ce00000959bc2e6e579f46e43303ac42ef3fd8bb7043955a50dfac4e9a","logIndex":"0x6","removed":false},
	{"address":"0x21be370d5312f44cb42ce377bc9b8a0cef1a4c83","topics":["0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef","0x000000000000000000000000ac97153e7ce86fb3e61681b969698af7c22b4b12","0x0000000000000000000000005023882f4d1ec10544fcb2066abe9c1645e95aa0"],"data":"0x00000000000000000000000000000000000000000000001b0984dde2bbd5595b","blockNumber":"0x3bc2bb2","transactionHash":"0xe49deb7ee8cf70813ff398460f7c7eb976b433a9be158d30f81a1653145dfcef","transactionIndex":"0x1","blockHash":"0x000339ce00000959bc2e6e579f46e43303ac42ef3fd8bb7043955a50dfac4e9a","logIndex":"0x7","removed":false},
	{"address":"0x21be370d5312f44cb42ce377bc9b8a0cef1a4c83","topics":["0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef","0x0000000000000000000000005023882f4d1ec10544fcb2066abe9c1645e95aa0","0x0000000000000000000000000000000000000000000000000000000000000000"],"data":"0x00000000000000000000000000000000000000000000001b0984dde2bbd5595b","blockNumber":"0x3bc2bb2","transactionHash":"0xe49deb7ee8cf70813ff398460f7c7eb976b433a9be158d30f81a1653145dfcef","transactionIndex":"0x1","blockHash":"0x000339ce00000959bc2e6e579f46e43303ac42ef3fd8bb7043955a50dfac4e9a","logIndex":"0xa","removed":false},
	{"address":"0x04068da6c83afcfa0e13ba15a6696662335d5b75","topics":["0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef","0x00000000000000000000000063aed236b4611c84f3950c3191b74b843c0b6e36","0x00000000000000000000000012edea9cd262006cc3c4e77c90d2cd2dd4b1eb97"],"data":"0x000000000000000000000000000000000000000000000000000000000cab6875","blockNumber":"0x3bc2bb3","transactionHash":"0xeaca4c65b6d0f1f8d3a38857ab40968e530a1a7e68954778d3bc14a81cf231c9","transactionIndex":"0x0","blockHash":"0x000339ce00000978ba22f1e8a0acdee4845374cdc5c8a52d1cac46e1712d4ac3","logIndex":"0x2","removed":false},
	{"address":"0x04068da6c83afcfa0e13ba15a6696662335d5b75","topics":["0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef","0x000000000000000000000000f3b6a8127561d556187a60c0d5b0a24eeab5b0f6","0x00000000000000000000000012edea9cd262006cc3c4e77c90d2cd2dd4b1eb97"],"data":"0x0000000000000000000000000000000000000000000000000000000006221dfd","blockNumber":"0x3bc2bb4","transactionHash":"0x63dd287af140cfca5a8e94cdf9cd6a7cf4045f0df97137b3e1a0be53f5294400","transactionIndex":"0x0","blockHash":"0x000339ce0000098394a8aa623aad93948955c5de074139ba6141fc1575f25839","logIndex":"0x2","removed":false},
	{"address":"0x141faa507855e56396eadbd25ec82656755cd61e","topics":["0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef","0x0000000000000000000000007a1f47c8a26fd895228947ffc0482f3dd9c2ca29","0x000000000000000000000000b8d86d6db117e21c27636034d3dd8859018daf9c"],"data":"0x00000000000000000000000000000000000000000000000003074fca9142c26f","blockNumber":"0x3bc2bb4","transactionHash":"0x6cbf2cc2e14385fa96197c29734649ffce79a1b9f5acd3cb29c1b5c27b9025be","transactionIndex":"0x3","blockHash":"0x000339ce0000098394a8aa623aad93948955c5de074139ba6141fc1575f25839","logIndex":"0xb","removed":false},
	{"address":"0x04068da6c83afcfa0e13ba15a6696662335d5b75","topics":["0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef","0x0000000000000000000000009b583af2f2eda21fd6ee8eb506753c392a507773","0x00000000000000000000000012edea9cd262006cc3c4e77c90d2cd2dd4b1eb97"],"data":"0x000000000000000000000000000000000000000000000000000000000346bad8","blockNumber":"0x3bc2bb5","transactionHash":"0x9d7c1e470c5f69310591f67c54ccc9d65b56ce0315e476a76823152c2bbde854","transactionIndex":"0x0","blockHash":"0x000339ce0000098ee88a93d8f9ffbff8f3c6c0f6314a2f2cf80953ce9a36cadd","logIndex":"0x2","removed":false},
	{"address":"0xc418b123885d732ed042b16e12e259741863f723","topics":["0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef","0x0000000000000000000000000f5a6cf441cd0bf64c9c00cd4e13cdc2439a7ff8","0x0000000000000000000000002a9c55b6dc56da178f9f9a566f1161237b73ba66"],"data":"0x0000000000000000000000000000000000000000000000037e2557a29e6b0000","blockNumber":"0x3bc2bb6","transactionHash":"0x46ac503d8e7dd956ff00fa1727b1548b79e1c7e1577621672a50624d5fc82b15","transactionIndex":"0x7","blockHash":"0x000339ce000009a5ce1e5f4672cc29ff2e5d52493a11a14d6bf89a3658d065a6","logIndex":"0x37","removed":false},
	{"address":"0x94d9e02d115646dfc407abde75fa45256d66e043","topics":["0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef","0x0000000000000000000000000000000000000000000000000000000000000000","0x00000000000000000000000048b40efc7ed5e697d481d2f5622b65c37b263e17","0x000000000000000000000000000000000000000000000000000000000003bcba"],"data":"0x","blockNumber":"0x3bc2bb6","transactionHash":"0xf9fdcc0f16934e098091140e1bb95f5232af0da8d8776963be1bb60022e4f6f2","transactionIndex":"0x4","blockHash":"0x000339ce000009a5ce1e5f4672cc29ff2e5d52493a11a14d6bf89a3658d065a6","logIndex":"0x33","removed":false},
	{"address":"0x2f6f07cdcf3588944bf4c42ac74ff24bf56e7590","topics":["0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef","0x00000000000000000000000029dcebd4bba9d8dc663f21ed5b426e7547035542","0x0000000000000000000000000000000000000000000000000000000000000000"],"data":"0x000000000000000000000000000000000000000000000008e22730c78f430000","blockNumber":"0x3bc2bb7","transactionHash":"0x79ccecec460506b93371970c80e165d02b4413c27837b3ea50556175f2b3aac1","transactionIndex":"0x2","blockHash":"0x000339ce000009afd58b2d1ecdf74cdd5d779aa793adca947556f7047f37122c","logIndex":"0x7","removed":false},
	{"address":"0x2f6f07cdcf3588944bf4c42ac74ff24bf56e7590","topics":["0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef","0x0000000000000000000000001978918dff96f84c5cf2d260ff3bc6d45712959c","0x0000000000000000000000000000000000000000000000000000000000000000"],"data":"0x000000000000000000000000000000000000000000000028dba3924d7e2c1d70","blockNumber":"0x3bc2bb7","transactionHash":"0x9f0960e5a27a9f412d2b34e79d9b2e00b568d26f3ec48303f31c3d80df8cf1cd","transactionIndex":"0x0","blockHash":"0x000339ce000009afd58b2d1ecdf74cdd5d779aa793adca947556f7047f37122c","logIndex":"0x0","removed":false},
	{"address":"0x04068da6c83afcfa0e13ba15a6696662335d5b75","topics":["0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef","0x00000000000000000000000012edea9cd262006cc3c4e77c90d2cd2dd4b1eb97","0x00000000000000000000000072dc7fa5eeb901a34173c874a7333c8d1b34bca9"],"data":"0x0000000000000000000000000000000000000000000000000000000000443e85","blockNumber":"0x3bc2bb8","transactionHash":"0x057d89dff3dc9142873a2912ee7bf62cf293dc85481f90a85e05d78b89893578","transactionIndex":"0xa","blockHash":"0x000339ce000009c43fc8215cf84968e050d5f6fc95250d62d74d989dc6124a25","logIndex":"0x21","removed":false},
	{"address":"0x04068da6c83afcfa0e13ba15a6696662335d5b75","topics":["0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef","0x00000000000000000000000072dc7fa5eeb901a34173c874a7333c8d1b34bca9","0x000000000000000000000000286ab107c5e9083dbed35a2b5fb0242538f4f9bf"],"data":"0x0000000000000000000000000000000000000000000000000000000000443e85","blockNumber":"0x3bc2bb8","transactionHash":"0x057d89dff3dc9142873a2912ee7bf62cf293dc85481f90a85e05d78b89893578","transactionIndex":"0xa","blockHash":"0x000339ce000009c43fc8215cf84968e050d5f6fc95250d62d74d989dc6124a25","logIndex":"0x25","removed":false},
	{"address":"0x21be370d5312f44cb42ce377bc9b8a0cef1a4c83","topics":["0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef","0x000000000000000000000000286ab107c5e9083dbed35a2b5fb0242538f4f9bf","0x000000000000000000000000382a9b0bc5d29e96c3a0b81ce9c64d6c8f150efb"],"data":"0x000000000000000000000000000000000000000000000000a78252e0ee56a5eb","blockNumber":"0x3bc2bb8","transactionHash":"0x057d89dff3dc9142873a2912ee7bf62cf293dc85481f90a85e05d78b89893578","transactionIndex":"0xa","blockHash":"0x000339ce000009c43fc8215cf84968e050d5f6fc95250d62d74d989dc6124a25","logIndex":"0x26","removed":false},
	{"address":"0x21be370d5312f44cb42ce377bc9b8a0cef1a4c83","topics":["0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef","0x000000000000000000000000382a9b0bc5d29e96c3a0b81ce9c64d6c8f150efb","0x0000000000000000000000000000000000000000000000000000000000000000"],"data":"0x000000000000000000000000000000000000000000000000a78252e0ee56a5eb","blockNumber":"0x3bc2bb8","transactionHash":"0x057d89dff3dc9142873a2912ee7bf62cf293dc85481f90a85e05d78b89893578","transactionIndex":"0xa","blockHash":"0x000339ce000009c43fc8215cf84968e050d5f6fc95250d62d74d989dc6124a25","logIndex":"0x28","removed":false},
	{"address":"0x04068da6c83afcfa0e13ba15a6696662335d5b75","topics":["0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef","0x00000000000000000000000012edea9cd262006cc3c4e77c90d2cd2dd4b1eb97","0x0000000000000000000000006b7c2bd701b30a03ee908062376dbba893247696"],"data":"0x000000000000000000000000000000000000000000000000000000001a95f03e","blockNumber":"0x3bc2bb8","transactionHash":"0x18a87347b7d4c17af88cdbe07e2e88668de955ed88f4073906a331d01656248e","transactionIndex":"0xe","blockHash":"0x000339ce000009c43fc8215cf84968e050d5f6fc95250d62d74d989dc6124a25","logIndex":"0x39","removed":false},
	{"address":"0x04068da6c83afcfa0e13ba15a6696662335d5b75","topics":["0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef","0x00000000000000000000000012edea9cd262006cc3c4e77c90d2cd2dd4b1eb97","0x000000000000000000000000bdb256414eff6c820238b2462f7ce7ac3f884263"],"data":"0x000000000000000000000000000000000000000000000000000000000eaabe12","blockNumber":"0x3bc2bb8","transactionHash":"0x20460e406df6d49f4ef6b914cff05fc2e7878d6dd5274d88112831297981cb08","transactionIndex":"0x13","blockHash":"0x000339ce000009c43fc8215cf84968e050d5f6fc95250d62d74d989dc6124a25","logIndex":"0x4d","removed":false},
	{"address":"0x04068da6c83afcfa0e13ba15a6696662335d5b75","topics":["0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef","0x00000000000000000000000012edea9cd262006cc3c4e77c90d2cd2dd4b1eb97","0x000000000000000000000000b7538572aedc2ac2d7f177397bcc7941741cc97c"],"data":"0x000000000000000000000000000000000000000000000000000000000004932c","blockNumber":"0x3bc2bb8","transactionHash":"0x60b19225eb974b47dee95e128f962d5d8df3faa9a57f1151f8ca96dd290da655","transactionIndex":"0x12","blockHash":"0x000339ce000009c43fc8215cf84968e050d5f6fc95250d62d74d989dc6124a25","logIndex":"0x49","removed":false},
	{"address":"0x04068da6c83afcfa0e13ba15a6696662335d5b75","topics":["0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef","0x00000000000000000000000012edea9cd262006cc3c4e77c90d2cd2dd4b1eb97","0x00000000000000000000000066653be27064134055687ee63a3a76f3ba43d24d"],"data":"0x000000000000000000000000000000000000000000000000000000000001d478","blockNumber":"0x3bc2bb8","transactionHash":"0x63e254c7fcb38898e4066adc0a813db614ed372e3d95f2d48cc44ba716157d8e","transactionIndex":"0xf","blockHash":"0x000339ce000009c43fc8215cf84968e050d5f6fc95250d62d74d989dc6124a25","logIndex":"0x3d","removed":false},
	{"address":"0x04068da6c83afcfa0e13ba15a6696662335d5b75","topics":["0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef","0x00000000000000000000000012edea9cd262006cc3c4e77c90d2cd2dd4b1eb97","0x00000000000000000000000072dc7fa5eeb901a34173c874a7333c8d1b34bca9"],"data":"0x000000000000000000000000000000000000000000000000000000000020bccf","blockNumber":"0x3bc2bb8","transactionHash":"0x76df593fb350f4767bda23202a05ecc4eff92be4f0885b4d1418e599adc23325","transactionIndex":"0x14","blockHash":"0x000339ce000009c43fc8215cf84968e050d5f6fc95250d62d74d989dc6124a25","logIndex":"0x51","removed":false},
	{"address":"0x04068da6c83afcfa0e13ba15a6696662335d5b75","topics":["0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef","0x00000000000000000000000072dc7fa5eeb901a34173c874a7333c8d1b34bca9","0x000000000000000000000000286ab107c5e9083dbed35a2b5fb0242538f4f9bf"],"data":"0x000000000000000000000000000000000000000000000000000000000020bccf","blockNumber":"0x3bc2bb8","transactionHash":"0x76df593fb350f4767bda23202a05ecc4eff92be4f0885b4d1418e599adc23325","transactionIndex":"0x14","blockHash":"0x000339ce000009c43fc8215cf84968e050d5f6fc95250d62d74d989dc6124a25","logIndex":"0x55","removed":false},
	{"address":"0x21be370d5312f44cb42ce377bc9b8a0cef1a4c83","topics":["0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef","0x000000000000000000000000286ab107c5e9083dbed35a2b5fb0242538f4f9bf","0x000000000000000000000000382a9b0bc5d29e96c3a0b81ce9c64d6c8f150efb"],"data":"0x000000000000000000000000000000000000000000000000505b252fa631a89b","blockNumber":"0x3bc2bb8","transactionHash":"0x76df593fb350f4767bda23202a05ecc4eff92be4f0885b4d1418e599adc23325","transactionIndex":"0x14","blockHash":"0x000339ce000009c43fc8215cf84968e050d5f6fc95250d62d74d989dc6124a25","logIndex":"0x56","removed":false},
	{"address":"0x21be370d5312f44cb42ce377bc9b8a0cef1a4c83","topics":["0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef","0x000000000000000000000000382a9b0bc5d29e96c3a0b81ce9c64d6c8f150efb","0x0000000000000000000000000000000000000000000000000000000000000000"],"data":"0x000000000000000000000000000000000000000000000000505b252fa631a89b","blockNumber":"0x3bc2bb8","transactionHash":"0x76df593fb350f4767bda23202a05ecc4eff92be4f0885b4d1418e599adc23325","transactionIndex":"0x14","blockHash":"0x000339ce000009c43fc8215cf84968e050d5f6fc95250d62d74d989dc6124a25","logIndex":"0x58","removed":false},
	{"address":"0x04068da6c83afcfa0e13ba15a6696662335d5b75","topics":["0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef","0x00000000000000000000000012edea9cd262006cc3c4e77c90d2cd2dd4b1eb97","0x00000000000000000000000056a238d2e3d2fca8c6eb294c953ae85f2ab3b2d4"],"data":"0x0000000000000000000000000000000000000000000000000000000006f4748b","blockNumber":"0x3bc2bb8","transactionHash":"0x7cb774b70903b23be1ebb587d5b4b7a47d82ead5440aff7ca6cfb38dc6cfef08","transactionIndex":"0xc","blockHash":"0x000339ce000009c43fc8215cf84968e050d5f6fc95250d62d74d989dc6124a25","logIndex":"0x31","removed":false},
	{"address":"0x04068da6c83afcfa0e13ba15a6696662335d5b75","topics":["0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef","0x000000000000000000000000a08f24f6e94e596d2296a3890483f60288718a18","0x00000000000000000000000012edea9cd262006cc3c4e77c90d2cd2dd4b1eb97"],"data":"0x000000000000000000000000000000000000000000000000000000000cf2e9b7","blockNumber":"0x3bc2bb8","transactionHash":"0x8621b6530e7cecd036fdc159979597c13dca569ea86a123a83ce49c101f38f49","transactionIndex":"0x2","blockHash":"0x000339ce000009c43fc8215cf84968e050d5f6fc95250d62d74d989dc6124a25","logIndex":"0x6","removed":false},
	{"address":"0x04068da6c83afcfa0e13ba15a6696662335d5b75","topics":["0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef","0x00000000000000000000000012edea9cd262006cc3c4e77c90d2cd2dd4b1eb97","0x0000000000000000000000003c3c8d09f876a38ac14d3f3a04e8a8ad86cd7d63"],"data":"0x0000000000000000000000000000000000000000000000000000000000000441","blockNumber":"0x3bc2bb8","transactionHash":"0xb64fcf49d177e251853718c3a1f3a22a6fd47e2a465e63b5f1fb3b7d79f98812","transactionIndex":"0x11","blockHash":"0x000339ce000009c43fc8215cf84968e050d5f6fc95250d62d74d989dc6124a25","logIndex":"0x45","removed":false},
	{"address":"0x54447cfa919c096d095d58c2b5814669897bffcd","topics":["0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef","0x0000000000000000000000002b76d89318b1c674bd1fe49491b2674326089cea","0x0000000000000000000000000000000000000000000000000000000000000000"],"data":"0x00000000000000000000000000000000000000000000000000000000ee6b2800","blockNumber":"0x3bc2bb8","transactionHash":"0xbcf96f019843ec6877072be2e29faf94a1f4eea89da959263139e38b52419ba3","transactionIndex":"0x1","blockHash":"0x000339ce000009c43fc8215cf84968e050d5f6fc95250d62d74d989dc6124a25","logIndex":"0x1","removed":false},
	{"address":"0x04068da6c83afcfa0e13ba15a6696662335d5b75","topics":["0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef","0x00000000000000000000000012edea9cd262006cc3c4e77c90d2cd2dd4b1eb97","0x0000000000000000000000002a8b608158ab521dfd669d45d52554b82831be53"],"data":"0x0000000000000000000000000000000000000000000000000000000000000447","blockNumber":"0x3bc2bb8","transactionHash":"0xbdcdf7f5bbf8c01d7aa22fef9f8737b6eec66d4db634a074b23438cdf27bcb9b","transactionIndex":"0xd","blockHash":"0x000339ce000009c43fc8215cf84968e050d5f6fc95250d62d74d989dc6124a25","logIndex":"0x35","removed":false},
	{"address":"0xef4b763385838fffc708000f884026b8c0434275","topics":["0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef","0x0000000000000000000000000000000000000000000000000000000000000000","0x000000000000000000000000b746678dcffa60d6510a8a5bae11302f645a04f3"],"data":"0x0000000000000000000000000000000000000000000000000000000000000000","blockNumber":"0x3bc2bb8","transactionHash":"0xbe81fd75e93f4f0fde63aa7bd68097a8b08c5ce18c149ca064dc2479548a1f0a","transactionIndex":"0x15","blockHash":"0x000339ce000009c43fc8215cf84968e050d5f6fc95250d62d74d989dc6124a25","logIndex":"0x5b","removed":false},
	{"address":"0xef4b763385838fffc708000f884026b8c0434275","topics":["0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef","0x0000000000000000000000000000000000000000000000000000000000000000","0x00000000000000000000000041e38fe4b4f5de23196f429b579775f1a5b3c710"],"data":"0x00000000000000000000000000000000000000000007ddde0f2a66f005b00000","blockNumber":"0x3bc2bb8","transactionHash":"0xbe81fd75e93f4f0fde63aa7bd68097a8b08c5ce18c149ca064dc2479548a1f0a","transactionIndex":"0x15","blockHash":"0x000339ce000009c43fc8215cf84968e050d5f6fc95250d62d74d989dc6124a25","logIndex":"0x5c","removed":false},
	{"address":"0x04068da6c83afcfa0e13ba15a6696662335d5b75","topics":["0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef","0x00000000000000000000000012edea9cd262006cc3c4e77c90d2cd2dd4b1eb97","0x0000000000000000000000000227baa0a5c4fa5eb6e1c6194b5b4a8fc6ff2d6f"],"data":"0x00000000000000000000000000000000000000000000000000000000007ac98b","blockNumber":"0x3bc2bb8","transactionHash":"0xe3a0945fa8094b75e06a7d42587fdbf7070a67df25e4769a8c830068064f769b","transactionIndex":"0x10","blockHash":"0x000339ce000009c43fc8215cf84968e050d5f6fc95250d62d74d989dc6124a25","logIndex":"0x41","removed":false},
	{"address":"0x04068da6c83afcfa0e13ba15a6696662335d5b75","topics":["0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef","0x00000000000000000000000012edea9cd262006cc3c4e77c90d2cd2dd4b1eb97","0x0000000000000000000000003788305c2443ef4fa893069778663b4618196430"],"data":"0x000000000000000000000000000000000000000000000000000000003c4641e4","blockNumber":"0x3bc2bb8","transactionHash":"0xee2a43e6a926ce666dd571377471fec030d66ea30baeba2a5271b377d5fd9565","transactionIndex":"0xb","blockHash":"0x000339ce000009c43fc8215cf84968e050d5f6fc95250d62d74d989dc6124a25","logIndex":"0x2d","removed":false},
	{"address":"0x04068da6c83afcfa0e13ba15a6696662335d5b75","topics":["0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef","0x00000000000000000000000028bc786155d0b5f74c9f55adeb6b50fd6587cd4d","0x00000000000000000000000012edea9cd262006cc3c4e77c90d2cd2dd4b1eb97"],"data":"0x00000000000000000000000000000000000000000000000000000000084bd260","blockNumber":"0x3bc2bb8","transactionHash":"0xfc6dce4e69acd6d79eb0467a9f64af85fa324fba55b91ad3cf48d41bc8650e0f","transactionIndex":"0x3","blockHash":"0x000339ce000009c43fc8215cf84968e050d5f6fc95250d62d74d989dc6124a25","logIndex":"0xf","removed":false}
]
*/
