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

package validator

import (
	"github.com/0xsoniclabs/aida/executor"
	"github.com/0xsoniclabs/aida/executor/extension"
	"github.com/0xsoniclabs/aida/logger"
	"github.com/0xsoniclabs/aida/txcontext"
	"github.com/0xsoniclabs/aida/utils"
)

// ethereumLfvmBlocksExceptions LFVM uses a uint16 program counter with a range from 0 to 65535.
// Starting with the Shanghai revision and eip-3860 this was fixed
// only post alloc is diverging for these ethereum block exceptions, so it needs to be overwritten
var ethereumLfvmBlockExceptions = map[utils.ChainID]map[int]struct{}{
	utils.EthereumChainID: {13803456: {}, 14340503: {}, 14953169: {}, 15025981: {}, 15427798: {}, 15445161: {}, 15445481: {}},
	//utils.SepoliaChainID:  {2259736: {}, 2259718: {}, 2259775: {}, 2261404: {}, 2261423: {}, 2267647: {}, 2299256: {}, 2513443: {}, 2612238: {}, 2656617: {}, 2825745: {}},
}

// MakeEthereumDbPostTransactionUpdater creates an extension which fixes Ethereum exceptions in LiveDB
func MakeEthereumDbPostTransactionUpdater(cfg *utils.Config) executor.Extension[txcontext.TxContext] {
	log := logger.NewLogger(cfg.LogLevel, "Ethereum-Exception-Updater")

	return makeEthereumDbPostTransactionUpdater(cfg, log)
}

func makeEthereumDbPostTransactionUpdater(cfg *utils.Config, log logger.Logger) executor.Extension[txcontext.TxContext] {
	if !utils.IsEthereumNetwork(cfg.ChainID) || cfg.VmImpl != "lfvm" {
		return extension.NilExtension[txcontext.TxContext]{}
	}

	return &ethereumDbPostTransactionUpdater{
		cfg: cfg,
		log: log,
	}
}

type ethereumDbPostTransactionUpdater struct {
	extension.NilExtension[txcontext.TxContext]
	cfg *utils.Config
	log logger.Logger
}

// PostTransaction fixes OutputAlloc ethereum exceptions in given substate
func (v *ethereumDbPostTransactionUpdater) PostTransaction(state executor.State[txcontext.TxContext], ctx *executor.Context) error {
	if _, ok := ethereumLfvmBlockExceptions[v.cfg.ChainID][state.Block]; ok {
		return updateStateDbOnEthereumChain(state.Data.GetOutputState(), ctx.State, true)
	}
	return nil
}

// PreRun informs the user that ethereumDbPostTransactionUpdater is enabled.
func (v *ethereumDbPostTransactionUpdater) PreRun(executor.State[txcontext.TxContext], *executor.Context) error {
	v.log.Warning("Ethereum exception post transaction updater is enabled.")

	return nil
}
