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

package utils

import (
	"context"
	"fmt"

	"github.com/0xsoniclabs/aida/logger"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/rpc"
	lru "github.com/hashicorp/golang-lru"
)

var StateHashQueue StateHashQueueProvider

func MakeStateHashQueueProvider(cfg *Config) StateHashProvider {
	ipcPath := cfg.OperaDb + "/sonic.ipc"

	log := logger.NewLogger("info", "StateHashQueueProvider")
	client, err := GetRpcOrIpcClient(context.Background(), cfg.ChainID, ipcPath, log)
	if err != nil {
		return nil
	}
	cache, _ := lru.New(1000) // Create an LRU cache with a capacity of 10
	StateHashQueue = StateHashQueueProvider{stateHashCache: cache, client: client}
	return &StateHashQueue
}

type StateHashQueueProvider struct {
	stateHashCache *lru.Cache
	client         *rpc.Client
}

func (p *StateHashQueueProvider) GetStateHash(number int) (common.Hash, error) {
	//stateRoot, ok := p.stateHashCache.Get(number)
	//if ok {
	//	return common.HexToHash(stateRoot.(string)), nil
	//}
	numberHex := fmt.Sprintf("0x%x", number)
	blk, err := RetrieveBlock(p.client, numberHex, false)
	if err != nil {
		return common.Hash{}, err
	}
	return common.HexToHash(blk["stateRoot"].(string)), nil
}

func (p *StateHashQueueProvider) AddStateHash(number int, stateHash string) {
	p.stateHashCache.Add(number, stateHash)
}
