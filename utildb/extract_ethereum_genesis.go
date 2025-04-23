package utildb

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"

	"github.com/0xsoniclabs/substate/substate"
	substatetypes "github.com/0xsoniclabs/substate/types"
	"github.com/holiman/uint256"
)

// LoadEthereumGenesisWorldState loads opera initial world state from worldstate-db as WorldState
func LoadEthereumGenesisWorldState(genesis string) (substate.WorldState, error) {
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
