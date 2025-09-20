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

package stochastic

import (
	"github.com/0xsoniclabs/aida/logger"
	"github.com/0xsoniclabs/aida/stochastic/recorder"
	"github.com/0xsoniclabs/aida/utils"
	"github.com/urfave/cli/v2"
)

// StochasticGenerateCommand data structure for the record app.
var StochasticGenerateCommand = cli.Command{
	Action:    stochasticGenerateAction,
	Name:      "generate",
	Usage:     "generate uniform state file",
	ArgsUsage: "",
	Flags: []cli.Flag{
		&logger.LogLevelFlag,
		&utils.BlockLengthFlag,
		&utils.SyncPeriodLengthFlag,
		&utils.TransactionLengthFlag,
		&utils.ContractNumberFlag,
		&utils.KeysNumberFlag,
		&utils.ValuesNumberFlag,
		&utils.SnapshotDepthFlag,
	},
	Description: "The stochastic produces an state.json file with uniform parameters",
}

// stochasticGenerateAction produces a state file with uniform parameters
// for stochastic testing.
func stochasticGenerateAction(ctx *cli.Context) error {
	cfg, err := utils.NewConfig(ctx, utils.NoArgs)
	if err != nil {
		return err
	}
	log := logger.NewLogger(cfg.LogLevel, "StochasticGenerate")
	log.Info("Produce uniform stochastic state")
	state, err := recorder.GenerateUniformState(cfg, log)
	if err != nil {
		return err
	}
	if cfg.Output == "" {
		cfg.Output = "./state.json"
	}
	log.Noticef("Write state file %v", cfg.Output)
	if err := state.Write(cfg.Output); err != nil {
		return err
	}
	return nil
}
