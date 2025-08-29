// Copyright 2025 Sonic Labs
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

package state

import (
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/tracing"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/stretchr/testify/require"
)

func TestState_StorageIsEmpty_SelfDestructAndReincarnation_EmptyStateIsReportedCorrectly(t *testing.T) {
	// The issues covered by this test was discovered on the Sepolia testnet
	// where a self-destructed account was re-created in the same block.
	// See https://github.com/0xsoniclabs/sonic-admin/issues/180 for more details.
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
			defer state.Close()
			runSelfDestructAndReincarnationTest(t, state)
		})
	}
}

func runSelfDestructAndReincarnationTest(t *testing.T, state StateDB) {
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

			// Set the nonce to 1, to make sure the account is not empty and is
			// not automatically removed at the end of the block.
			state.SetNonce(address, 1, tracing.NonceChangeUnspecified)
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
			require.False(t, isEmptyStorageRoot(state.GetStorageRoot(address)))
		}
		state.EndTransaction()
	}
	state.EndBlock()
}

func TestState_StorageIsEmpty_SelfDestruct6780_EmptyStateIsReportedCorrectly(t *testing.T) {
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
			defer state.Close()
			runSelfDestruct6780AndReincarnationTest(t, state)
		})
	}
}

func runSelfDestruct6780AndReincarnationTest(t *testing.T, state StateDB) {
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
			state.SelfDestruct6780(address)

			// The self-destruct6780 operation should have no effect on the
			// storage root, as the account was not created in this transaction.
			require.False(t, isEmptyStorageRoot(state.GetStorageRoot(address)))

			// Also, the nonce should not be affected by the self-destruct6780.
			require.Equal(t, uint64(1), state.GetNonce(address))

		}
		state.EndTransaction()

		// 2. The account still exists in the following transaction.
		state.BeginTransaction(2)
		{
			// The self-destruct6780 in the previous transaction did not change
			// the storage root, so it should still be non-empty.
			require.False(t, isEmptyStorageRoot(state.GetStorageRoot(address)))

			// Also the nonce should still be 1, since the account was not
			// deleted by the self-destruct6780.
			require.Equal(t, uint64(1), state.GetNonce(address))
		}
		state.EndTransaction()
	}
	state.EndBlock()

	// Check the state after the block.
	state.BeginBlock(3)
	{
		state.BeginTransaction(1)
		{
			// The end-of-block has no effect on the account.
			require.False(t, isEmptyStorageRoot(state.GetStorageRoot(address)))
			require.Equal(t, uint64(1), state.GetNonce(address))
		}
		state.EndTransaction()
	}
	state.EndBlock()
}

func TestState_StorageIsEmpty_SelfDestruct6780InSameTransaction_EmptyStateIsReportedCorrectly(t *testing.T) {
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
			defer state.Close()
			runSelfDestruct6780InSameTransactionTest(t, state)
		})
	}
}

func runSelfDestruct6780InSameTransactionTest(t *testing.T, state StateDB) {
	address := common.Address{1, 2, 3}
	key := common.Hash{4, 5, 6}
	value := common.Hash{7, 8, 9}

	// We start by creating an account with some non-empty state.
	state.BeginBlock(1)
	{
		state.BeginTransaction(1)
		{
			// Initially, the storage root is empty.
			require.True(t, isEmptyStorageRoot(state.GetStorageRoot(address)))
			require.Equal(t, uint64(0), state.GetNonce(address))

			state.CreateAccount(address)
			state.CreateContract(address)
			state.SetNonce(address, 1, tracing.NonceChangeUnspecified)
			state.SetState(address, key, value)

			// Since the storage root is not updated while running the
			// transaction, it should still be empty.
			require.True(t, isEmptyStorageRoot(state.GetStorageRoot(address)))

			// The self-destruct6780 operation should mark the account for
			// deletion at the end of the transaction.
			state.SelfDestruct6780(address)

			// But until the ned of the transaction, the account should still
			// be considered to be present.
			require.Equal(t, uint64(1), state.GetNonce(address))
			require.True(t, isEmptyStorageRoot(state.GetStorageRoot(address)))
		}
		state.EndTransaction()

		state.BeginTransaction(2)
		{
			// It should look like the account never existed.
			require.True(t, isEmptyStorageRoot(state.GetStorageRoot(address)))
			require.Equal(t, uint64(0), state.GetNonce(address))
		}
		state.EndTransaction()
	}
	state.EndBlock()
}

func isEmptyStorageRoot(hash common.Hash) bool {
	// The empty state root is defined as the hash of an empty Merkle trie.
	return hash == common.Hash{} || hash == types.EmptyRootHash
}
