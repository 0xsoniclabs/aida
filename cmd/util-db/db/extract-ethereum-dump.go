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

	"github.com/0xsoniclabs/aida/logger"
	substatecontext "github.com/0xsoniclabs/aida/txcontext/substate"
	"github.com/0xsoniclabs/aida/utildb"
	"github.com/0xsoniclabs/aida/utils"
	"github.com/0xsoniclabs/substate/substate"
	"github.com/0xsoniclabs/substate/types"
	"github.com/holiman/uint256"
	"github.com/urfave/cli/v2"
)

var ExtractEthereumDumpCommand = cli.Command{
	Action: extractEthereumDump,
	Name:   "extract-ethereum-dump",
	Usage:  "Extracts WorldState from json into first updateset",
	Flags: []cli.Flag{
		&utils.CarmenSchemaFlag,
		&utils.StateDbImplementationFlag,
		&utils.StateDbVariantFlag,
		&utils.DbTmpFlag,
	},
	Description: `
Extracts WorldState from ethereum dump.json into first updateset.`}

func extractEthereumDump(ctx *cli.Context) error {
	// process arguments and flags
	if ctx.Args().Len() != 1 {
		return fmt.Errorf("ethereum-extract-dump command requires exactly 1 arguments")
	}
	cfg, argErr := utils.NewConfig(ctx, utils.NoArgs)
	if argErr != nil {
		return argErr
	}

	log := logger.NewLogger(cfg.LogLevel, "Ethereum Extract Dump")

	cfg.KeepDb = true
	stateDb, _, err := utils.PrepareStateDB(cfg)
	if err != nil {
		return fmt.Errorf("cannot prepare state-db; %v", err)
	}

	primeCtx := utils.NewPrimeContext(cfg, stateDb, 0, log)

	log.Notice("Load Ethereum dump world state")
	ws, errChan, err := utildb.LoadEthereumDumpWorldState(ctx.Args().Get(0))
	if err != nil {
		return fmt.Errorf("cannot load ethereum dump world state; %v", err)
	}

	ww := make(substate.WorldState)

store:
	for {
		select {
		case err = <-errChan:
			if err != nil {
				return fmt.Errorf("cannot load ethereum dump world state; %v", err)
			}
		case acc, ok := <-ws:
			{
				if !ok {
					break store
				}
				balance := new(uint256.Int)
				if len(acc.Balance) > 0 {
					err = balance.SetFromDecimal(acc.Balance)
					if err != nil {
						return fmt.Errorf("cannot parse balance %s for account %s; %v", acc.Balance, acc.Address, err)
					}
				}

				log.Infof("Processing account %s with balance %s", acc.Address, balance)

				// TODO look up with RPC
				//code = acc.CodeHash
				code := make([]byte, 0)

				account := substate.NewAccount(acc.Nonce, balance, code)
				// TODO look up with RPC
				//account.Storage = acc.Root

				if len(acc.Address) > 0 {
					ww[types.Address([]byte(acc.Address))] = account
					//ww[types.Address{0x1}] = substate.NewAccount(1, uint256.NewInt(1000), nil)
				}

				if len(ww) > 2 {
					//if len(ww) > 10000 {
					log.Infof("Priming state-db with %d accounts", len(ww))
					if err = primeCtx.PrimeStateDB(substatecontext.NewWorldState(ww), stateDb); err != nil {
						return fmt.Errorf("cannot prime state-db; %v", err)
					}
					// reset the world state to avoid too large memory usage
					ww = make(substate.WorldState)
				}
			}
		}
	}

	if len(ww) > 0 {
		if err = primeCtx.PrimeStateDB(substatecontext.NewWorldState(ww), stateDb); err != nil {
			return fmt.Errorf("cannot prime state-db; %v", err)
		}
	}

	return nil
}
