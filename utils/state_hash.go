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

package utils

//go:generate mockgen -source state_hash.go -destination state_hash_mock.go -package utils

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/0xsoniclabs/aida/logger"
	"github.com/0xsoniclabs/substate/db"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/status-im/keycard-go/hexutils"
)

const StateHashPrefix = "dbh"

// ClientInterface defines the methods that an RPC client must implement.
type IRpcClient interface {
	RegisterName(name string, receiver interface{}) error
	SupportedModules() (map[string]string, error)
	Close()
	SetHeader(key, value string)
	Call(result interface{}, method string, args ...interface{}) error
	CallContext(ctx context.Context, result interface{}, method string, args ...interface{}) error
	BatchCall(b []rpc.BatchElem) error
	BatchCallContext(ctx context.Context, b []rpc.BatchElem) error
	Notify(ctx context.Context, method string, args ...interface{}) error
	EthSubscribe(ctx context.Context, channel interface{}, args ...interface{}) (*rpc.ClientSubscription, error)
	ShhSubscribe(ctx context.Context, channel interface{}, args ...interface{}) (*rpc.ClientSubscription, error)
	Subscribe(ctx context.Context, namespace string, channel interface{}, args ...interface{}) (*rpc.ClientSubscription, error)
	SupportsSubscriptions() bool
}

type StateHashProvider interface {
	GetStateHash(blockNumber int) (common.Hash, error)
}

func MakeStateHashProvider(db db.BaseDB) StateHashProvider {
	return &stateHashProvider{db}
}

type stateHashProvider struct {
	db db.BaseDB
}

func (p *stateHashProvider) GetStateHash(number int) (common.Hash, error) {
	hex := strconv.FormatUint(uint64(number), 16)
	stateRoot, err := p.db.Get([]byte(StateHashPrefix + "0x" + hex))
	if err != nil {
		return common.Hash{}, err
	}

	if stateRoot == nil {
		return common.Hash{}, nil
	}

	return common.Hash(stateRoot), nil
}

// StateHashScraper scrapes state hashes from a node and saves them to a leveldb database
func StateHashScraper(ctx context.Context, chainId ChainID, operaPath string, db db.BaseDB, firstBlock, lastBlock uint64, log logger.Logger) error {
	ipcPath := operaPath + "/sonic.ipc"

	client, err := getClient(ctx, chainId, ipcPath, log)
	if err != nil {
		return err
	}
	defer client.Close()

	var i = firstBlock

	// If firstBlock is 0, we need to get the state root for block 1 and save it as the state root for block 0
	// this is because the correct state root for block 0 is not available from the rpc node (at least in fantom mainnet and testnet)
	if firstBlock == 0 {
		block, err := retrieveStateRoot(client, "0x1")
		if err != nil {
			return err
		}

		if block == nil {
			return fmt.Errorf("block 1 not found")
		}

		err = SaveStateRoot(db, "0x0", block["stateRoot"].(string))
		if err != nil {
			return err
		}
		i++
	}

	for ; i <= lastBlock; i++ {
		blockNumber := fmt.Sprintf("0x%x", i)
		block, err := retrieveStateRoot(client, blockNumber)
		if err != nil {
			return err
		}

		if block == nil {
			return fmt.Errorf("block %d not found", i)
		}

		err = SaveStateRoot(db, blockNumber, block["stateRoot"].(string))
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
func getClient(ctx context.Context, chainId ChainID, ipcPath string, log logger.Logger) (*rpc.Client, error) {
	var client *rpc.Client
	var err error

	_, errIpc := os.Stat(ipcPath)
	if errIpc == nil {
		// ipc file exists
		client, err = rpc.DialIPC(ctx, ipcPath)
		if err != nil {
			return nil, fmt.Errorf("failed to connect to IPC at %s: %v", ipcPath, err)
		}
		log.Infof("Connected to IPC at %s", ipcPath)
		return client, err
	} else {
		var provider string
		provider, err = GetProvider(chainId)
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
}

// SaveStateRoot saves the state root hash to the database
func SaveStateRoot(db db.BaseDB, blockNumber string, stateRoot string) error {
	fullPrefix := StateHashPrefix + blockNumber
	err := db.Put([]byte(fullPrefix), hexutils.HexToBytes(strings.TrimPrefix(stateRoot, "0x")))
	if err != nil {
		return fmt.Errorf("unable to put state hash for block %s: %v", blockNumber, err)
	}
	return nil
}

// retrieveStateRoot gets the state root hash from the rpc node
func retrieveStateRoot(client IRpcClient, blockNumber string) (map[string]interface{}, error) {
	var block map[string]interface{}
	err := client.Call(&block, "eth_getBlockByNumber", blockNumber, false)
	if err != nil {
		return nil, fmt.Errorf("failed to get block %s: %v", blockNumber, err)
	}
	return block, nil
}

// StateHashKeyToUint64 converts a state hash key to a uint64
func StateHashKeyToUint64(hexBytes []byte) (uint64, error) {
	prefix := []byte(StateHashPrefix)

	if len(hexBytes) >= len(prefix) && bytes.HasPrefix(hexBytes, prefix) {
		hexBytes = hexBytes[len(prefix):]
	}

	res, err := strconv.ParseUint(string(hexBytes), 0, 64)

	if err != nil {
		return 0, fmt.Errorf("cannot parse uint %v; %v", string(hexBytes), err)
	}
	return res, nil
}

// GetFirstStateHash returns the first block number for which we have a state hash
func GetFirstStateHash(db db.BaseDB) (uint64, error) {
	// TODO MATEJ will be fixed in future commit
	//iter := db.NewIterator([]byte(StateHashPrefix), []byte("0x"))
	//
	//defer iter.Release()
	//
	//// start with writing first block
	//if !iter.Next() {
	//	return 0, fmt.Errorf("no state hash found")
	//}
	//
	//firstStateHashBlock, err := StateHashKeyToUint64(iter.Key())
	//if err != nil {
	//	return 0, err
	//}
	//return firstStateHashBlock, nil
	return 0, fmt.Errorf("not implemented")
}

// GetLastStateHash returns the last block number for which we have a state hash
func GetLastStateHash(db db.BaseDB) (uint64, error) {
	// TODO MATEJ will be fixed in future commit
	//return GetLastKey(db, StateHashPrefix)
	return 0, fmt.Errorf("not implemented")
}
