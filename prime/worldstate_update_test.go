package prime

import (
	"fmt"
	"testing"

	"github.com/0xsoniclabs/aida/logger"
	"github.com/0xsoniclabs/aida/state"
	"github.com/0xsoniclabs/aida/utils"
	substateDb "github.com/0xsoniclabs/substate/db"
	"github.com/0xsoniclabs/substate/rlp"
	"github.com/0xsoniclabs/substate/substate"
	substatetypes "github.com/0xsoniclabs/substate/types"
	trlp "github.com/0xsoniclabs/substate/types/rlp"
	"github.com/0xsoniclabs/substate/updateset"
	"github.com/ethereum/go-ethereum/common"
	"github.com/holiman/uint256"
	"github.com/stretchr/testify/assert"
	"github.com/syndtr/goleveldb/leveldb/iterator"
	"github.com/syndtr/goleveldb/leveldb/testutil"
	"go.uber.org/mock/gomock"
)

func TestWorldStateUpdate_GenerateUpdateSet(t *testing.T) {
	// TODO Protobuf encoding not supported yet
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	baseDb := substateDb.NewMockBaseDB(ctrl)
	mockDb := substateDb.NewMockDbAdapter(ctrl)

	input := utils.GetTestSubstate("default")
	input.Block = 0
	input.Transaction = 1
	encoded, err := trlp.EncodeToBytes(rlp.NewRLP(input))
	if err != nil {
		t.Fatalf("Failed to encode substate: %v", err)
	}

	expectedDestroyed := []substatetypes.Address{{1}, {2}}
	expectedResurrected := []substatetypes.Address{{3}}
	list := substateDb.SuicidedAccountLists{DestroyedAccounts: expectedDestroyed, ResurrectedAccounts: expectedResurrected}
	value, _ := trlp.EncodeToBytes(list)

	kv := &testutil.KeyValue{}
	iter1 := iterator.NewArrayIterator(kv)
	iter2 := iterator.NewArrayIterator(kv)
	iter3 := iterator.NewArrayIterator(kv)
	iter4 := iterator.NewArrayIterator(kv)
	kv.PutU(substateDb.SubstateDBKey(input.Block, input.Transaction), encoded)
	mockDb.EXPECT().Get(gomock.Any(), gomock.Any()).Return(encoded, nil).AnyTimes()
	// Encoding is being checked against iterator - that's why we need multiple iterators
	mockDb.EXPECT().NewIterator(gomock.Any(), gomock.Any()).Return(iter1)
	mockDb.EXPECT().NewIterator(gomock.Any(), gomock.Any()).Return(iter2)
	mockDb.EXPECT().NewIterator(gomock.Any(), gomock.Any()).Return(iter3)
	mockDb.EXPECT().NewIterator(gomock.Any(), gomock.Any()).Return(iter4)
	baseDb.EXPECT().GetBackend().Return(mockDb)
	baseDb.EXPECT().Get(gomock.Any()).Return(value, nil).AnyTimes()

	set, i, err := generateUpdateSet(0, 2, &utils.Config{
		Workers: 1,
	}, baseDb)
	assert.NoError(t, iter1.Error())
	assert.NoError(t, iter2.Error())
	assert.NoError(t, iter3.Error())
	assert.NoError(t, iter4.Error())
	assert.NoError(t, err)
	assert.NotNil(t, set)
	assert.Equal(t, 1, len(set))
	assert.Equal(t, 3, len(i))
}

func TestWorldStateUpdate_ClearAccountStorage(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	addr := substatetypes.BytesToAddress([]byte("test"))
	ws := substate.WorldState{}
	ws[addr] = &substate.Account{
		Nonce:   1,
		Balance: nil,
		Storage: map[substatetypes.Hash]substatetypes.Hash{
			substatetypes.BytesToHash([]byte("key1")): substatetypes.BytesToHash([]byte("value1")),
		},
		Code: nil,
	}
	ClearAccountStorage(ws, []substatetypes.Address{addr})
	assert.Equal(t, 0, len(ws[addr].Storage))
}

var testUpdateSet = &updateset.UpdateSet{
	WorldState: substate.WorldState{
		substatetypes.Address{1}: &substate.Account{
			Nonce:   1,
			Balance: new(uint256.Int).SetUint64(1),
		},
		substatetypes.Address{2}: &substate.Account{
			Nonce:   2,
			Balance: new(uint256.Int).SetUint64(2),
		},
	},
	Block: 1,
}

var testDeletedAccounts = []substatetypes.Address{{3}, {4}}

func createTestUpdateDB(dbPath string) (substateDb.UpdateDB, error) {
	db, err := substateDb.NewUpdateDB(dbPath, nil, nil, nil)
	if err != nil {
		return nil, err
	}
	return db, nil
}

// TestStatedb_DeleteDestroyedAccountsFromWorldState tests removal of destroyed accounts from given world state
func TestStatedb_DeleteDestroyedAccountsFromWorldState(t *testing.T) {
	for _, tc := range utils.GetStateDbTestCases() {
		t.Run(fmt.Sprintf("DB variant: %s; shadowImpl: %s; archive variant: %s", tc.Variant, tc.ShadowImpl, tc.ArchiveVariant), func(t *testing.T) {
			cfg := utils.MakeTestConfig(tc)
			// Generating randomized world state
			ws, addrList := utils.MakeWorldState(t)
			// Init directory for destroyed accounts DB
			deletionDb := t.TempDir()
			// Pick two account which will represent destroyed ones
			destroyedAccounts := []substatetypes.Address{
				substatetypes.Address(addrList[0]),
				substatetypes.Address(addrList[50]),
			}

			// Update config to enable removal of destroyed accounts
			cfg.DeletionDb = deletionDb

			// Initializing backend DB for storing destroyed accounts
			daBackend, err := substateDb.NewDefaultBaseDB(deletionDb)
			if err != nil {
				t.Fatalf("failed to create backend DB: %s; %v", deletionDb, err)
			}

			// Creating new destroyed accounts DB
			daDB := substateDb.MakeDefaultDestroyedAccountDBFromBaseDB(daBackend)

			// Storing two picked accounts from destroyedAccounts slice to destroyed accounts DB
			err = daDB.SetDestroyedAccounts(5, 1, destroyedAccounts, []substatetypes.Address{})
			if err != nil {
				t.Fatalf("failed to set destroyed accounts into DB: %v", err)
			}

			// Closing destroyed accounts DB
			err = daDB.Close()
			if err != nil {
				t.Fatalf("failed to close destroyed accounts DB: %v", err)
			}

			// Call for removal of destroyed accounts from given world state
			err = deleteDestroyedAccountsFromWorldState(ws, cfg, 5)
			if err != nil {
				t.Fatalf("failed to delete accounts from the world state: %v", err)
			}

			// check if accounts are not present anymore
			if ws.Get(common.Address(destroyedAccounts[0])) != nil || ws.Get(common.Address(destroyedAccounts[1])) != nil {
				t.Fatalf("failed to delete accounts from the world state")
			}
		})
	}
}

// TestStatedb_DeleteDestroyedAccountsFromWorldState tests removal of deleted accounts from given state DB
func TestStatedb_DeleteDestroyedAccountsFromStateDB(t *testing.T) {
	for _, tc := range utils.GetStateDbTestCases() {
		t.Run(fmt.Sprintf("DB variant: %s; shadowImpl: %s; archive variant: %s", tc.Variant, tc.ShadowImpl, tc.ArchiveVariant), func(t *testing.T) {
			cfg := utils.MakeTestConfig(tc)
			// Generating randomized world state
			ws, addrList := utils.MakeWorldState(t)
			// Init directory for destroyed accounts DB
			deletedAccountsDir := t.TempDir()
			// Pick two account which will represent destroyed ones
			destroyedAccounts := []substatetypes.Address{
				substatetypes.Address(addrList[0]),
				substatetypes.Address(addrList[50]),
			}

			// Update config to enable removal of destroyed accounts
			cfg.DeletionDb = deletedAccountsDir

			// Initializing backend DB for storing destroyed accounts
			base, err := substateDb.NewDefaultBaseDB(deletedAccountsDir)
			if err != nil {
				t.Fatalf("failed to create backend DB: %s; %v", deletedAccountsDir, err)
			}

			// Creating new destroyed accounts DB
			daDB := substateDb.MakeDefaultDestroyedAccountDBFromBaseDB(base)

			// Storing two picked accounts from destroyedAccounts slice to destroyed accounts DB
			err = daDB.SetDestroyedAccounts(5, 1, destroyedAccounts, []substatetypes.Address{})
			if err != nil {
				t.Fatalf("failed to set destroyed accounts into DB: %v", err)
			}

			defer func(daDB *substateDb.DestroyedAccountDB) {
				e := daDB.Close()
				if e != nil {
					t.Fatalf("failed to close destroyed accounts DB: %v", e)
				}
			}(daDB)

			// Initialization of state DB
			sDB, _, err := utils.PrepareStateDB(cfg)
			if err != nil {
				t.Fatalf("failed to create state DB: %v", err)
			}

			log := logger.NewLogger("INFO", "TestStateDb")

			p := MakePrimer(cfg, sDB, log)
			// Priming state DB with given world state
			err = p.ctx.PrimeStateDB(ws, sDB)
			if err != nil {
				t.Fatalf("cannot prime statedb; %v", err)
			}

			// Call for removal of destroyed accounts from state DB
			err = p.mayDeleteDestroyedAccountsFromStateDB(5, base)
			if err != nil {
				t.Fatalf("failed to delete accounts from the state DB: %v", err)
			}

			err = state.BeginCarmenDbTestContext(sDB)
			if err != nil {
				t.Fatal(err)
			}

			// check if accounts are not present anymore
			for _, da := range destroyedAccounts {
				if sDB.Exist(common.Address(da)) {
					t.Fatalf("failed to delete destroyed accounts from the state DB")
				}
			}

			err = state.CloseCarmenDbTestContext(sDB)
			if err != nil {
				t.Fatal(err)
			}
		})
	}
}
