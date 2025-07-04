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

	// We start by creating an account with some non-empty state.
	state.BeginBlock(1)
	{
		state.BeginTransaction(1)
		{
			state.CreateAccount(address)
			state.CreateContract(address)
			state.SetNonce(address, 1, tracing.NonceChangeUnspecified)
			state.SetState(address, key, value)
		}
		state.EndTransaction()
	}
	state.EndBlock()

	// In the next block we have two transactions:
	state.BeginBlock(2)
	{
		// 1. Self-destruct the account.
		state.BeginTransaction(1)
		{
			// At this point, the account storage is not empty.
			require.False(t, isEmptyStorageRoot(state.GetStorageRoot(address)))
			require.Equal(t, uint64(1), state.GetNonce(address))
			state.SelfDestruct(address)

			// The effects of the self-destruct are only visible after the end
			// of the transaction. So the storage root is still not empty.
			require.False(t, isEmptyStorageRoot(state.GetStorageRoot(address)))
			require.Equal(t, uint64(1), state.GetNonce(address))

		}
		state.EndTransaction()

		// 2. Create the same account again.
		state.BeginTransaction(2)
		{
			// The effects of the self-destruct are now visible. The storage
			// root should be empty and the nonce should be reset to 0.
			require.True(t, isEmptyStorageRoot(state.GetStorageRoot(address)))
			require.Equal(t, uint64(0), state.GetNonce(address))

			state.CreateAccount(address)
			state.CreateContract(address)

			// The account is still empty, even after its re-creation.
			require.True(t, isEmptyStorageRoot(state.GetStorageRoot(address)))
			require.Equal(t, uint64(0), state.GetNonce(address))

			// Writing to the storage does not change the storage root, since it
			// is only re-evaluated at the end of the block.
			state.SetState(address, key, value)
			require.True(t, isEmptyStorageRoot(state.GetStorageRoot(address)))
		}
		state.EndTransaction()

		// 3. The view on the re-born account after the end of the transaction.
		state.BeginTransaction(3)
		{
			// Even here, the storage root remains empty, as the storage root is
			// not re-evaluated until the end of the block.
			require.True(t, isEmptyStorageRoot(state.GetStorageRoot(address)))
		}
		state.EndTransaction()

	}
	state.EndBlock()

	// Check the state after the block.
	state.BeginBlock(3)
	{
		state.BeginTransaction(1)
		{
			require.True(t, isEmptyStorageRoot(state.GetStorageRoot(address)))
		}
		state.EndTransaction()
	}
	state.EndBlock()
}

func isEmptyStorageRoot(hash common.Hash) bool {
	// The empty state root is defined as the hash of an empty Merkle trie.
	return hash == common.Hash{} || hash == types.EmptyRootHash
}
