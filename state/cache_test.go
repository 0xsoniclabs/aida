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
