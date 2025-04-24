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

type GenesisData struct {
	Alloc map[string]struct {
		Balance string            `json:"balance,omitempty"`
		Nonce   string            `json:"nonce,omitempty"`
		Code    string            `json:"code,omitempty"`
		Storage map[string]string `json:"storage,omitempty"`
	} `json:"alloc"`
}

// LoadEthereumGenesisWorldState loads opera initial world state from worldstate-db as WorldState
func LoadEthereumGenesisWorldState(genesis string) (substate.WorldState, error) {
	var jsData GenesisData
	// Read the content of the JSON file
	jsonData, err := ioutil.ReadFile(genesis)
	if err != nil {
		return nil, fmt.Errorf("failed to read genesis file: %account", err)
	}

	// Unmarshal JSON data
	if err := json.Unmarshal(jsonData, &jsData); err != nil {
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

		if account.Storage != nil && len(account.Storage) > 0 {
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
