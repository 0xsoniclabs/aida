package stochastic

import (
	"github.com/0xsoniclabs/aida/state"
	"github.com/ethereum/go-ethereum/common"
	"go.uber.org/mock/gomock"
	"testing"
)

// TODO test all proxy calls

func TestEventProxy_GetLogs(t *testing.T) {
	ctrl := gomock.NewController(t)
	base := state.NewMockStateDB(ctrl)
	reg := NewEventRegistry()
	proxy := NewEventProxy(base, &reg)
	hash := common.Hash{0x12}
	blk := uint64(2)
	blkHash := common.Hash{2}
	blkTimestamp := uint64(13)
	base.EXPECT().GetLogs(hash, blk, blkHash, blkTimestamp)
	proxy.GetLogs(hash, blk, blkHash, blkTimestamp)
}
