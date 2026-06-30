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

	mc_aida "github.com/0xsoniclabs/mini-client/pkg/aida"
)

// MakeMiniClientSonicStateDB creates the mini-client sonic SUT.
//
// Implementation note: mini-client's sonic backend is structurally a
// Carmen S5 LiveDB / ArchiveDB with the go-file variant (it opens
// carmen.GetCarmenGoS5WithoutArchiveConfiguration in
// internal/executor/evm/statedb). Aida already vendors Carmen and
// exposes MakeCarmenStateDB with exactly that variant/schema, so the
// SUT delegates straight to it with mini-client's defaults pulled
// from mc_aida (SonicCarmenVariant / SonicCarmenSchema /
// SonicCarmenArchiveVariant). This avoids the cgo
// `multiple definition of carmen_keccak256` linker error that would
// otherwise happen if aida linked both its own vendored Carmen and a
// second copy reached through mini-client/internal/executor.
//
// The SUT identity ("mini-client-sonic") is preserved so aida can
// attribute the replay to mini-client's expected production
// configuration even though the in-process StateDB is the same
// Carmen aida already supports.
//
// chainConduit is accepted for signature symmetry with
// MakeMiniClientGethStateDB but is unused — Carmen does not consume
// chain config the way go-ethereum's geth StateDB does.
func MakeMiniClientSonicStateDB(
	directory string,
	variant string,
	rootHash common.Hash,
	isArchiveMode bool,
	chainConduit *ChainConduit,
	liveDbCacheSize int,
	archiveCacheSize int,
	checkpointInterval int,
	checkpointPeriod int,
) (StateDB, error) {
	if variant != "" && variant != mc_aida.SonicCarmenVariant {
		return nil, fmt.Errorf("unknown mini-client-sonic variant: %v (only %q is supported)", variant, mc_aida.SonicCarmenVariant)
	}
	if err := mc_aida.ValidateSonicConfig(mc_aida.SonicConfig{
		Dir:     directory,
		Archive: isArchiveMode,
	}); err != nil {
		return nil, err
	}
	_ = rootHash
	_ = chainConduit
	if !mc_aida.SonicAvailable() {
		return nil, fmt.Errorf("mini-client sonic SUT is not available (mc_aida.SonicAvailable() returned false)")
	}

	archive := mc_aida.SonicCarmenArchiveVariant
	if !isArchiveMode {
		archive = "none"
	}

	return MakeCarmenStateDB(
		directory,
		mc_aida.SonicCarmenVariant,
		mc_aida.SonicCarmenSchema,
		archive,
		liveDbCacheSize,
		archiveCacheSize,
		checkpointInterval,
		checkpointPeriod,
	)
}
