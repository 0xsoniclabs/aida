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

// ethereumLfvmBlockExceptions LFVM uses a uint16 program counter with a range from 0 to 65535.
// Starting with the Shanghai revision and eip-3860 this was fixed
// only post alloc is diverging for these block exceptions, so it needs to be overwritten
var ethereumLfvmBlockExceptions = map[utils.ChainID]map[int]struct{}{
	utils.EthereumChainID: {
		10880015: {}, 10880604: {}, 10880608: {}, 13142297: {}, 13163624: {}, 13163650: {}, 13656943: {}, 13658320: {},
		13715268: {}, 13803456: {}, 13810899: {}, 13815453: {}, 13854983: {}, 13854989: {}, 13854993: {}, 13854995: {},
		13856623: {}, 14171562: {}, 14340503: {}, 14369546: {}, 14643356: {}, 14729777: {}, 14740700: {}, 14764663: {},
		14764876: {}, 14771243: {}, 14785590: {}, 14791002: {}, 14818939: {}, 14849270: {}, 14953169: {}, 15025981: {},
		15104710: {}, 15120267: {}, 15125210: {}, 15140417: {}, 15245196: {}, 15328635: {}, 15343934: {}, 15344082: {},
		15344120: {}, 15344129: {}, 15344131: {}, 15344416: {}, 15344426: {}, 15344608: {}, 15344639: {}, 15344645: {},
		15349414: {}, 15349420: {}, 15349442: {}, 15349453: {}, 15349459: {}, 15349507: {}, 15349509: {}, 15349511: {},
		15349513: {}, 15349581: {}, 15349586: {}, 15349629: {}, 15349630: {}, 15349636: {}, 15349640: {}, 15349642: {},
		15349650: {}, 15349652: {}, 15349654: {}, 15349668: {}, 15349673: {}, 15349679: {}, 15349680: {}, 15349682: {},
		15349684: {}, 15349686: {}, 15349705: {}, 15349707: {}, 15349709: {}, 15349712: {}, 15349721: {}, 15349749: {},
		15350091: {}, 15350201: {}, 15350237: {}, 15350255: {}, 15350264: {}, 15350319: {}, 15350346: {}, 15350371: {},
		15350418: {}, 15350425: {}, 15350440: {}, 15350447: {}, 15350485: {}, 15350502: {}, 15350516: {}, 15350577: {},
		15350579: {}, 15350620: {}, 15350621: {}, 15350641: {}, 15350681: {}, 15350713: {}, 15350754: {}, 15350767: {},
		15350823: {}, 15427798: {}, 15428029: {}, 15445161: {}, 15445481: {}, 15512207: {}, 15514257: {}, 15514260: {},
		15598155: {}, 15620621: {}, 15648330: {}, 15819645: {}, 15832920: {}, 15833466: {}, 15833599: {}, 15840085: {},
		15840211: {}, 15840338: {}, 15893119: {}, 15994188: {}, 16150497: {}, 16328317: {}, 16505992: {}, 16568023: {},
		16592162: {}, 16782381: {}, 16832475: {}, 16832667: {}, 16832676: {}, 16832890: {}, 16840977: {}, 16844950: {},
		16845000: {}, 16881016: {}},
	utils.SepoliaChainID: {
		2259736: {}, 2259718: {}, 2259775: {}, 2261404: {}, 2261423: {}, 2267647: {}, 2299256: {},
		2513443: {}, 2612238: {}, 2656617: {}, 2825745: {},
	},
	utils.TestnetChainID: {
		4805316: {}, 4805731: {}, 8114395: {}, 8151297: {}, 8151512: {}, 8151721: {}, 8152084: {}, 8152229: {},
		8152313: {}, 10162070: {}, 10188748: {}, 10203245: {}, 10203269: {}, 14025784: {}, 14025973: {}, 14048179: {},
		14051667: {}, 14051693: {}, 14053489: {}, 14053495: {}, 14055433: {}, 14055450: {}, 14055560: {}, 14055809: {},
	},
}

// MakeEthereumDbPostTransactionUpdater creates an extension which fixes Ethereum exceptions in LiveDB
func MakeEthereumDbPostTransactionUpdater(cfg *utils.Config) executor.Extension[txcontext.TxContext] {
	log := logger.NewLogger(cfg.LogLevel, "Ethereum-Exception-Updater")

	return makeEthereumDbPostTransactionUpdater(cfg, log)
}

func makeEthereumDbPostTransactionUpdater(cfg *utils.Config, log logger.Logger) executor.Extension[txcontext.TxContext] {
	if cfg.VmImpl != "lfvm" {
		return extension.NilExtension[txcontext.TxContext]{}
	}

	// check if the chainID has at least one exception - this is mainly to avoid unnecessary extension creation for sonic mainnet
	_, ok := ethereumLfvmBlockExceptions[cfg.ChainID]
	if !ok {
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
