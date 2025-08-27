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

package executor

import (
	"context"
	"errors"
	"math/big"
	"testing"

	"github.com/0xsoniclabs/aida/state"
	"github.com/0xsoniclabs/aida/txcontext"
	"github.com/0xsoniclabs/aida/utils"
	"github.com/0xsoniclabs/norma/load/app"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/holiman/uint256"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func Expect_initializeTreasureAccount(dbMock *state.MockStateDB) {
	gomock.InOrder(
		dbMock.EXPECT().BeginBlock(gomock.Any()),
		dbMock.EXPECT().BeginTransaction(gomock.Any()),
		dbMock.EXPECT().CreateAccount(gomock.Any()),
		dbMock.EXPECT().AddBalance(gomock.Any(), gomock.Any(), gomock.Any()),
		dbMock.EXPECT().EndTransaction(),
		dbMock.EXPECT().EndBlock(),
	)
}

func Expect_initializeAppContext(dbMock *state.MockStateDB) {
	gomock.InOrder(
		dbMock.EXPECT().BeginTransaction(gomock.Any()),
		dbMock.EXPECT().GetNonce(gomock.Any()).Return(uint64(0)),
		dbMock.EXPECT().EndTransaction(),
	)
}

func Expect_initializeApp_Counter(dbMock *state.MockStateDB) {
	gomock.InOrder(
		// contract deployment
		dbMock.EXPECT().BeginTransaction(gomock.Any()),
		dbMock.EXPECT().GetNonce(gomock.Any()).Return(uint64(0)),
		dbMock.EXPECT().EndTransaction(),

		// funding accounts
		dbMock.EXPECT().BeginTransaction(gomock.Any()),
		dbMock.EXPECT().GetNonce(gomock.Any()).Return(uint64(0)),
		dbMock.EXPECT().EndTransaction(),

		// waiting for contract deployment requires checking the nonce
		// of funded accounts, since we did no funding, then nonce is 0
		dbMock.EXPECT().BeginTransaction(gomock.Any()),
		dbMock.EXPECT().GetNonce(gomock.Any()).Return(uint64(0)),
		dbMock.EXPECT().EndTransaction(),
	)
}

func Expect_initializeApp_Erc20(dbMock *state.MockStateDB) {
	gomock.InOrder(
		// contract deployment
		dbMock.EXPECT().BeginTransaction(gomock.Any()),
		dbMock.EXPECT().GetNonce(gomock.Any()).Return(uint64(0)),
		dbMock.EXPECT().EndTransaction(),

		// funding accounts
		dbMock.EXPECT().BeginTransaction(gomock.Any()),
		dbMock.EXPECT().GetNonce(gomock.Any()).Return(uint64(0)),
		dbMock.EXPECT().EndTransaction(),
		dbMock.EXPECT().BeginTransaction(gomock.Any()),
		dbMock.EXPECT().GetNonce(gomock.Any()).Return(uint64(0)),
		dbMock.EXPECT().EndTransaction(),

		// waiting for contract deployment requires checking the nonce
		// of funded accounts, since we did no funding, then nonce is 0
		dbMock.EXPECT().BeginTransaction(gomock.Any()),
		dbMock.EXPECT().GetNonce(gomock.Any()).Return(uint64(0)),
		dbMock.EXPECT().EndTransaction(),
	)
}

func Expect_initializeApp_Store(dbMock *state.MockStateDB) {
	gomock.InOrder(
		// contract deployment
		dbMock.EXPECT().BeginTransaction(gomock.Any()),
		dbMock.EXPECT().GetNonce(gomock.Any()).Return(uint64(0)),
		dbMock.EXPECT().EndTransaction(),

		// funding accounts
		dbMock.EXPECT().BeginTransaction(gomock.Any()),
		dbMock.EXPECT().GetNonce(gomock.Any()).Return(uint64(0)),
		dbMock.EXPECT().EndTransaction(),

		// waiting for contract deployment requires checking the nonce
		// of funded accounts, since we did no funding, then nonce is 0
		dbMock.EXPECT().BeginTransaction(gomock.Any()),
		dbMock.EXPECT().GetNonce(gomock.Any()).Return(uint64(0)),
		dbMock.EXPECT().EndTransaction(),
	)
}

func Expect_initializeApp_Uniswap(dbMock *state.MockStateDB) {
	gomock.InOrder(
		// deploy router
		dbMock.EXPECT().BeginTransaction(gomock.Any()),
		dbMock.EXPECT().GetNonce(gomock.Any()).Return(uint64(0)),
		dbMock.EXPECT().EndTransaction(),

		// deploy tokens, deploy pairs
		dbMock.EXPECT().BeginTransaction(gomock.Any()),
		dbMock.EXPECT().GetNonce(gomock.Any()).Return(uint64(0)),
		dbMock.EXPECT().EndTransaction(),
		dbMock.EXPECT().BeginTransaction(gomock.Any()),
		dbMock.EXPECT().GetNonce(gomock.Any()).Return(uint64(0)),
		dbMock.EXPECT().EndTransaction(),

		// funding accounts
		dbMock.EXPECT().BeginTransaction(gomock.Any()),
		dbMock.EXPECT().GetNonce(gomock.Any()).Return(uint64(0)),
		dbMock.EXPECT().EndTransaction(),

		// waiting for contract deployment requires checking the nonce
		// of funded accounts, since we did no funding, then nonce is 0
		dbMock.EXPECT().BeginTransaction(gomock.Any()),
		dbMock.EXPECT().GetNonce(gomock.Any()).Return(uint64(0)),
		dbMock.EXPECT().EndTransaction(),
	)
}

func TestNormaTxProvider_Run(t *testing.T) {
	tests := []struct {
		name string
		from int
		to   int
		cfg  utils.Config
	}{
		{
			name: "simple-counter",
			from: 1,
			to:   3,
			cfg: utils.Config{
				BlockLength:     uint64(3),
				TxGeneratorType: []string{"counter"},
				ChainID:         297,
				Fork:            "fork",
			},
		},
		{
			name: "simple-erc20",
			from: 1,
			to:   3,
			cfg: utils.Config{
				BlockLength:     uint64(3),
				TxGeneratorType: []string{"erc20"},
				ChainID:         297,
				Fork:            "fork",
			},
		},
		{
			name: "simple-store",
			from: 1,
			to:   3,
			cfg: utils.Config{
				BlockLength:     uint64(3),
				TxGeneratorType: []string{"store"},
				ChainID:         297,
				Fork:            "fork",
			},
		},
		{
			name: "simple-uniswap",
			from: 1,
			to:   10,
			cfg: utils.Config{
				BlockLength:     uint64(3),
				TxGeneratorType: []string{"uniswap"},
				ChainID:         297,
				Fork:            "fork",
			},
		},
		{
			name: "pair-long",
			from: 1,
			to:   234,
			cfg: utils.Config{
				BlockLength:     uint64(5),
				TxGeneratorType: []string{"erc20", "counter"},
				ChainID:         297,
				Fork:            "fork",
			},
		},
		{
			name: "legacy-run-all",
			from: 1,
			to:   3,
			cfg: utils.Config{
				BlockLength:     uint64(5),
				TxGeneratorType: []string{"erc20", "counter", "store"},
				ChainID:         297,
				Fork:            "fork",
			},
		},
		{
			name: "run-all",
			from: 1,
			to:   50,
			cfg: utils.Config{
				BlockLength:     uint64(30),
				TxGeneratorType: []string{"all"},
				ChainID:         297,
				Fork:            "fork",
			},
		},
	}

	expectFunc := map[string]func(dbMock *state.MockStateDB){
		"counter": Expect_initializeApp_Counter,
		"erc20":   Expect_initializeApp_Erc20,
		"store":   Expect_initializeApp_Store,
		"uniswap": Expect_initializeApp_Uniswap,
		"all": func(dbMock *state.MockStateDB) {
			Expect_initializeApp_Erc20(dbMock)
			Expect_initializeApp_Counter(dbMock)
			Expect_initializeApp_Store(dbMock)
			Expect_initializeApp_Uniswap(dbMock)
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			dbMock := state.NewMockStateDB(ctrl)

			provider := NewNormaTxProvider(&test.cfg, dbMock)
			consumer := NewMockTxConsumer(ctrl)

			Expect_initializeTreasureAccount(dbMock)
			Expect_initializeAppContext(dbMock)
			for _, appType := range test.cfg.TxGeneratorType {
				expectFunc[appType](dbMock)
			}

			// make sure that each Consume is called
			var calls []any
			for block := test.from + 1; block <= test.to; block++ {
				for tx := 0; tx < int(test.cfg.BlockLength); tx++ {
					call := consumer.EXPECT().Consume(block, tx, gomock.Any()).Return(nil)
					calls = append(calls, call)
				}
			}
			gomock.InOrder(calls...)

			err := provider.Run(test.from, test.to, toSubstateConsumer(consumer))
			if err != nil {
				t.Fatalf("failed to run provider: %v", err)
			}
		})
	}
}

func TestFakeRpcClient_CodeAt(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStateDb := state.NewMockStateDB(ctrl)
	client := newFakeRpcClient(mockStateDb, nil, 0)
	addr := common.HexToAddress("0x123")
	expectedCode := []byte{0x60, 0x80, 0x60, 0x40}

	t.Run("Success", func(t *testing.T) {
		mockStateDb.EXPECT().BeginTransaction(uint32(0)).Return(nil)
		mockStateDb.EXPECT().GetCode(addr).Return(expectedCode)
		mockStateDb.EXPECT().EndTransaction().Return(nil)

		code, err := client.CodeAt(context.Background(), addr, nil)
		require.NoError(t, err)
		assert.Equal(t, expectedCode, code)
	})

	t.Run("BeginTransactionError", func(t *testing.T) {
		expectedErr := errors.New("begin tx error")
		mockStateDb.EXPECT().BeginTransaction(uint32(0)).Return(expectedErr)

		code, err := client.CodeAt(context.Background(), addr, nil)
		require.Error(t, err)
		assert.Equal(t, expectedErr, err)
		assert.Nil(t, code)
	})

	t.Run("EndTransactionError", func(t *testing.T) {
		expectedErr := errors.New("end tx error")
		mockStateDb.EXPECT().BeginTransaction(uint32(0)).Return(nil)
		mockStateDb.EXPECT().GetCode(addr).Return(expectedCode) // This will still be called
		mockStateDb.EXPECT().EndTransaction().Return(expectedErr)

		code, err := client.CodeAt(context.Background(), addr, nil)
		require.Error(t, err)
		assert.Equal(t, expectedErr, err)
		assert.Nil(t, code) // Code should be nil if EndTransaction fails
	})
}

func TestFakeRpcClient_CallContract(t *testing.T) {
	client := newFakeRpcClient(nil, nil, 0)
	ret, err := client.CallContract(context.Background(), ethereum.CallMsg{}, nil)
	assert.NoError(t, err)
	assert.Nil(t, ret)
}

func TestFakeRpcClient_HeaderByNumber(t *testing.T) {
	client := newFakeRpcClient(nil, nil, 0)
	header, err := client.HeaderByNumber(context.Background(), nil)
	assert.NoError(t, err)
	assert.Equal(t, &types.Header{}, header)
}

func TestFakeRpcClient_PendingNonceAt(t *testing.T) {
	client := newFakeRpcClient(nil, nil, 0)
	nonce, err := client.PendingNonceAt(context.Background(), common.Address{})
	assert.NoError(t, err)
	assert.Equal(t, uint64(0), nonce)
}

func TestFakeRpcClient_SuggestGasTipCap(t *testing.T) {
	client := newFakeRpcClient(nil, nil, 0)
	tipCap, err := client.SuggestGasTipCap(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, big.NewInt(0), tipCap)
}

func TestFakeRpcClient_FilterLogs(t *testing.T) {
	client := newFakeRpcClient(nil, nil, 0)
	logs, err := client.FilterLogs(context.Background(), ethereum.FilterQuery{})
	assert.NoError(t, err)
	assert.Nil(t, logs)
}

func TestFakeRpcClient_SubscribeFilterLogs(t *testing.T) {
	client := newFakeRpcClient(nil, nil, 0)
	sub, err := client.SubscribeFilterLogs(context.Background(), ethereum.FilterQuery{}, nil)
	assert.NoError(t, err)
	assert.Nil(t, sub)
}

func TestFakeRpcClient_Call(t *testing.T) {
	client := newFakeRpcClient(nil, nil, 0)
	err := client.Call(nil, "")
	assert.NoError(t, err)
}

func TestFakeRpcClient_NonceAt(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStateDb := state.NewMockStateDB(ctrl)
	client := newFakeRpcClient(mockStateDb, nil, 0)
	addr := common.HexToAddress("0xabc")
	expectedNonce := uint64(5)

	t.Run("Success", func(t *testing.T) {
		mockStateDb.EXPECT().BeginTransaction(uint32(0)).Return(nil)
		mockStateDb.EXPECT().GetNonce(addr).Return(expectedNonce)
		mockStateDb.EXPECT().EndTransaction().Return(nil)

		nonce, err := client.NonceAt(context.Background(), addr, nil)
		require.NoError(t, err)
		assert.Equal(t, expectedNonce, nonce)
	})

	t.Run("BeginTransactionError", func(t *testing.T) {
		expectedErr := errors.New("begin tx error")
		mockStateDb.EXPECT().BeginTransaction(uint32(0)).Return(expectedErr)

		nonce, err := client.NonceAt(context.Background(), addr, nil)
		require.Error(t, err)
		assert.Equal(t, expectedErr, err)
		assert.Equal(t, uint64(0), nonce)
	})

	t.Run("EndTransactionError", func(t *testing.T) {
		expectedErr := errors.New("end tx error")
		mockStateDb.EXPECT().BeginTransaction(uint32(0)).Return(nil)
		mockStateDb.EXPECT().GetNonce(addr).Return(expectedNonce)
		mockStateDb.EXPECT().EndTransaction().Return(expectedErr)

		nonce, err := client.NonceAt(context.Background(), addr, nil)
		require.Error(t, err)
		assert.Equal(t, expectedErr, err)
		assert.Equal(t, uint64(0), nonce)
	})
}

func TestFakeRpcClient_BalanceAt(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStateDb := state.NewMockStateDB(ctrl)
	client := newFakeRpcClient(mockStateDb, nil, 0)
	addr := common.HexToAddress("0xdef")
	expectedBalance := uint256.NewInt(1000)

	t.Run("Success", func(t *testing.T) {
		mockStateDb.EXPECT().BeginTransaction(uint32(0)).Return(nil)
		mockStateDb.EXPECT().GetBalance(addr).Return(expectedBalance)
		mockStateDb.EXPECT().EndTransaction().Return(nil)

		balance, err := client.BalanceAt(context.Background(), addr, nil)
		require.NoError(t, err)
		assert.Equal(t, expectedBalance.ToBig(), balance)
	})

	t.Run("BeginTransactionError", func(t *testing.T) {
		expectedErr := errors.New("begin tx error")
		mockStateDb.EXPECT().BeginTransaction(uint32(0)).Return(expectedErr)

		balance, err := client.BalanceAt(context.Background(), addr, nil)
		require.Error(t, err)
		assert.Equal(t, expectedErr, err)
		assert.Nil(t, balance)
	})

	t.Run("EndTransactionError", func(t *testing.T) {
		expectedErr := errors.New("end tx error")
		mockStateDb.EXPECT().BeginTransaction(uint32(0)).Return(nil)
		mockStateDb.EXPECT().GetBalance(addr).Return(expectedBalance)
		mockStateDb.EXPECT().EndTransaction().Return(expectedErr)

		balance, err := client.BalanceAt(context.Background(), addr, nil)
		require.Error(t, err)
		assert.Equal(t, expectedErr, err)
		assert.Nil(t, balance)
	})
}

func TestFakeRpcClient_SendTransaction(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Generate a new private key for the sender
	privateKey, err := crypto.GenerateKey()
	require.NoError(t, err)
	senderAddress := crypto.PubkeyToAddress(privateKey.PublicKey)

	chainID := big.NewInt(297) // Example Chain ID from normaTxProvider
	signer := types.NewLondonSigner(chainID)

	t.Run("ContractDeployment_Success", func(t *testing.T) {
		var consumedTx *types.Transaction
		var consumedSender *common.Address

		mockConsumer := func(tx *types.Transaction, sender *common.Address) (bool, error) {
			consumedTx = tx
			consumedSender = sender
			return true, nil
		}
		client := newFakeRpcClient(nil, mockConsumer, 0)

		txData := []byte{0x60, 0x01, 0x60, 0x02}
		gas := uint64(21000)
		gasPrice := big.NewInt(1000000000) // 1 Gwei
		nonce := uint64(0)

		tx := types.NewTx(&types.LegacyTx{
			Nonce:    nonce,
			GasPrice: gasPrice,
			Gas:      gas,
			To:       nil, // Contract deployment
			Value:    big.NewInt(0),
			Data:     txData,
		})
		signedTx, err := types.SignTx(tx, signer, privateKey)
		require.NoError(t, err)

		err = client.SendTransaction(context.Background(), signedTx)
		require.NoError(t, err)

		assert.NotNil(t, consumedTx)
		assert.Equal(t, signedTx.Hash(), consumedTx.Hash())
		assert.Nil(t, consumedSender) // Sender is nil in consumer for SendTransaction

		expectedContractAddress := crypto.CreateAddress(senderAddress, nonce)
		assert.Equal(t, txData, client.pendingCodes[expectedContractAddress])
	})

	t.Run("RegularTransaction_Success", func(t *testing.T) {
		var consumedTx *types.Transaction
		var consumedSender *common.Address

		mockConsumer := func(tx *types.Transaction, sender *common.Address) (bool, error) {
			consumedTx = tx
			consumedSender = sender
			return true, nil
		}
		client := newFakeRpcClient(nil, mockConsumer, 0)
		initialPendingCodesCount := len(client.pendingCodes)

		toAddress := common.HexToAddress("0xRecipient")
		tx := types.NewTx(&types.LegacyTx{
			Nonce:    uint64(1),
			GasPrice: big.NewInt(1000000000),
			Gas:      uint64(21000),
			To:       &toAddress,
			Value:    big.NewInt(100),
			Data:     nil,
		})
		signedTx, err := types.SignTx(tx, signer, privateKey)
		require.NoError(t, err)

		err = client.SendTransaction(context.Background(), signedTx)
		require.NoError(t, err)

		assert.NotNil(t, consumedTx)
		assert.Equal(t, signedTx.Hash(), consumedTx.Hash())
		assert.Nil(t, consumedSender)
		assert.Equal(t, initialPendingCodesCount, len(client.pendingCodes), "Pending codes should not change for regular tx")
	})

	t.Run("ConsumerError", func(t *testing.T) {
		expectedErr := errors.New("consumer failed")
		mockConsumer := func(tx *types.Transaction, sender *common.Address) (bool, error) {
			return false, expectedErr
		}
		client := newFakeRpcClient(nil, mockConsumer, 0)

		tx := types.NewTx(&types.LegacyTx{Nonce: 0, To: &common.Address{}}) // Minimal valid tx
		signedTx, err := types.SignTx(tx, signer, privateKey)
		require.NoError(t, err)

		err = client.SendTransaction(context.Background(), signedTx)
		require.Error(t, err)
		assert.Equal(t, expectedErr, err)
	})
}

func TestFakeRpcClient_ChainID(t *testing.T) {
	client := newFakeRpcClient(nil, nil, 123)
	chainId, err := client.ChainID(context.Background())
	assert.Equal(t, chainId, big.NewInt(int64(123)))
	assert.Nil(t, err)
}

func TestFakeRpcClient_TransactionReceipt(t *testing.T) {
	client := newFakeRpcClient(nil, nil, 0)
	receipt, err := client.TransactionReceipt(context.Background(), common.Hash{})
	assert.Equal(t, receipt, &types.Receipt{Status: types.ReceiptStatusSuccessful})
	require.NoError(t, err)
}

func TestFakeRpcClient_WaitTransactionReceipt(t *testing.T) {
	client := newFakeRpcClient(nil, nil, 0)
	receipt, err := client.WaitTransactionReceipt(common.Hash{})
	assert.Equal(t, receipt, &types.Receipt{Status: types.ReceiptStatusSuccessful})
	require.NoError(t, err)
}

func TestGenerateUsers(t *testing.T) {
	tests := []struct {
		name      string
		cfg       utils.Config
		userCount int
	}{
		{
			name: "single-counter",
			cfg: utils.Config{
				TxGeneratorType: []string{"counter"},
				ChainID:         297,
			},
			userCount: 1,
		},
		{
			name: "double-counter",
			cfg: utils.Config{
				TxGeneratorType: []string{"counter", "counter"},
				ChainID:         297,
			},
			userCount: 2,
		},
		{
			name: "all",
			cfg: utils.Config{
				TxGeneratorType: []string{"all"},
				ChainID:         297,
			},
			userCount: 4,
		},
		{
			name: "all-supported-app-types",
			cfg: utils.Config{
				TxGeneratorType: []string{"counter", "erc20", "store", "uniswap"},
				ChainID:         297,
			},
			userCount: 4,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)

			dbMock := state.NewMockStateDB(ctrl)
			dbMock.EXPECT().BeginTransaction(gomock.Any()).AnyTimes()
			dbMock.EXPECT().GetNonce(gomock.Any()).Return(uint64(0)).AnyTimes()
			dbMock.EXPECT().EndTransaction().AnyTimes()

			primaryAccount, err := app.NewAccount(0, PrivateKey, nil, int64(test.cfg.ChainID))
			if err != nil {
				t.Fatalf("failed to create primary account: %v", err)
			}

			doNothing := func(tx *types.Transaction, sender *common.Address) (bool, error) {
				return true, nil
			}
			doNothingRpc := newFakeRpcClient(dbMock, doNothing, int64(test.cfg.ChainID))
			defer doNothingRpc.Close()

			appContext, err := app.NewContext(fakeRpcClientFactory{client: doNothingRpc}, primaryAccount)
			if err != nil {
				t.Fatalf("failed to create app context: %v", err)
			}

			users, err := generateUsers(appContext, test.cfg.TxGeneratorType)
			if err != nil {
				t.Fatalf("failed to create users: %v", err)
			}

			assert.Equal(t, len(users), test.userCount)
		})
	}
}

func TestRoundRobinSelector(t *testing.T) {
	tests := []struct {
		name string
		cfg  utils.Config
	}{
		{
			name: "single-counter",
			cfg: utils.Config{
				TxGeneratorType: []string{"counter"},
			},
		},
		{
			name: "all-supported-app-types",
			cfg: utils.Config{
				TxGeneratorType: []string{"counter", "erc20", "store", "uniswap"},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			nextType := RoundRobinSelector[string](test.cfg.TxGeneratorType)

			n := len(test.cfg.TxGeneratorType)
			for i := 0; i < 300; i++ {
				assert.Equal(t, nextType(), test.cfg.TxGeneratorType[i%n])
			}
		})
	}
}

func TestNormaConsumerBoundary(t *testing.T) {
	tests := []struct {
		name  string
		cfg   normaConsumerConfig
		types []string
	}{
		{
			name: "mini-blocks",
			cfg: normaConsumerConfig{
				fromBlock:   1,
				toBlock:     5,
				blockLength: 3,
			},
		},
		{
			name: "long-blocks",
			cfg: normaConsumerConfig{
				fromBlock:   1,
				toBlock:     10_000,
				blockLength: 2,
			},
		},
		{
			name: "wide-blocks",
			cfg: normaConsumerConfig{
				fromBlock:   1,
				toBlock:     5,
				blockLength: 297,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			dbMock := state.NewMockStateDB(ctrl)
			dbMock.EXPECT().BeginTransaction(gomock.Any()).AnyTimes()
			dbMock.EXPECT().GetNonce(gomock.Any()).Return(uint64(0)).AnyTimes()
			dbMock.EXPECT().EndTransaction().AnyTimes()

			provider := &normaTxProvider{
				cfg: &utils.Config{
					BlockLength: test.cfg.blockLength,
					ChainID:     297,
				},
				stateDb: dbMock,
			}

			// Generate a new private key for the sender
			privateKey, err := crypto.GenerateKey()
			require.NoError(t, err)
			tx := types.NewTx(&types.DynamicFeeTx{ChainID: big.NewInt(297)})
			signedTx, err := types.SignTx(tx, types.NewLondonSigner(big.NewInt(297)), privateKey)
			require.NoError(t, err)

			mockUser := app.NewMockUser(ctrl)
			mockUser.EXPECT().GenerateTx().Return(signedTx, nil).AnyTimes()

			consumeCount := int(0)
			counter := func(TransactionInfo[txcontext.TxContext]) error {
				consumeCount++
				return nil
			}
			normaConsumer := newNormaConsumer(counter, test.cfg)

			provider.run([]app.User{mockUser}, normaConsumer)

			// expected: block length * (from - to + 1)
			expected := int(test.cfg.blockLength) * (test.cfg.toBlock - test.cfg.fromBlock + 1)
			assert.Equal(t, expected, consumeCount)
		})
	}
}
