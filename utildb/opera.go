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

package utildb

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"

	"github.com/0xsoniclabs/aida/logger"
	"github.com/0xsoniclabs/aida/utils"
	"github.com/urfave/cli/v2"
)

// aidaOpera represents running opera as a subprocess
type aidaOpera struct {
	firstBlock, lastBlock uint64
	FirstEpoch, lastEpoch uint64
	ctx                   *cli.Context
	cfg                   *utils.Config
	log                   logger.Logger
	isNew                 bool
}

// newAidaOpera returns new instance of Opera
func newAidaOpera(ctx *cli.Context, cfg *utils.Config, log logger.Logger) *aidaOpera {
	return &aidaOpera{
		ctx: ctx,
		cfg: cfg,
		log: log,
	}
}

// init aidaOpera by executing command to start (and stop) opera and preparing dump context
func (opera *aidaOpera) init() error {
	// TODO resolve dependencies
	panic("feature not supported")
	/*
		var err error

		_, err = os.Stat(opera.cfg.ClientDb)
		if os.IsNotExist(err) {
			opera.isNew = true

			opera.log.Noticef("Initialising opera from genesis")

			// previous opera database isn't used - generate new one from genesis
			err = opera.initFromGenesis()
			if err != nil {
				return fmt.Errorf("cannot init opera from gensis; %v", err)
			}

			// create tmpDir for worldstate
			var tmpDir string
			tmpDir, err = createTmpDir(opera.cfg)
			if err != nil {
				return fmt.Errorf("cannot create tmp dir; %v", err)
			}
			opera.cfg.WorldStateDb = filepath.Join(tmpDir, "worldstate")

			// dumping the MPT into world state
			if err = opera.prepareDumpCliContext(); err != nil {
				return fmt.Errorf("cannot prepare dump; %v", err)
			}
		}

		// get first block and epoch
		// running this command before starting opera results in getting first block and epoch on which opera starts
		err = opera.getOperaBlockAndEpoch(true)
		if err != nil {
			return fmt.Errorf("cannot retrieve block from existing opera database %v; %v", opera.cfg.ClientDb, err)
		}

		opera.log.Noticef("Opera block from last run is: %v", opera.firstBlock)

		// starting generation one block later
		opera.firstBlock += 1
		opera.FirstEpoch += 1
	*/
	return nil
}

func createTmpDir(cfg *utils.Config) (string, error) {
	// create a temporary working directory
	fName, err := os.MkdirTemp(cfg.DbTmp, "aida_db_tmp_*")
	if err != nil {
		return "", fmt.Errorf("failed to create a temporary directory. %v", err)
	}

	return fName, nil
}

// initFromGenesis file
func (opera *aidaOpera) initFromGenesis() error {
	cmd := exec.Command(getOperaBinary(opera.cfg), "--datadir", opera.cfg.ClientDb, "--genesis", opera.cfg.Genesis,
		"--exitwhensynced.epoch=0", "--cache", strconv.Itoa(opera.cfg.Cache), "--db.preset=legacy-ldb", "--maxpeers=0")

	err := runCommand(cmd, nil, nil, opera.log)
	if err != nil {
		return fmt.Errorf("load opera genesis; %v", err.Error())
	}

	return nil
}

// getOperaBlockAndEpoch retrieves current block of opera head
func (opera *aidaOpera) getOperaBlockAndEpoch(isFirst bool) error {
	// TODO resolve dependencies
	panic("feature not supported")
	/*
		operaPath := filepath.Join(opera.cfg.ClientDb, "/chaindata/leveldb-fsh/")
		store, err := wsOpera.Connect("ldb", operaPath, "main")
		if err != nil {
			return err
		}
		defer wsOpera.MustCloseStore(store)

		_, blockNumber, epochNumber, err := wsOpera.LatestStateRoot(store)
		if err != nil {
			return fmt.Errorf("state root not found; %v", err)
		}

		if blockNumber < 1 {
			return fmt.Errorf("opera; block number not found; %v", err)
		}

		// we are assuming that we are at brink of epochs
		// in this special case epochNumber is already one number higher
		epochNumber -= 1

		if isFirst {
			opera.firstBlock = blockNumber
			opera.FirstEpoch = epochNumber
		} else {
			opera.lastBlock = blockNumber
			opera.lastEpoch = epochNumber
		}
	*/
	return nil
}

// prepareDumpCliContext prepares cli context for dumping the MPT into world state
func (opera *aidaOpera) prepareDumpCliContext() error {
	// TODO: resolve dependencies
	/*
		// Dump uses cfg.ClientDb and overwrites it therefore the original value needs to be saved and retrieved after DumpState
		tmpSaveDbPath := opera.cfg.ClientDb
		defer func() {
			opera.cfg.ClientDb = tmpSaveDbPath
		}()
		opera.cfg.ClientDb = filepath.Join(opera.cfg.ClientDb, "chaindata/leveldb-fsh/")
		opera.cfg.DbVariant = "ldb"
		err := state.DumpState(opera.ctx, opera.cfg)
		if err != nil {
			return err
		}
	*/
	return nil
}
