package proxy

import (
	"bytes"
	"encoding/gob"
	"github.com/0xsoniclabs/aida/state"
	"github.com/0xsoniclabs/aida/tracer"
	"github.com/0xsoniclabs/aida/txcontext"
	"github.com/Fantom-foundation/lachesis-base/common/bigendian"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/params"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"math/big"
	"testing"
)

func TestTracerProxy_PrepareSubstate(t *testing.T) {
	ctrl := gomock.NewController(t)
	base := state.NewMockStateDB(ctrl)
	ctx := tracer.NewMockContext(ctrl)
	proxy := NewTracerProxy(base, ctx)
	ws := txcontext.NewWorldState(map[common.Address]txcontext.Account{
		{0x1}: txcontext.NewAccount(
			[]byte{0x2},
			map[common.Hash]common.Hash{{0x3}: {0x4}},
			big.NewInt(12),
			13,
		),
	})
	blk := uint64(123)
	buf := bytes.Buffer{}
	gob.Register(txcontext.NewNilAccount())
	err := gob.NewEncoder(&buf).Encode(ws)
	require.NoError(t, err)
	wantData := append(bigendian.Uint32ToBytes(uint32(buf.Len())), buf.Bytes()...)

	ctx.EXPECT().WriteOp(tracer.PrepareSubstateID, wantData)
	base.EXPECT().PrepareSubstate(ws, blk)

	proxy.PrepareSubstate(ws, blk)
}
func TestTracerProxy_Prepare(t *testing.T) {
	ctrl := gomock.NewController(t)
	base := state.NewMockStateDB(ctrl)
	ctx := tracer.NewMockContext(ctrl)
	proxy := NewTracerProxy(base, ctx)
	rules := params.Rules{
		ChainID:    big.NewInt(146),
		IsIstanbul: true,
		IsBerlin:   true,
		IsLondon:   true,
		IsMerge:    true,
		IsShanghai: true,
		IsCancun:   true,
		IsPrague:   true,
		IsOsaka:    true,
		IsVerkle:   true,
	}
	sender := common.Address{0x1}
	coinbase := common.Address{0x2}
	precompiles := []common.Address{
		{0x3}, {0x4},
	}
	txAccess := types.AccessList{
		{
			Address: common.Address{0x6},
			StorageKeys: []common.Hash{
				{0x7},
			},
		},
	}
	gob.Register(params.Rules{})
	gob.Register(types.AccessList{})
	buf := bytes.Buffer{}
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(rules)
	require.NoError(t, err)
	wantData := append(bigendian.Uint32ToBytes(uint32(buf.Len())), buf.Bytes()...)
	wantData = append(wantData, sender.Bytes()...)
	wantData = append(wantData, coinbase.Bytes()...)
	// 0 byte to signal dest is nil
	wantData = append(wantData, byte(0))
	wantData = append(wantData, bigendian.Uint16ToBytes(uint16(len(precompiles)))...)
	for _, addr := range precompiles {
		wantData = append(wantData, addr.Bytes()...)
	}
	buf.Reset()
	err = enc.Encode(txAccess)
	require.NoError(t, err)
	wantData = append(wantData, bigendian.Uint32ToBytes(uint32(buf.Len()))...)
	wantData = append(wantData, buf.Bytes()...)

	ctx.EXPECT().WriteOp(tracer.PrepareID, wantData)
	base.EXPECT().Prepare(rules, sender, coinbase, nil, precompiles, txAccess)

	proxy.Prepare(rules, sender, coinbase, nil, precompiles, txAccess)
}

func TestTracerProxy_AddLog(t *testing.T) {
	ctrl := gomock.NewController(t)
	base := state.NewMockStateDB(ctrl)
	ctx := tracer.NewMockContext(ctrl)
	proxy := NewTracerProxy(base, ctx)
	log := types.Log{
		Address:        common.Address{0x1},
		Topics:         []common.Hash{{0x2}, {0x3}},
		Data:           []byte{0x4, 0x5},
		BlockNumber:    11,
		TxHash:         common.Hash{0x6},
		TxIndex:        12,
		BlockHash:      common.Hash{0x7},
		BlockTimestamp: 13,
		Index:          14,
		Removed:        true,
	}

	buf := bytes.Buffer{}
	gob.Register(log)
	err := gob.NewEncoder(&buf).Encode(log)
	require.NoError(t, err)
	wantData := append(bigendian.Uint32ToBytes(uint32(buf.Len())), buf.Bytes()...)
	ctx.EXPECT().WriteOp(tracer.AddLogID, wantData)
	base.EXPECT().AddLog(&log)

	proxy.AddLog(&log)
}
