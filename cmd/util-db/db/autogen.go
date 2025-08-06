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

package db

import (
	"fmt"
	"github.com/0xsoniclabs/aida/config"
	"github.com/0xsoniclabs/aida/config/chainid"

	"github.com/0xsoniclabs/aida/logger"
	"github.com/0xsoniclabs/aida/utildb"
	"github.com/0xsoniclabs/aida/utils"
	"github.com/urfave/cli/v2"
)

// AutoGenCommand generates aida-db patches and handles second opera for event generation
var AutoGenCommand = cli.Command{
	Action: autogen,
	Name:   "autogen",
	Usage:  "autogen generates aida-db periodically",
	Flags: []cli.Flag{
		&utils.AidaDbFlag,
		&chainid.ChainIDFlag,
		&utils.ClientDbFlag,
		&utils.GenesisFlag,
		&utils.DbTmpFlag,
		&utils.OperaBinaryFlag,
		&utils.OutputFlag,
		&utils.TargetEpochFlag,
		&utils.UpdateBufferSizeFlag,
		&utils.WorkersFlag,
		&logger.LogLevelFlag,
	},
	Description: `
AutoGen generates aida-db patches and handles second opera for event generation. Generates event file, which is supplied into doGenerations to create aida-db patch.
`,
}

// autogen command is used to record/update aida-db periodically
func autogen(ctx *cli.Context) error {
	cfg, err := config.NewConfig(ctx, config.NoArgs)
	if err != nil {
		return err
	}

	locked, err := utildb.GetLock(cfg)
	if err != nil {
		return err
	}
	if locked != "" {
		return fmt.Errorf("GENERATION BLOCKED: autogen failed in last run; %v", locked)
	}

	var g *Generator
	var ok bool
	g, ok, err = prepareAutogen(ctx, cfg)
	if err != nil {
		return fmt.Errorf("cannot start autogen; %v", err)
	}
	if !ok {
		g.Log.Warningf("supplied targetEpoch %d is already reached; latest generated epoch %d", g.TargetEpoch, g.Opera.FirstEpoch-1)
		return nil
	}

	err = utildb.AutogenRun(cfg, g)
	if err != nil {
		errLock := utildb.SetLock(cfg, err.Error())
		if errLock != nil {
			return fmt.Errorf("%v; %v", errLock, err)
		}
	}
	return err
}

// PrepareAutogen initializes a generator object, opera binary and adjust target range
func prepareAutogen(ctx *cli.Context, cfg *config.Config) (*Generator, bool, error) {
	// this explicit overwrite is necessary at first autogen run,
	// in later runs the paths are correctly set in adjustMissingConfigValues
	config.OverwriteDbPathsByAidaDb(cfg)

	g, err := NewGenerator(ctx, cfg)
	if err != nil {
		return nil, false, err
	}

	err = g.Opera.init()
	if err != nil {
		return nil, false, err
	}

	// user specified targetEpoch
	if cfg.TargetEpoch > 0 {
		g.TargetEpoch = cfg.TargetEpoch
	} else {
		err = g.calculatePatchEnd()
		if err != nil {
			return nil, false, err
		}
	}

	g.AidaDb.Close()
	// start epoch is last epoch + 1
	if g.Opera.FirstEpoch > g.TargetEpoch {
		return g, false, nil
	}
	return g, true, nil
}
