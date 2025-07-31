package utils

import (
	"math/big"

	substateDb "github.com/0xsoniclabs/substate/db"
	"github.com/0xsoniclabs/substate/substate"
	"github.com/0xsoniclabs/substate/types"
	"github.com/0xsoniclabs/substate/updateset"
	"github.com/holiman/uint256"
)

var testUpdateSet = &updateset.UpdateSet{
	WorldState: substate.WorldState{
		types.Address{1}: &substate.Account{
			Nonce:   1,
			Balance: new(uint256.Int).SetUint64(1),
		},
		types.Address{2}: &substate.Account{
			Nonce:   2,
			Balance: new(uint256.Int).SetUint64(2),
		},
	},
	Block: 1,
}

var testDeletedAccounts = []types.Address{{3}, {4}}

func createTestUpdateDB(dbPath string) (substateDb.UpdateDB, error) {
	db, err := substateDb.NewUpdateDB(dbPath, nil, nil, nil)
	if err != nil {
		return nil, err
	}
	return db, nil
}

func GetTestSubstate(encoding string) *substate.Substate {
	txType := int32(substate.SetCodeTxType)
	ss := &substate.Substate{
		InputSubstate:  substate.NewWorldState().Add(types.Address{1}, 1, new(uint256.Int).SetUint64(1), nil),
		OutputSubstate: substate.NewWorldState().Add(types.Address{2}, 2, new(uint256.Int).SetUint64(2), nil),
		Env: &substate.Env{
			Coinbase:   types.Address{1},
			Difficulty: new(big.Int).SetUint64(1),
			GasLimit:   1,
			Number:     1,
			Timestamp:  1,
			BaseFee:    new(big.Int).SetUint64(1),
			Random:     &types.Hash{1},
		},
		Message: substate.NewMessage(
			1,
			true,
			new(big.Int).SetUint64(1),
			1,
			types.Address{1},
			new(types.Address), new(big.Int).SetUint64(1), []byte{1}, nil, &txType,
			types.AccessList{{Address: types.Address{1}, StorageKeys: []types.Hash{{1}, {2}}}}, new(big.Int).SetUint64(1),
			new(big.Int).SetUint64(1), new(big.Int).SetUint64(1), make([]types.Hash, 0),
			[]types.SetCodeAuthorization{
				{ChainID: *uint256.NewInt(1), Address: types.Address{1}, Nonce: 1, V: 1, R: *uint256.NewInt(1), S: *uint256.NewInt(1)},
			}),
		Result: substate.NewResult(1, types.Bloom{1}, []*types.Log{
			{
				Address: types.Address{1},
				Topics:  []types.Hash{{1}, {2}},
				Data:    []byte{1, 2, 3},
				// intentionally skipped: BlockNumber, TxIndex, Index - because protobuf Substate encoding doesn't use these values
				TxHash:    types.Hash{1},
				BlockHash: types.Hash{1},
				Removed:   false,
			},
		},
			// intentionally skipped: ContractAddress - because protobuf Substate encoding doesn't use this value,
			// instead the ContractAddress is derived from Message.From and Message.Nonce
			types.Address{},
			1),
		Block:       37_534_834,
		Transaction: 1,
	}

	// remove fields that are not supported in rlp encoding
	// TODO once protobuf becomes default add ' && encoding != "default" ' to the condition
	if encoding != "protobuf" {
		ss.Env.Random = nil
		ss.Message.AccessList = types.AccessList{}
		ss.Message.SetCodeAuthorizations = nil
	}
	return ss
}
