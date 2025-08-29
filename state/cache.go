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

//go:generate mockgen -source cache.go -destination cache_mock.go -package state
import (
	"sync"

	cc "github.com/0xsoniclabs/carmen/go/common"
	"github.com/ethereum/go-ethereum/common"
)

// CodeCache serves a cache for hashed address code.
type CodeCache interface {
	// Get returns code hash for given addr and code.
	Get(addr common.Address, code []byte) common.Hash
}

// codeKey represents the key for CodeCache.
type codeKey struct {
	addr common.Address
	code string
}

// NewCodeCache creates new instance of CodeCache that stores already retrieved code hashes.
func NewCodeCache(capacity int) CodeCache {
	if capacity <= 0 {
		return &codeCache{}
	}
	return &codeCache{
		cache: cc.NewLruCache[codeKey, common.Hash](capacity),
		mutex: sync.Mutex{},
	}
}

type codeCache struct {
	cache cc.Cache[codeKey, common.Hash]
	mutex sync.Mutex
}

// Get returns code hash for given addr and code.
// If hash does not exist within the cache, it is created and stored.
// This operation is thread-safe.
func (c *codeCache) Get(addr common.Address, code []byte) common.Hash {
	k := codeKey{addr: addr, code: string(code)}
	c.mutex.Lock()
	defer c.mutex.Unlock()
	h, exists := c.cache.Get(k)
	if exists {
		return h
	}

	h = createCodeHash(code)
	c.set(k, h)
	return h
}

func (c *codeCache) set(k codeKey, v common.Hash) {
	c.cache.Set(k, v)
}

func createCodeHash(code []byte) common.Hash {
	return common.Hash(cc.Keccak256(code))
}
