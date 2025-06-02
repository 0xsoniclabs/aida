package state

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCache_NewCodeCache(t *testing.T) {
	// case < 0
	cc := NewCodeCache(-1)
	assert.Equal(t, &codeCache{}, cc)

	// case 0
	cc = NewCodeCache(0)
	assert.Equal(t, &codeCache{}, cc)

	// case > 0
	cc = NewCodeCache(10)
	ccx, ok := cc.(*codeCache)
	assert.True(t, ok)
	assert.NotNil(t, ccx.cache)
}

func TestCache_Access(t *testing.T) {
	cc := NewCodeCache(10)
	h := cc.Get(common.HexToAddress("0x1"), []byte("code"))
	assert.Equal(t, common.HexToHash("0x2dc081a8d6d4714c79b5abd2e9b08c3a33b4ef1dcf946ef8b8cf6c495014f47b"), h)

	h = cc.Get(common.HexToAddress("0x1"), []byte("code"))
	assert.Equal(t, common.HexToHash("0x2dc081a8d6d4714c79b5abd2e9b08c3a33b4ef1dcf946ef8b8cf6c495014f47b"), h)
}
