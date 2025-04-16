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
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"

	"github.com/0xsoniclabs/aida/logger"
	"github.com/0xsoniclabs/aida/utils"
	"github.com/0xsoniclabs/substate/db"
	"github.com/0xsoniclabs/substate/substate"
	substatetypes "github.com/0xsoniclabs/substate/types"
	"github.com/0xsoniclabs/substate/updateset"
	"github.com/holiman/uint256"
	"github.com/urfave/cli/v2"
)

var ExtractEthereumGenesisCommand = cli.Command{
	Action: extractEthereumGenesis,
	Name:   "extract-ethereum-genesis",
	Usage:  "Extracts WorldState from json into first updateset",
	Flags: []cli.Flag{
		&utils.ChainIDFlag,
		&utils.UpdateDbFlag,
		&logger.LogLevelFlag,
	},
	Description: `
Extracts WorldState from ethereum genesis.json into first updateset.`}

func extractEthereumGenesis(ctx *cli.Context) error {
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
	defer udb.Close()

	log.Noticef("PutUpdateSet(0, %v, []common.Address{})", ws)

	return udb.PutUpdateSet(&updateset.UpdateSet{WorldState: ws, Block: 0}, make([]substatetypes.Address, 0))
}

// loadEthereumGenesisWorldState loads opera initial world state from worldstate-db as WorldState
func loadEthereumGenesisWorldState(genesis string) (substate.WorldState, error) {
	var jsData map[string]interface{}
	// Read the content of the JSON file
	jsonData, err := ioutil.ReadFile(genesis)
	if err != nil {
		return nil, fmt.Errorf("failed to read genesis file: %v", err)
	}

	// Unmarshal JSON data
	if err := json.Unmarshal(jsonData, &jsData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal genesis file: %v", err)
	}

	// get field alloc
	alloc, ok := jsData["alloc"]
	if !ok {
		return nil, fmt.Errorf("failed to get alloc field from genesis file")
	}

	ssAccounts := make(substate.WorldState)

	// loop over all the accounts
	for k, v := range alloc.(map[string]interface{}) {
		// Convert the string key back to a common.Address
		address := substatetypes.HexToAddress(k)

		balance := uint256.MustFromHex(v.(map[string]interface{})["balance"].(string))
		var nonce uint64
		nonceS, ok := v.(map[string]interface{})["nonce"].(string)
		if ok {
			nonce, err = strconv.ParseUint(strings.TrimPrefix(nonceS, "0x"), 10, 64)
			if err != nil {
				return nil, fmt.Errorf("failed to parse nonce: %v", err)
			}
		}

		var code []byte
		codeS, ok := v.(map[string]interface{})["code"].(string)
		if ok {
			code, err = hex.DecodeString(strings.TrimPrefix(codeS, "0x"))
			if err != nil {
				return nil, fmt.Errorf("failed to decode code: %v", err)
			}
		}

		acc := substate.NewAccount(nonce, balance, code)

		storageMap, ok := v.(map[string]interface{})["storage"]
		if ok {
			for key, value := range storageMap.(map[string]interface{}) {
				decodedKey, err := hex.DecodeString(strings.TrimPrefix(key, "0x"))
				if err != nil {
					return nil, fmt.Errorf("failed to decode storage key: %v", err)
				}
				decodedValue, err := hex.DecodeString(strings.TrimPrefix(value.(string), "0x"))
				if err != nil {
					return nil, fmt.Errorf("failed to decode storage value: %v", err)
				}
				acc.Storage[substatetypes.BytesToHash(decodedKey)] = substatetypes.BytesToHash(decodedValue)
			}
		}

		ssAccounts[address] = acc
	}

	return ssAccounts, err
}
