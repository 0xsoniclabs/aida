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

package primer

import (
	"github.com/0xsoniclabs/aida/executor"
	"github.com/0xsoniclabs/aida/executor/extension"
	"github.com/0xsoniclabs/aida/logger"
	"github.com/0xsoniclabs/aida/txcontext"
	"github.com/0xsoniclabs/aida/utils"
)

func MakeEthStateTestDbPrimer(cfg *utils.Config) executor.Extension[txcontext.TxContext] {
	return makeEthStateTestDbPrimer(logger.NewLogger(cfg.LogLevel, "EthStatePrimer"), cfg)
}

func makeEthStateTestDbPrimer(log logger.Logger, cfg *utils.Config) *ethStateTestDbPrimer {
	return &ethStateTestDbPrimer{
		cfg: cfg,
		log: log,
	}
}

type ethStateTestDbPrimer struct {
	extension.NilExtension[txcontext.TxContext]
	cfg *utils.Config
	log logger.Logger
}

func (e ethStateTestDbPrimer) PreBlock(st executor.State[txcontext.TxContext], ctx *executor.Context) error {
	primeCtx := utils.NewPrimeContext(e.cfg, ctx.State, 0, e.log)
	return primeCtx.PrimeStateDB(st.Data.GetInputState(), ctx.State)
}
