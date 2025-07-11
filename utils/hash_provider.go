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

//go:generate mockgen -source hash_provider.go -destination hash_provider_mock.go -package utils

import (
	"bytes"
	"context"
	"encoding/binary"
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

const (
	StateRootHashPrefix = "dbh"
	BlockHashPrefix     = "bh"
)

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

type HashProvider interface {
	GetStateRootHash(blockNumber int) (common.Hash, error)
	GetBlockHash(blockNumber int) (common.Hash, error)
}

func MakeHashProvider(db db.BaseDB) HashProvider {
	return &hashProvider{db}
}

type hashProvider struct {
	db db.BaseDB
}

func (p *hashProvider) GetBlockHash(number int) (common.Hash, error) {
	blockHash, err := p.db.Get(BlockHashDBKey(uint64(number)))
	if err != nil {
		return common.Hash{}, err
	}

	if blockHash == nil {
		return common.Hash{}, nil
	}

	if len(blockHash) != 32 {
		return common.Hash{}, fmt.Errorf("invalid block hash length for block %d: expected 32 bytes, got %d bytes", number, len(blockHash))
	}

	return common.Hash(blockHash), nil
}

func (p *hashProvider) GetStateRootHash(number int) (common.Hash, error) {
	hex := strconv.FormatUint(uint64(number), 16)
	stateRoot, err := p.db.Get([]byte(StateRootHashPrefix + "0x" + hex))
	if err != nil {
		return common.Hash{}, err
	}

	if stateRoot == nil {
		return common.Hash{}, nil
	}

	if len(stateRoot) != 32 {
		return common.Hash{}, fmt.Errorf("invalid state root length for block %d: expected 32 bytes, got %d bytes", number, len(stateRoot))
	}

	return common.BytesToHash(stateRoot), nil
}

// StateAndBlockHashScraper scrapes state and block hashes from a node and saves them to a leveldb database
func StateAndBlockHashScraper(ctx context.Context, chainId ChainID, clientDb string, db db.BaseDB, firstBlock, lastBlock uint64, log logger.Logger) error {
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

		err = SaveStateRoot(db, "0x0", block["stateRoot"].(string))
		if err != nil {
			return err
		}
		err = SaveBlockHash(db, "0x1", block["hash"].(string))
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

		err = SaveStateRoot(db, blockNumber, block["stateRoot"].(string))
		if err != nil {
			return err
		}
		err = SaveBlockHash(db, blockNumber, block["hash"].(string))
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
func getClient(ctx context.Context, chainId ChainID, clientDb string, log logger.Logger) (*rpc.Client, error) {
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

// SaveStateRoot saves the state root hash to the database
func SaveStateRoot(db db.BaseDB, blockNumber string, stateRoot string) error {
	fullPrefix := StateRootHashPrefix + blockNumber
	err := db.Put([]byte(fullPrefix), hexutils.HexToBytes(strings.TrimPrefix(stateRoot, "0x")))
	if err != nil {
		return fmt.Errorf("unable to put state hash for block %s: %v", blockNumber, err)
	}
	return nil
}

// SaveBlockHash saves the block hash to the database
func SaveBlockHash(db db.BaseDB, blockNumber string, hash string) error {
	bn, err := strconv.ParseUint(strings.TrimPrefix(blockNumber, "0x"), 16, 64)
	if err != nil {
		return fmt.Errorf("invalid block number %s: %v", blockNumber, err)
	}
	fullPrefix := BlockHashDBKey(bn)
	err = db.Put(fullPrefix, hexutils.HexToBytes(strings.TrimPrefix(hash, "0x")))
	if err != nil {
		return fmt.Errorf("unable to put state hash for block %s: %v", blockNumber, err)
	}
	return nil
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

// StateHashKeyToUint64 converts a state hash key to a uint64
func StateHashKeyToUint64(hexBytes []byte) (uint64, error) {
	prefix := []byte(StateRootHashPrefix)

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
	//iter := db.NewIterator([]byte(StateRootHashPrefix), []byte("0x"))
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
	//return GetLastKey(db, StateRootHashPrefix)
	return 0, fmt.Errorf("not implemented")
}

// GetFirstBlockHash returns the first block number for which we have a block hash
func GetFirstBlockHash(db db.BaseDB) (uint64, error) {
	iter := db.NewIterator([]byte(BlockHashPrefix), nil)
	defer iter.Release()

	if !iter.Next() {
		return 0, fmt.Errorf("no block hash found")
	}

	firstBlock, err := DecodeBlockHashDBKey(iter.Key())
	if err != nil {
		return 0, err
	}
	return firstBlock, nil
}

// GetLastBlockHash returns the last block number for which we have a block hash
func GetLastBlockHash(db db.BaseDB) (uint64, error) {
	iter := db.NewIterator([]byte(BlockHashPrefix), nil)
	defer iter.Release()

	if !iter.Last() {
		return 0, fmt.Errorf("no block hash found")
	}

	lastBlock, err := DecodeBlockHashDBKey(iter.Key())
	if err != nil {
		return 0, err
	}
	return lastBlock, nil
}

func BlockHashDBKey(block uint64) []byte {
	prefix := []byte(BlockHashPrefix)
	blockByte := make([]byte, 8)
	binary.BigEndian.PutUint64(blockByte[0:8], block)
	return append(prefix, blockByte...)
}

// DecodeBlockHashDBKey decodes a block hash key into a block number
func DecodeBlockHashDBKey(data []byte) (uint64, error) {
	if len(data) < len(BlockHashPrefix)+8 {
		return 0, fmt.Errorf("invalid length of block hash key, expected at least %d, got %d", len(BlockHashPrefix)+8, len(data))
	}
	if !bytes.HasPrefix(data, []byte(BlockHashPrefix)) {
		return 0, fmt.Errorf("invalid prefix of block hash key")
	}
	block := binary.BigEndian.Uint64(data[len(BlockHashPrefix):])
	return block, nil
}
