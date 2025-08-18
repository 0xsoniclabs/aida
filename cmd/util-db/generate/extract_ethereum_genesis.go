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

package generate

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/0xsoniclabs/substate/substate"
	"github.com/cockroachdb/errors"
	"github.com/holiman/uint256"

	"github.com/0xsoniclabs/aida/logger"
	"github.com/0xsoniclabs/aida/utils"
	"github.com/0xsoniclabs/substate/db"
	substatetypes "github.com/0xsoniclabs/substate/types"
	"github.com/0xsoniclabs/substate/updateset"
	"github.com/urfave/cli/v2"
)

var extractEthereumGenesisCommand = cli.Command{
	Action: extractEthereumGenesisAction,
	Name:   "extract-ethereum-genesis",
	Usage:  "Extracts WorldState from json into first updateset",
	Flags: []cli.Flag{
		&utils.ChainIDFlag,
		&utils.UpdateDbFlag,
		&logger.LogLevelFlag,
	},
	Description: `
Extracts WorldState from ethereum genesis.json into first updateset.`,
}

func extractEthereumGenesisAction(ctx *cli.Context) (finalErr error) {
	// process arguments and flags
	if ctx.Args().Len() != 1 {
		return fmt.Errorf("ethereum-update command requires exactly 1 arguments")
	}
	cfg, argErr := utils.NewConfig(ctx, utils.NoArgs)
	if argErr != nil {
		return argErr
	}
	log := logger.NewLogger(cfg.LogLevel, "Ethereum Update")

	log.Notice("Load Ethereum initial world state")
	ws, err := loadEthereumGenesisWorldState(ctx.Args().Get(0))
	if err != nil {
		return err
	}

	udb, err := db.NewDefaultUpdateDB(cfg.UpdateDb)
	if err != nil {
		return err
	}
	defer func() {
		finalErr = errors.Join(finalErr, udb.Close())
	}()

	log.Noticef("PutUpdateSet(0, %v, []common.Address{})", ws)

	return udb.PutUpdateSet(&updateset.UpdateSet{WorldState: ws, Block: 0}, make([]substatetypes.Address, 0))
}

type GenesisData struct {
	Alloc map[string]struct {
		Balance string            `json:"balance,omitempty"`
		Nonce   string            `json:"nonce,omitempty"`
		Code    string            `json:"code,omitempty"`
		Storage map[string]string `json:"storage,omitempty"`
	} `json:"alloc"`
}

// loadEthereumGenesisWorldState loads opera initial world state from worldstate-db as WorldState
func loadEthereumGenesisWorldState(genesisPath string) (substate.WorldState, error) {
	var jsData GenesisData
	// Read the content of the JSON file
	jsonData, err := os.ReadFile(genesisPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read genesis file: %account", err)
	}

	err = json.Unmarshal(jsonData, &jsData)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal genesis file: %account", err)
	}

	ssAccounts := make(substate.WorldState)

	// loop over all the accounts
	for addr, account := range jsData.Alloc {
		// Convert the string key back to a common.Address
		address := substatetypes.HexToAddress(addr)

		balance := uint256.MustFromHex(account.Balance)
		var nonce uint64
		if len(account.Nonce) > 2 {
			nonce, err = strconv.ParseUint(strings.TrimPrefix(account.Nonce, "0x"), 10, 64)
			if err != nil {
				return nil, fmt.Errorf("failed to parse nonce: %account", err)
			}
		}

		var code []byte
		if len(account.Code) > 2 {
			code, err = hex.DecodeString(strings.TrimPrefix(account.Code, "0x"))
			if err != nil {
				return nil, fmt.Errorf("failed to decode code: %account", err)
			}
		}

		acc := substate.NewAccount(nonce, balance, code)

		if len(account.Storage) > 0 {
			for key, value := range account.Storage {
				decodedKey, err := hex.DecodeString(strings.TrimPrefix(key, "0x"))
				if err != nil {
					return nil, fmt.Errorf("failed to decode storage key: %account", err)
				}
				decodedValue, err := hex.DecodeString(strings.TrimPrefix(value, "0x"))
				if err != nil {
					return nil, fmt.Errorf("failed to decode storage value: %account", err)
				}
				acc.Storage[substatetypes.BytesToHash(decodedKey)] = substatetypes.BytesToHash(decodedValue)
			}
		}

		ssAccounts[address] = acc
	}

	return ssAccounts, err
}
