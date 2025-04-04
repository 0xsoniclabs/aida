// Copyright 2024 Fantom Foundation
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

package substate

import (
	"fmt"
	"math/big"

	"github.com/0xsoniclabs/aida/txcontext"
	"github.com/0xsoniclabs/substate/substate"
	"github.com/ethereum/go-ethereum/common"
)

func NewBlockEnvironment(env *substate.Env) txcontext.BlockEnvironment {
	return &blockEnvironment{env}
}

type blockEnvironment struct {
	*substate.Env
}

func (e *blockEnvironment) GetRandom() *common.Hash {
	if e.Random == nil {
		return nil
	}

	h := common.Hash(e.Random.Bytes())
	return &h
}

func (e *blockEnvironment) GetBlockHash(block uint64) (common.Hash, error) {
	if e.BlockHashes == nil {
		return common.Hash{}, fmt.Errorf("getHash(%d) invoked, no blockhashes provided", block)
	}
	h, ok := e.BlockHashes[block]
	if !ok {
		return common.Hash(h), fmt.Errorf("getHash(%d) invoked, blockhash for that block not provided", block)
	}
	return common.Hash(h), nil
}

func (e *blockEnvironment) GetCoinbase() common.Address {
	return common.Address(e.Coinbase)
}

func (e *blockEnvironment) GetDifficulty() *big.Int {
	return e.Difficulty
}

func (e *blockEnvironment) GetGasLimit() uint64 {
	return e.GasLimit
}

func (e *blockEnvironment) GetNumber() uint64 {
	return e.Number
}

func (e *blockEnvironment) GetTimestamp() uint64 {
	return e.Timestamp
}

func (e *blockEnvironment) GetBaseFee() *big.Int {
	return e.BaseFee
}

func (e *blockEnvironment) GetBlobBaseFee() *big.Int {
	return e.BlobBaseFee
}

func (e *blockEnvironment) GetFork() string {
	// for now, only necessary for get-tests
	return ""
}
