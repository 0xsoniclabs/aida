package state

import (
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/tracing"
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
	require.True(t, evmCreateCheckIsEmpty(state, address), "Account should be empty before creation")

	evmCreate(state, address)
	state.SetState(address, key, value)
	state.EndTransaction()
	state.EndBlock()

	state.BeginBlock(2)
	state.BeginTransaction(1)
	require.Equal(t, uint64(1), state.GetNonce(address))
	require.False(t, evmCreateCheckIsEmpty(state, address), "Account should not be empty after creation")
	state.EndTransaction()
	state.EndBlock()

	state.BeginBlock(3)

	// Destroy the account.
	state.BeginTransaction(1)
	state.SelfDestruct(address)
	state.EndTransaction()

	state.BeginTransaction(2)
	require.True(t, evmCreateCheckIsEmpty(state, address), "Account should be empty after self-destruct")
	state.EndTransaction()

	// Create the same account again.
	state.BeginTransaction(3)
	evmCreate(state, address)
	state.EndTransaction()

	state.EndBlock()

	state.BeginBlock(4)
	state.BeginTransaction(1)
	require.False(t, evmCreateCheckIsEmpty(state, address), "Account should not be empty after re-creation")
	state.EndTransaction()
	state.EndBlock()
}

func evmCreate(state StateDB, address common.Address) {
	code := []byte{10, 11, 12}

	if !evmCreateCheckIsEmpty(state, address) {
		return
	}

	if !state.Exist(address) {
		state.CreateAccount(address)
	}
	state.CreateContract(address)

	// By EIP-161, the nonce of a newly created account should be set to 1.
	// See: https://eips.ethereum.org/EIPS/eip-161
	state.SetNonce(address, 1, tracing.NonceChangeContractCreator)

	state.SetCode(address, code)
}

func evmCreateCheckIsEmpty(state StateDB, address common.Address) bool {
	// Check copied from evm.go create function.
	contractHash := state.GetCodeHash(address)
	storageRoot := state.GetStorageRoot(address)
	notEmpty := state.GetNonce(address) != 0 ||
		(contractHash != (common.Hash{}) && contractHash != types.EmptyCodeHash) || // non-empty code
		(storageRoot != (common.Hash{}) && storageRoot != types.EmptyRootHash)

	return !notEmpty
}
