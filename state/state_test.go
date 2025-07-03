package state

import (
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/stretchr/testify/require"
)

func TestState_SepoliaIssue(t *testing.T) {
	impl := map[string]func(t *testing.T) StateDB{
		"geth": func(t *testing.T) StateDB {
			db, err := MakeGethStateDB(t.TempDir(), "", common.Hash{}, false, nil)
			require.NoError(t, err, "Failed to create Geth StateDB")
			return db
		},
		"carmen": func(t *testing.T) StateDB {
			db, err := MakeCarmenStateDB(t.TempDir(), "go-file", 5, "", 0, 0, 0, 0)
			require.NoError(t, err, "Failed to create Carmen StateDB")
			return db
		},
	}

	for name, factory := range impl {
		t.Run(name, func(t *testing.T) {
			state := factory(t)
			runSepoliaIssue(t, state)
		})
	}
}

func runSepoliaIssue(t *testing.T, state StateDB) {
	address := common.Address{1, 2, 3}
	key := common.Hash{4, 5, 6}
	value := common.Hash{7, 8, 9}

	state.BeginBlock(1)
	// Create an non-empty account.
	state.BeginTransaction(1)
	state.CreateAccount(address)
	state.CreateContract(address)
	state.SetState(address, key, value)
	state.EndTransaction()
	state.EndBlock()

	require.False(t, evmCreateCheckIsEmpty(state, address), "Account should not be empty after creation")

	state.BeginBlock(2)
	// Destroy the account.
	state.BeginTransaction(1)
	state.SelfDestruct(address)
	state.EndTransaction()

	// Create the same account again.
	state.BeginTransaction(2)
	state.CreateAccount(address)
	state.CreateContract(address)
	state.SetState(address, key, value)
	state.EndTransaction()
	state.EndBlock()

	require.False(t, evmCreateCheckIsEmpty(state, address), "Account should not be empty after re-creation")
}

func evmCreateCheckIsEmpty(state StateDB, address common.Address) bool {
	contractHash := state.GetCodeHash(address)
	storageRoot := state.GetStorageRoot(address)
	nonce := state.GetNonce(address)

	// Check copied from evm.go create function.
	return nonce != 0 ||
		(contractHash != (common.Hash{}) && contractHash != types.EmptyCodeHash) || // non-empty code
		(storageRoot != (common.Hash{}) && storageRoot != types.EmptyRootHash)
}
