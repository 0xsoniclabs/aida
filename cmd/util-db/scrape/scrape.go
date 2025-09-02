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

package scrape

import (
	"context"
	"fmt"
	"os"

	"github.com/0xsoniclabs/aida/logger"
	"github.com/0xsoniclabs/aida/utils"
	"github.com/0xsoniclabs/substate/db"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/urfave/cli/v2"
)

var Command = cli.Command{
	Action:    scrapeAction,
	Name:      "scrape",
	Usage:     "Stores state hashes into TargetDb for given range",
	ArgsUsage: "<blockNumFirst> <blockNumLast>",
	Flags: []cli.Flag{
		&utils.TargetDbFlag,
		&utils.ChainIDFlag,
		&utils.ClientDbFlag,
		&logger.LogLevelFlag,
	},
}

// scrapeAction stores state hashes into Target for given range
func scrapeAction(ctx *cli.Context) error {
	cfg, argErr := utils.NewConfig(ctx, utils.BlockRangeArgs)
	if argErr != nil {
		return argErr
	}

	log := logger.NewLogger(cfg.LogLevel, "UtilDb-Scrape")
	log.Infof("Scraping for range %d-%d", cfg.First, cfg.Last)

	database, err := db.NewDefaultBaseDB(cfg.TargetDb)
	if err != nil {
		return fmt.Errorf("error opening stateHash leveldb %s: %v", cfg.TargetDb, err)
	}
	defer database.Close()

	err = StateAndBlockHashScraper(ctx.Context, cfg.ChainID, cfg.ClientDb, database, cfg.First, cfg.Last, log)
	if err != nil {
		return err
	}

	log.Infof("Scraping finished")
	return nil
}

// StateAndBlockHashScraper scrapes state and block hashes from a node and saves them to a leveldb database
func StateAndBlockHashScraper(ctx context.Context, chainId utils.ChainID, clientDb string, db db.BaseDB, firstBlock, lastBlock uint64, log logger.Logger) error {
	client, err := getClient(ctx, chainId, clientDb, log)
	if err != nil {
		return err
	}
	defer client.Close()

	var i = firstBlock

	// If firstBlock is 0, we need to get the state root for block 1 and save it as the state root for block 0
	// this is because the correct state root for block 0 is not available from the rpc node (at least in fantom mainnet and testnet)
	if firstBlock == 0 {
		block, err := getBlockByNumber(client, "0x1")
		if err != nil {
			return err
		}

		if block == nil {
			return fmt.Errorf("block 1 not found")
		}

		err = utils.SaveStateRoot(db, "0x0", block["stateRoot"].(string))
		if err != nil {
			return err
		}
		err = utils.SaveBlockHash(db, "0x1", block["hash"].(string))
		if err != nil {
			return err
		}
		i++
	}

	for ; i <= lastBlock; i++ {
		blockNumber := fmt.Sprintf("0x%x", i)
		block, err := getBlockByNumber(client, blockNumber)
		if err != nil {
			return err
		}

		if block == nil {
			return fmt.Errorf("block %d not found", i)
		}

		err = utils.SaveStateRoot(db, blockNumber, block["stateRoot"].(string))
		if err != nil {
			return err
		}
		err = utils.SaveBlockHash(db, blockNumber, block["hash"].(string))
		if err != nil {
			return err
		}

		if i%10000 == 0 {
			log.Infof("Scraping block %d done!\n", i)
		}
	}

	return nil
}

// getClient returns a rpc/ipc client
func getClient(ctx context.Context, chainId utils.ChainID, clientDb string, log logger.Logger) (*rpc.Client, error) {
	var client *rpc.Client
	var err error

	// try both sonic and geth ipcs
	ipcPaths := []string{
		clientDb + "/sonic.ipc",
		clientDb + "/geth.ipc",
	}
	for _, ipcPath := range ipcPaths {
		_, errIpc := os.Stat(ipcPath)
		if errIpc == nil {
			// ipc file exists
			client, err = rpc.DialIPC(ctx, ipcPath)
			if err != nil {
				return nil, fmt.Errorf("failed to connect to IPC at %s: %v", ipcPath, err)
			}
			log.Infof("Connected to IPC at %s", ipcPath)
			return client, err
		}
	}

	// if ipc file does not exist, try to connect to RPC
	var provider string
	provider, err = utils.GetProvider(chainId)
	if err != nil {
		return nil, err
	}
	client, err = rpc.Dial(provider)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to the RPC client at %s: %v", provider, err)
	}
	log.Infof("Connected to RPC at %s", provider)
	return client, nil
}

//go:generate mockgen -source scrape.go -destination scrape_mock.go -package scrape

type IRpcClient interface {
	Call(result interface{}, method string, args ...interface{}) error
}

// getBlockByNumber get block from the rpc node
func getBlockByNumber(client IRpcClient, blockNumber string) (map[string]interface{}, error) {
	var block map[string]interface{}
	err := client.Call(&block, "eth_getBlockByNumber", blockNumber, false)
	if err != nil {
		return nil, fmt.Errorf("failed to get block %s: %v", blockNumber, err)
	}
	return block, nil
}
