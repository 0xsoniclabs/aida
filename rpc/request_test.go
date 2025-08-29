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
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRequestAndResults_DecodeInfoPendingBlocksSkipValidation(t *testing.T) {
	var r = &RequestAndResults{
		Response: &Response{
			BlockID: 10,
		},
		Query: &Body{
			Params: []interface{}{
				"test", "pending",
			},
		},
		SkipValidation: false,
	}
	r.DecodeInfo()
	if !r.SkipValidation {
		t.Fatal("skip validation must be true")
	}
}

func TestRequestAndResults_DecodeInfo(t *testing.T) {
	t.Run("response not nil", func(t *testing.T) {
		var r = &RequestAndResults{
			Response: &Response{
				BlockID:   10,
				Timestamp: 10000000000000000,
			},
			Query:          &Body{},
			SkipValidation: false,
		}
		r.DecodeInfo()
		assert.Equal(t, 10, r.RequestedBlock)
		assert.Equal(t, uint64(0x989680), r.Timestamp)
	})

	t.Run("response nil", func(t *testing.T) {
		var r = &RequestAndResults{
			Query: &Body{},
			Error: &ErrorResponse{
				BlockID:   10,
				Timestamp: 10000000000000000,
			},
			SkipValidation: false,
		}
		r.DecodeInfo()
		assert.Equal(t, 10, r.RequestedBlock)
		assert.Equal(t, uint64(0x989680), r.Timestamp)
	})
}

func TestRequestAndResults_findRequestedBlock(t *testing.T) {
	t.Run("single param", func(t *testing.T) {
		var r = &RequestAndResults{
			Query:         &Body{},
			RecordedBlock: 10,
		}
		r.Query.Params = []interface{}{"test"}
		r.findRequestedBlock()
		assert.Equal(t, r.SkipValidation, false)
		assert.Equal(t, r.RequestedBlock, 10)
	})

	t.Run("pending", func(t *testing.T) {
		var r = &RequestAndResults{
			Query:         &Body{},
			RecordedBlock: 10,
		}
		r.Query.Params = []interface{}{"test", "pending"}
		r.findRequestedBlock()
		assert.Equal(t, r.SkipValidation, true)
		assert.Equal(t, r.RequestedBlock, 10)
	})

	t.Run("latest", func(t *testing.T) {
		var r = &RequestAndResults{
			Query:         &Body{},
			RecordedBlock: 10,
		}
		r.Query.Params = []interface{}{"test", "latest"}
		r.findRequestedBlock()
		assert.Equal(t, r.SkipValidation, false)
		assert.Equal(t, r.RequestedBlock, 10)
	})

	t.Run("earliest", func(t *testing.T) {
		var r = &RequestAndResults{
			Query:         &Body{},
			RecordedBlock: 10,
		}
		r.Query.Params = []interface{}{"test", "earliest"}
		r.findRequestedBlock()
		assert.Equal(t, r.SkipValidation, false)
		assert.Equal(t, r.RequestedBlock, 0)
	})

	t.Run("invalid", func(t *testing.T) {
		var r = &RequestAndResults{
			Query:         &Body{},
			RecordedBlock: 10,
		}
		r.Query.Params = []interface{}{"test", "0x1234"}
		r.findRequestedBlock()
		assert.Equal(t, r.SkipValidation, false)
		assert.Equal(t, r.RequestedBlock, 4660)
	})
}
