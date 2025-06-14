package utils

import (
	"math/big"
	"testing"

	"github.com/0xsoniclabs/aida/txcontext"
	"github.com/ethereum/go-ethereum/common"
	"go.uber.org/mock/gomock"
)

func TestPrepareBlockCtx(t *testing.T) {
	ctrl := gomock.NewController(t)
	env := txcontext.NewMockBlockEnvironment(ctrl)

	coinbase := common.HexToAddress("0x1")
	number := uint64(42)
	timestamp := uint64(1000)
	difficulty := big.NewInt(100)
	random := common.HexToHash("0x2")
	gasLimit := uint64(8000000)
	baseFee := big.NewInt(200)
	blobBaseFee := big.NewInt(300)
	blockHash := common.HexToHash("0x3")

	env.EXPECT().GetCoinbase().Return(coinbase)
	env.EXPECT().GetNumber().Return(number)
	env.EXPECT().GetTimestamp().Return(timestamp)
	env.EXPECT().GetDifficulty().Return(difficulty)
	env.EXPECT().GetRandom().Return(&random)
	env.EXPECT().GetGasLimit().Return(gasLimit)
	env.EXPECT().GetBaseFee().Return(baseFee)
	env.EXPECT().GetBlobBaseFee().Return(blobBaseFee)
	env.EXPECT().GetBlockHash(uint64(10)).Return(blockHash, nil)

	var hashErr error
	ctx := PrepareBlockCtx(env, &hashErr)

	if ctx.Coinbase != coinbase {
		t.Errorf("expected coinbase %v, got %v", coinbase, ctx.Coinbase)
	}
	if ctx.BlockNumber.Uint64() != number {
		t.Errorf("expected block number %v, got %v", number, ctx.BlockNumber)
	}
	if ctx.Time != timestamp {
		t.Errorf("expected time %v, got %v", timestamp, ctx.Time)
	}
	if ctx.Difficulty.Cmp(difficulty) != 0 {
		t.Errorf("expected difficulty %v, got %v", difficulty, ctx.Difficulty)
	}
	if *ctx.Random != random {
		t.Errorf("expected random %v, got %v", random, ctx.Random)
	}
	if ctx.GasLimit != gasLimit {
		t.Errorf("expected gas limit %v, got %v", gasLimit, ctx.GasLimit)
	}
	if ctx.BaseFee.Cmp(baseFee) != 0 {
		t.Errorf("expected base fee %v, got %v", baseFee, ctx.BaseFee)
	}
	if ctx.BlobBaseFee.Cmp(blobBaseFee) != 0 {
		t.Errorf("expected blob base fee %v, got %v", blobBaseFee, ctx.BlobBaseFee)
	}

	// Test GetHash function
	gotHash := ctx.GetHash(10)
	if gotHash != blockHash {
		t.Errorf("expected block hash %v, got %v", blockHash, gotHash)
	}
	if hashErr != nil {
		t.Errorf("expected hashErr to be nil, got %v", hashErr)
	}
}
