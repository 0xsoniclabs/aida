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
	"bytes"
	"encoding/json"
	"fmt"
	"math/big"
	"reflect"

	"github.com/0xsoniclabs/aida/logger"
	"github.com/0xsoniclabs/substate/db"
	"github.com/0xsoniclabs/substate/substate"
	"github.com/0xsoniclabs/substate/types"
)

var emptyStorage = types.Hash{0x0}

// SubstateDumpTxTraceDiffFunc returns a function that converts single tx from substate to transactionTrace diff mode format
func SubstateDumpTxTraceDiffFunc(getTransactionTraceFromRpc func(block uint64, tx int) (map[string]interface{}, error), log logger.Logger) func(block uint64, tx int, recording *substate.Substate, taskPool *db.SubstateTaskPool) error {
	return func(block uint64, tx int, recording *substate.Substate, taskPool *db.SubstateTaskPool) error {
		wanted, err := getTransactionTraceFromRpc(block, tx)
		if err != nil {
			return fmt.Errorf("unable to get the diff from rpc: %v", err)
		}

		inputAlloc := recording.InputSubstate
		outputAlloc := recording.OutputSubstate

		// postAllocPrinting should contain all keys from outputAlloc which are not same as in inputAlloc
		postAllocPrinting := substate.NewWorldState()
		// preAllocPrinting should contain all unchanged keys from inputAlloc that were not 0x0
		preAllocPrinting := substate.NewWorldState()

		for key, accPost := range outputAlloc {
			if accPre, ok := inputAlloc[key]; !ok {
				// if account was newly created balance is skipped
				accPost2 := accPost.Copy()
				if accPost2.Balance.Cmp(big.NewInt(0)) == 0 {
					accPost2.Balance = nil
				}
				postAllocPrinting[key] = accPost2
				continue
			} else {
				if accPre.Equal(accPost) {
					continue
				}

				accPreDiff := substate.Account{}
				accPostDiff := substate.Account{}

				if accPost.Balance.Cmp(accPre.Balance) == 0 {
					accPreDiff.Balance = accPre.Balance
					// balance is set to nil to avoid printing
					accPostDiff.Balance = nil
				} else {
					accPreDiff.Balance = accPre.Balance
					accPostDiff.Balance = accPost.Balance
				}

				if accPost.Nonce == accPre.Nonce {
					accPreDiff.Nonce = accPre.Nonce
					// nonce is set to avoid printing
					accPostDiff.Nonce = 0
				} else {
					accPreDiff.Nonce = accPre.Nonce
					accPostDiff.Nonce = accPost.Nonce
				}

				if bytes.Compare(accPre.Code, accPost.Code) == 0 {
					accPreDiff.Code = accPre.Code
					accPostDiff.Code = nil
				} else {
					accPreDiff.Code = accPre.Code
					accPostDiff.Code = accPost.Code
				}

				accPreDiff.Storage = make(map[types.Hash]types.Hash)
				accPostDiff.Storage = make(map[types.Hash]types.Hash)
				for k, sPost := range accPost.Storage {
					if sPre, ok2 := accPre.Storage[k]; !ok2 {
						// new storage is listed just in post
						accPostDiff.Storage[k] = sPost
					} else {
						if sPost.Compare(sPre) != 0 {
							accPostDiff.Storage[k] = sPost
							if sPre.Compare(emptyStorage) != 0 {
								accPreDiff.Storage[k] = sPre
							}
						}
					}
				}

				preAllocPrinting[key] = &accPreDiff
				postAllocPrinting[key] = &accPostDiff
			}
		}

		fmt.Printf("block: %v Transaction: %v\n", block, tx)
		postAllocMap := formatWorldState(postAllocPrinting)
		preAllocMap := formatWorldState(preAllocPrinting)

		// holding debug_traceTransaction diff mode format
		result := map[string]interface{}{
			"post": postAllocMap,
			"pre":  preAllocMap,
		}

		jbytes, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			return err
		}

		// jbytes back to map[string]interface{}
		resultOrdered := make(map[string]interface{})
		err = json.Unmarshal(jbytes, &resultOrdered)

		if !reflect.DeepEqual(resultOrdered, wanted) {
			return fmt.Errorf("produced substate generated diff does not match expected diff from rpc")
		}

		fmt.Println(string(jbytes))

		return nil
	}
}

// formatWorldState formats world state for tx trace diff mode printing
func formatWorldState(alloc substate.WorldState) map[string]interface{} {
	result := make(map[string]interface{})
	for addr, acc := range alloc {
		result[addr.String()] = formatAccount(acc)
	}
	return result
}

// formatAccount formats account for tx trace diff mode printing
func formatAccount(a *substate.Account) map[string]interface{} {
	result := make(map[string]interface{})

	// balance is set to nil to avoid printing
	if a.Balance != nil {
		result["balance"] = fmt.Sprintf("0x%v", a.Balance.Text(16))
	}
	// nonce is set to avoid printing
	if a.Nonce != 0 {
		result["nonce"] = a.Nonce
	}

	if a.Code != nil && len(a.Code) > 0 {
		result["code"] = fmt.Sprintf("0x%x", a.Code)
	}

	if a.Storage != nil && len(a.Storage) > 0 {
		resultStorage := make(map[string]interface{})
		for key, val := range a.Storage {
			if val.Compare(emptyStorage) != 0 {
				resultStorage[key.String()] = val.String()
			}
		}

		// check if all values were not empty
		if len(resultStorage) > 0 {
			result["storage"] = resultStorage
		}
	}
	return result
}
