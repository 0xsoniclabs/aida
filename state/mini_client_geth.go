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

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/prque"

	mc_aida "github.com/0xsoniclabs/mini-client/pkg/aida"
)

// MakeMiniClientGethStateDB constructs a StateDB that exercises the
// same trie configuration mini-client's internal/executor uses at
// validator startup: HashDB scheme (not the v1.14 path scheme), nil
// snapshot tree, mini-client's default clean-cache + leveldb knobs.
//
// The signature deliberately mirrors MakeGethStateDB so call sites
// (registry, factory tables, harness configs) only need to swap the
// constructor name. The returned StateDB is wrapped in the same
// gethStateDB shell vanilla geth uses, so every proxy in
// state/proxy/ (logger, profiler, shadow, cache) works unchanged.
//
// `variant` is reserved for future use; an empty string is the only
// currently-accepted value. `chainConduit` is forwarded directly.
func MakeMiniClientGethStateDB(directory, variant string, rootHash common.Hash, isArchiveMode bool, chainConduit *ChainConduit) (StateDB, error) {
	if variant != "" {
		return nil, fmt.Errorf("unknown mini-client-geth variant: %v", variant)
	}
	bundle, err := mc_aida.NewGethStateDB(mc_aida.GethStateDBConfig{
		DataDir:   directory,
		RootHash:  rootHash,
		IsArchive: isArchiveMode,
	})
	if err != nil {
		return nil, fmt.Errorf("mini-client geth state init: %w", err)
	}
	return &gethStateDB{
		db:            bundle.StateDB,
		evmState:      bundle.EvmState,
		stateRoot:     bundle.RootHash,
		triegc:        prque.New[uint64, common.Hash](nil),
		isArchiveMode: isArchiveMode,
		chainConduit:  chainConduit,
		backend:       bundle.Backend,
	}, nil
}
