package utils

import (
	"math/big"

	"github.com/0xsoniclabs/aida/txcontext"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/vm"
)

// PrepareBlockCtx creates a block context for evm call from given BlockEnvironment.
func PrepareBlockCtx(inputEnv txcontext.BlockEnvironment, hashError *error) *vm.BlockContext {
	getHash := func(num uint64) common.Hash {
		var h common.Hash
		h, *hashError = inputEnv.GetBlockHash(num)
		return h
	}

	blockCtx := &vm.BlockContext{
		CanTransfer: core.CanTransfer,
		Transfer:    core.Transfer,
		Coinbase:    inputEnv.GetCoinbase(),
		BlockNumber: new(big.Int).SetUint64(inputEnv.GetNumber()),
		Time:        inputEnv.GetTimestamp(),
		Difficulty:  inputEnv.GetDifficulty(),
		Random:      inputEnv.GetRandom(),
		GasLimit:    inputEnv.GetGasLimit(),
		GetHash:     getHash,
	}
	// If currentBaseFee is defined, add it to the vmContext.
	baseFee := inputEnv.GetBaseFee()
	if baseFee != nil {
		blockCtx.BaseFee = new(big.Int).Set(baseFee)
	}

	blobBaseFee := inputEnv.GetBlobBaseFee()
	if blobBaseFee != nil {
		blockCtx.BlobBaseFee = new(big.Int).Set(blobBaseFee)
	}
	return blockCtx
}
