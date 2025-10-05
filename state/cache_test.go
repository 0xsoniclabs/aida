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

package state

import (
	"sync"
	"testing"

	cc "github.com/0xsoniclabs/carmen/go/common"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
)

func TestState_NewCodeCache(t *testing.T) {

	t.Run("capacity negative", func(t *testing.T) {
		cache := NewCodeCache(-100)
		assert.Equal(t, &codeCache{}, cache)
	})

	t.Run("capacity zero", func(t *testing.T) {
		cache := NewCodeCache(0)
		assert.Equal(t, &codeCache{}, cache)
	})

	t.Run("capacity positive", func(t *testing.T) {
		cache := NewCodeCache(10)
		assert.NotNil(t, cache)
		c, ok := cache.(*codeCache)
		assert.True(t, ok)
		assert.NotNil(t, c.cache)
	})

}

func TestCodeCache_Get(t *testing.T) {

	t.Run("no cache", func(t *testing.T) {
		cache := codeCache{
			cache: cc.NewLruCache[codeKey, common.Hash](10),
			mutex: sync.Mutex{},
		}
		addr := common.HexToAddress("0x1234")
		expected := createCodeHash([]byte("testcode"))
		out := cache.Get(addr, []byte("testcode"))
		assert.Equal(t, expected, out)
	})

	t.Run("with cache", func(t *testing.T) {
		cache := codeCache{
			cache: cc.NewLruCache[codeKey, common.Hash](10),
			mutex: sync.Mutex{},
		}
		addr := common.HexToAddress("0x1234")
		code := []byte("testcode")
		expected := createCodeHash([]byte("testcode"))
		cache.set(codeKey{addr: addr, code: string(code)}, expected)
		out := cache.Get(addr, []byte("testcode"))
		assert.Equal(t, expected, out)
	})

}
