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
	"github.com/0xsoniclabs/aida/utils"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/params"
	"github.com/holiman/uint256"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestNormaTxProvider_Run(t *testing.T) {
	ctrl := gomock.NewController(t)
	dbMock := state.NewMockStateDB(ctrl)

	cfg := &utils.Config{
		BlockLength:     uint64(3),
		TxGeneratorType: []string{"counter"},
		ChainID:         297,
	}
	provider := NewNormaTxProvider(cfg, dbMock)
	consumer := NewMockTxConsumer(ctrl)

	gomock.InOrder(
		// treasure account initialization
		dbMock.EXPECT().BeginBlock(gomock.Any()),
		dbMock.EXPECT().BeginTransaction(gomock.Any()),
		dbMock.EXPECT().CreateAccount(gomock.Any()),
		dbMock.EXPECT().AddBalance(gomock.Any(), gomock.Any(), gomock.Any()),
		dbMock.EXPECT().EndTransaction(),
		dbMock.EXPECT().EndBlock(),

		// contract deployment

		// expected on block 2, because block 1 is treasure account initialization
		// and we are starting from block 1
		consumer.EXPECT().Consume(2, 0, gomock.Any()).Return(nil),
		dbMock.EXPECT().BeginTransaction(gomock.Any()),
		dbMock.EXPECT().GetNonce(gomock.Any()).Return(uint64(0)),
		dbMock.EXPECT().EndTransaction(),

		// funding accounts
		// we return a lot of tokens, so we don't have to "fund" them
		dbMock.EXPECT().BeginTransaction(gomock.Any()),
		dbMock.EXPECT().GetBalance(gomock.Any()).Return(uint256.NewInt(0).Mul(uint256.NewInt(params.Ether), uint256.NewInt(1_000_000))),
		dbMock.EXPECT().EndTransaction(),
		// nonce for account deploying the contract has to be 1
		dbMock.EXPECT().BeginTransaction(gomock.Any()),
		dbMock.EXPECT().GetNonce(gomock.Any()).Return(uint64(1)),
		dbMock.EXPECT().EndTransaction(),
		// nonce for funded accounts has to be 0
		dbMock.EXPECT().BeginTransaction(gomock.Any()),
		dbMock.EXPECT().GetNonce(gomock.Any()).Return(uint64(0)),
		dbMock.EXPECT().EndTransaction(),
		dbMock.EXPECT().BeginTransaction(gomock.Any()),
		dbMock.EXPECT().GetBalance(gomock.Any()).Return(uint256.NewInt(0).Mul(uint256.NewInt(params.Ether), uint256.NewInt(1_000_000))),
		dbMock.EXPECT().EndTransaction(),

		// waiting for contract deployment requires checking the nonce
		// of funded accounts, since we did no funding, then nonce is 0
		dbMock.EXPECT().BeginTransaction(gomock.Any()),
		dbMock.EXPECT().GetNonce(gomock.Any()).Return(uint64(0)),
		dbMock.EXPECT().EndTransaction(),

		// generating transactions, starting from transaction 1 (0 was contract deployment)
		consumer.EXPECT().Consume(2, 1, gomock.Any()).Return(nil),
		consumer.EXPECT().Consume(2, 2, gomock.Any()).Return(nil),
		consumer.EXPECT().Consume(3, 0, gomock.Any()).Return(nil),
		consumer.EXPECT().Consume(3, 1, gomock.Any()).Return(nil),
		consumer.EXPECT().Consume(3, 2, gomock.Any()).Return(nil),
	)

	err := provider.Run(1, 3, toSubstateConsumer(consumer))
	if err != nil {
		t.Fatalf("failed to run provider: %v", err)
	}
}

func TestNormaTxProvider_RunAll(t *testing.T) {
	ctrl := gomock.NewController(t)
	dbMock := state.NewMockStateDB(ctrl)

	cfg := &utils.Config{
		BlockLength:     uint64(5),
		TxGeneratorType: []string{"erc20", "counter", "store"},
		ChainID:         297,
	}
	provider := NewNormaTxProvider(cfg, dbMock)
	consumer := NewMockTxConsumer(ctrl)

	balance := uint256.NewInt(0).Mul(uint256.NewInt(params.Ether), uint256.NewInt(1_000_000))

	gomock.InOrder(
		// treasure account initialization
		dbMock.EXPECT().BeginBlock(gomock.Any()),
		dbMock.EXPECT().BeginTransaction(gomock.Any()),
		dbMock.EXPECT().CreateAccount(gomock.Any()),
		dbMock.EXPECT().AddBalance(gomock.Any(), gomock.Any(), gomock.Any()),
		dbMock.EXPECT().EndTransaction(),
		dbMock.EXPECT().EndBlock(),

		// contract deployment in order: erc20 -> counter -> store

		// expected on block 2, because block 1 is treasure account initialization
		// and we are starting from block 1

		// ERC 20
		consumer.EXPECT().Consume(2, 0, gomock.Any()).Return(nil),
		dbMock.EXPECT().BeginTransaction(gomock.Any()),
		dbMock.EXPECT().GetNonce(gomock.Any()).Return(uint64(0)),
		dbMock.EXPECT().EndTransaction(),
		dbMock.EXPECT().BeginTransaction(gomock.Any()),
		dbMock.EXPECT().GetBalance(gomock.Any()).Return(balance),
		dbMock.EXPECT().EndTransaction(),
		dbMock.EXPECT().BeginTransaction(gomock.Any()),
		dbMock.EXPECT().GetNonce(gomock.Any()).Return(uint64(1)),
		dbMock.EXPECT().EndTransaction(),
		// funding accounts
		dbMock.EXPECT().BeginTransaction(gomock.Any()),
		dbMock.EXPECT().GetNonce(gomock.Any()).Return(uint64(0)),
		dbMock.EXPECT().EndTransaction(),
		dbMock.EXPECT().BeginTransaction(gomock.Any()),
		dbMock.EXPECT().GetBalance(gomock.Any()).Return(balance),
		dbMock.EXPECT().EndTransaction(),
		// mint nf tokens
		consumer.EXPECT().Consume(2, 1, gomock.Any()).Return(nil),
		dbMock.EXPECT().BeginTransaction(gomock.Any()),
		dbMock.EXPECT().GetNonce(gomock.Any()).Return(uint64(1)),
		dbMock.EXPECT().EndTransaction(),
		// COUNTER
		consumer.EXPECT().Consume(2, 2, gomock.Any()).Return(nil),
		dbMock.EXPECT().BeginTransaction(gomock.Any()),
		dbMock.EXPECT().GetNonce(gomock.Any()).Return(uint64(1)),
		dbMock.EXPECT().EndTransaction(),
		dbMock.EXPECT().BeginTransaction(gomock.Any()),
		dbMock.EXPECT().GetBalance(gomock.Any()).Return(balance),
		dbMock.EXPECT().EndTransaction(),
		dbMock.EXPECT().BeginTransaction(gomock.Any()),
		dbMock.EXPECT().GetNonce(gomock.Any()).Return(uint64(2)),
		dbMock.EXPECT().EndTransaction(),
		// funding accounts
		dbMock.EXPECT().BeginTransaction(gomock.Any()),
		dbMock.EXPECT().GetNonce(gomock.Any()).Return(uint64(0)),
		dbMock.EXPECT().EndTransaction(),
		dbMock.EXPECT().BeginTransaction(gomock.Any()),
		dbMock.EXPECT().GetBalance(gomock.Any()).Return(balance),
		dbMock.EXPECT().EndTransaction(),
		dbMock.EXPECT().BeginTransaction(gomock.Any()),
		dbMock.EXPECT().GetNonce(gomock.Any()).Return(uint64(1)),
		dbMock.EXPECT().EndTransaction(),
		// STORE
		consumer.EXPECT().Consume(2, 3, gomock.Any()).Return(nil),
		dbMock.EXPECT().BeginTransaction(gomock.Any()),
		dbMock.EXPECT().GetNonce(gomock.Any()).Return(uint64(2)),
		dbMock.EXPECT().EndTransaction(),
		dbMock.EXPECT().BeginTransaction(gomock.Any()),
		dbMock.EXPECT().GetBalance(gomock.Any()).Return(balance),
		dbMock.EXPECT().EndTransaction(),
		dbMock.EXPECT().BeginTransaction(gomock.Any()),
		dbMock.EXPECT().GetNonce(gomock.Any()).Return(uint64(3)),
		dbMock.EXPECT().EndTransaction(),
		// funding accounts
		dbMock.EXPECT().BeginTransaction(gomock.Any()),
		dbMock.EXPECT().GetNonce(gomock.Any()).Return(uint64(0)),
		dbMock.EXPECT().EndTransaction(),
		dbMock.EXPECT().BeginTransaction(gomock.Any()),
		dbMock.EXPECT().GetBalance(gomock.Any()).Return(balance),
		dbMock.EXPECT().EndTransaction(),
		dbMock.EXPECT().BeginTransaction(gomock.Any()),
		dbMock.EXPECT().GetNonce(gomock.Any()).Return(uint64(2)),
		dbMock.EXPECT().EndTransaction(),
		// generating transactions
		consumer.EXPECT().Consume(2, 4, gomock.Any()).Return(nil),
		consumer.EXPECT().Consume(3, 0, gomock.Any()).Return(nil),
		consumer.EXPECT().Consume(3, 1, gomock.Any()).Return(nil),
		consumer.EXPECT().Consume(3, 2, gomock.Any()).Return(nil),
		consumer.EXPECT().Consume(3, 3, gomock.Any()).Return(nil),
		consumer.EXPECT().Consume(3, 4, gomock.Any()).Return(nil),
	)

	err := provider.Run(1, 3, toSubstateConsumer(consumer))
	if err != nil {
		t.Fatalf("failed to run provider: %v", err)
	}
}

func TestFakeRpcClient_CodeAt(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStateDb := state.NewMockStateDB(ctrl)
	client := newFakeRpcClient(mockStateDb, nil)
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
	client := newFakeRpcClient(nil, nil)
	ret, err := client.CallContract(context.Background(), ethereum.CallMsg{}, nil)
	assert.NoError(t, err)
	assert.Nil(t, ret)
}

func TestFakeRpcClient_HeaderByNumber(t *testing.T) {
	client := newFakeRpcClient(nil, nil)
	header, err := client.HeaderByNumber(context.Background(), nil)
	assert.NoError(t, err)
	assert.Equal(t, &types.Header{}, header)
}

func TestFakeRpcClient_PendingNonceAt(t *testing.T) {
	client := newFakeRpcClient(nil, nil)
	nonce, err := client.PendingNonceAt(context.Background(), common.Address{})
	assert.NoError(t, err)
	assert.Equal(t, uint64(0), nonce)
}

func TestFakeRpcClient_SuggestGasTipCap(t *testing.T) {
	client := newFakeRpcClient(nil, nil)
	tipCap, err := client.SuggestGasTipCap(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, big.NewInt(0), tipCap)
}

func TestFakeRpcClient_FilterLogs(t *testing.T) {
	client := newFakeRpcClient(nil, nil)
	logs, err := client.FilterLogs(context.Background(), ethereum.FilterQuery{})
	assert.NoError(t, err)
	assert.Nil(t, logs)
}

func TestFakeRpcClient_SubscribeFilterLogs(t *testing.T) {
	client := newFakeRpcClient(nil, nil)
	sub, err := client.SubscribeFilterLogs(context.Background(), ethereum.FilterQuery{}, nil)
	assert.NoError(t, err)
	assert.Nil(t, sub)
}

func TestFakeRpcClient_Call(t *testing.T) {
	client := newFakeRpcClient(nil, nil)
	err := client.Call(nil, "")
	assert.NoError(t, err)
}

func TestFakeRpcClient_NonceAt(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStateDb := state.NewMockStateDB(ctrl)
	client := newFakeRpcClient(mockStateDb, nil)
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
	client := newFakeRpcClient(mockStateDb, nil)
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
	signer := types.NewEIP155Signer(chainID)

	t.Run("ContractDeployment_Success", func(t *testing.T) {
		var consumedTx *types.Transaction
		var consumedSender *common.Address

		mockConsumer := func(tx *types.Transaction, sender *common.Address) error {
			consumedTx = tx
			consumedSender = sender
			return nil
		}
		client := newFakeRpcClient(nil, mockConsumer)

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

		mockConsumer := func(tx *types.Transaction, sender *common.Address) error {
			consumedTx = tx
			consumedSender = sender
			return nil
		}
		client := newFakeRpcClient(nil, mockConsumer)
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
		mockConsumer := func(tx *types.Transaction, sender *common.Address) error {
			return expectedErr
		}
		client := newFakeRpcClient(nil, mockConsumer)

		tx := types.NewTx(&types.LegacyTx{Nonce: 0, To: &common.Address{}}) // Minimal valid tx
		signedTx, err := types.SignTx(tx, signer, privateKey)
		require.NoError(t, err)

		err = client.SendTransaction(context.Background(), signedTx)
		require.Error(t, err)
		assert.Equal(t, expectedErr, err)
	})
}
