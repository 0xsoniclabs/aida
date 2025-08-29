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
