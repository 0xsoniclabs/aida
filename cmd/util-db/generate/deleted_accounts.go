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

package generate

import (
	"errors"
	"fmt"
	"time"

	"github.com/0xsoniclabs/aida/executor"
	"github.com/0xsoniclabs/aida/logger"
	"github.com/0xsoniclabs/aida/state"
	"github.com/0xsoniclabs/aida/state/proxy"
	substatecontext "github.com/0xsoniclabs/aida/txcontext/substate"
	"github.com/0xsoniclabs/aida/utils"
	"github.com/0xsoniclabs/substate/db"
	"github.com/0xsoniclabs/substate/substate"
	substatetypes "github.com/0xsoniclabs/substate/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/urfave/cli/v2"
)

const channelSize = 100000 // size of deletion channel

var generateDeletedAccountsCommand = cli.Command{
	Action:    generateDeletedAccountsAction,
	Name:      "deleted-accounts",
	Usage:     "executes full state transitions and record suicided and resurrected accounts",
	ArgsUsage: "<blockNumFirst> <blockNumLast>",
	Flags: []cli.Flag{
		&utils.WorkersFlag,
		&utils.AidaDbFlag,
		&utils.ChainIDFlag,
		&utils.DeletionDbFlag,
		&utils.CpuProfileFlag,
		&logger.LogLevelFlag,
		&utils.SubstateEncodingFlag,
	},
	Description: `
The util-db gen-deleted-accounts command requires two arguments:
<blockNumFirst> <blockNumLast>
<blockNumFirst> and <blockNumLast> are the first and
last block of the inclusive range of blocks to replay transactions.`,
}

func generateDeletedAccountsAction(ctx *cli.Context) (finalErr error) {
	cfg, err := utils.NewConfig(ctx, utils.BlockRangeArgs)
	if err != nil {
		return err
	}

	if !ctx.IsSet(utils.DeletionDbFlag.Name) {
		return fmt.Errorf("you need to specify where you want deletion-db to save (--deletion-db)")
	}

	sdb, err := db.NewReadOnlySubstateDB(cfg.AidaDb)
	if err != nil {
		return fmt.Errorf("cannot open aida-db; %w", err)
	}
	defer func() {
		finalErr = errors.Join(finalErr, sdb.Close())
	}()

	ddb, err := db.NewDefaultDestroyedAccountDB(cfg.DeletionDb)
	if err != nil {
		return err
	}
	defer func() {
		finalErr = errors.Join(finalErr, ddb.Close())
	}()

	return generateDeletedAccounts(cfg, sdb, ddb, cfg.First, cfg.Last)
}

// generateDeletedAccounts replays transactions and record self-destructed accounts and resurrected accounts.
func generateDeletedAccounts(cfg *utils.Config, sdb db.SubstateDB, ddb db.DestroyedAccountDB, firstBlock uint64, lastBlock uint64) error {
	var err error

	err = utils.StartCPUProfile(cfg)
	if err != nil {
		return err
	}
	defer utils.StopCPUProfile(cfg)

	log := logger.NewLogger(cfg.LogLevel, "Generate Deleted Accounts")
	log.Noticef("Generate deleted accounts from block %v to block %v", firstBlock, lastBlock)

	start := time.Now()
	lastSec := time.Since(start).Seconds()
	txCount := uint64(0)
	lastTxCount := uint64(0)
	var deleteHistory = make(map[common.Address]bool)
	// explicitly set to nil to release memory as soon as possible
	defer func() {
		deleteHistory = nil
	}()

	iter := sdb.NewSubstateIterator(int(firstBlock), cfg.Workers)
	defer iter.Release()

	processor, err := executor.MakeTxProcessor(cfg)
	if err != nil {
		return err
	}

	for iter.Next() {
		tx := iter.Value()
		if tx.Block > lastBlock {
			break
		}

		if tx.Transaction < utils.PseudoTx {
			err = recordDeletedAccountsFromTx(tx, processor, ddb, &deleteHistory, cfg)
			if err != nil {
				return err
			}

			txCount++
			sec := time.Since(start).Seconds()
			diff := sec - lastSec
			if diff >= 30 {
				numTx := txCount - lastTxCount
				lastTxCount = txCount
				log.Infof("aida-vm: gen-del-acc: Elapsed time: %.0f s, at block %v (~%.1f Tx/s)", sec, tx.Block, float64(numTx)/diff)
				lastSec = sec
			}
		}
	}
	return err
}

// recordDeletedAccountsFromTx process a transaction substate then records
// self-destructed accounts and resurrected accounts to a database.
func recordDeletedAccountsFromTx(
	tx *substate.Substate,
	processor *executor.TxProcessor,
	ddb db.DestroyedAccountDB,
	deleteHistory *map[common.Address]bool,
	cfg *utils.Config,
) (finalErr error) {
	ch := make(chan proxy.ContractLiveliness, channelSize)
	var statedb state.StateDB
	var err error
	ss := substatecontext.NewTxContext(tx)

	chainCfg, err := cfg.GetChainConfig("")
	if err != nil {
		return fmt.Errorf("cannot get chain config: %w", err)
	}

	conduit := state.NewChainConduit(utils.IsEthereumNetwork(cfg.ChainID), chainCfg)
	statedb, err = state.MakeOffTheChainStateDB(ss.GetInputState(), tx.Block, conduit)
	if err != nil {
		return err
	}

	defer func() {
		finalErr = errors.Join(finalErr, statedb.Close())
	}()

	//wrapper
	statedb = proxy.NewDeletionProxy(statedb, ch, cfg.LogLevel)

	_, err = processor.ProcessTransaction(statedb, int(tx.Block), tx.Transaction, ss)
	if err != nil {
		return err
	}

	close(ch)
	return writeDeletedAccounts(ddb, tx, ch, deleteHistory)
}

// writeDeletedAccounts reads contracts which were suicided or created and adds them to lists
func writeDeletedAccounts(
	ddb db.DestroyedAccountDB,
	ss *substate.Substate,
	ch chan proxy.ContractLiveliness,
	deleteHistory *map[common.Address]bool,
) error {
	deletedAccounts := make(map[common.Address]bool)
	resurrectedAccounts := make(map[common.Address]bool)
	for contract := range ch {
		addr := contract.Addr
		if contract.IsDeleted {
			// if a contract was resurrected before suicided in the same tx,
			// only keep the last action.
			delete(resurrectedAccounts, addr)
			(*deleteHistory)[addr] = true // meta list
			deletedAccounts[addr] = true
		} else {
			// if a contract was suicided before resurrected in the same tx,
			// only keep the last action.
			delete(deletedAccounts, addr)
			// an account is considered as resurrected if it was recently deleted.
			if recentlyDeleted, found := (*deleteHistory)[addr]; found && recentlyDeleted {
				(*deleteHistory)[addr] = false
				resurrectedAccounts[addr] = true
			}
		}
	}

	if len(deletedAccounts)+len(resurrectedAccounts) > 0 {
		// if transaction completed successfully, put destroyed accounts
		// and resurrected accounts to a database
		if ss.Result.Status == types.ReceiptStatusSuccessful {
			var destroyed, resurrected []substatetypes.Address
			for addr := range deletedAccounts {
				destroyed = append(destroyed, substatetypes.Address(addr))
			}

			for addr := range resurrectedAccounts {
				resurrected = append(resurrected, substatetypes.Address(addr))
			}
			err := ddb.SetDestroyedAccounts(ss.Block, ss.Transaction, destroyed, resurrected)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
