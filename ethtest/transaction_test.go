package ethtest

import (
	"encoding/json"
	"github.com/ethereum/go-ethereum/core/types"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/stretchr/testify/assert"
)

func TestStTransaction_ToMessage(t *testing.T) {
	post := stPost{
		RootHash:        common.HexToHash("0x1234"),
		LogsHash:        common.HexToHash("0xabcd"),
		TxBytes:         hexutil.Bytes{0x01, 0x02},
		ExpectException: "err",
		Indexes:         Index{Data: 0, Gas: 0, Value: 0},
	}
	stAuth := stAuthorization{
		ChainID: big.NewInt(1),
		Address: common.HexToAddress("0x9abc"),
		Nonce:   1,
		V:       27,
		R:       big.NewInt(123456789),
		S:       big.NewInt(987654321),
	}
	to := common.HexToAddress("0x1234")
	bytesTo, err := to.MarshalText()
	if err != nil {
		t.Fatalf("failed to marshal address: %v", err)
	}
	st := stTransaction{
		Data:                 []string{"0x1234"},
		Value:                []string{"1234"},
		GasLimit:             []*BigInt{newBigInt(1000000)},
		Nonce:                newBigInt(1),
		MaxFeePerGas:         newBigInt(1),
		MaxPriorityFeePerGas: newBigInt(1),
		BlobGasFeeCap:        newBigInt(1),
		To:                   string(bytesTo),
		AuthorizationList: []*stAuthorization{
			&stAuth,
		},
	}
	msg, err := st.toMessage(post, newBigInt(1))
	assert.NoError(t, err)
	assert.Equal(t, &to, msg.To)
	assert.Equal(t, st.Nonce.Uint64(), msg.Nonce)
	assert.Equal(t, st.BlobGasFeeCap.Convert(), msg.BlobGasFeeCap)
	assert.Equal(t, []byte{0x12, 0x34}, msg.Data)
	assert.Equal(t, st.Value[0], msg.Value.String())
	assert.Equal(t, *st.AuthorizationList[0], stAuth)
}

func TestStTransaction_ToMessageError(t *testing.T) {
	t.Run("error invalid private key", func(t *testing.T) {
		post := stPost{
			RootHash:        common.HexToHash("0x1234"),
			LogsHash:        common.HexToHash("0xabcd"),
			TxBytes:         hexutil.Bytes{0x01, 0x02},
			ExpectException: "err",
			Indexes:         Index{Data: 0, Gas: 0, Value: 0},
		}
		st := stTransaction{
			To:         "1234",
			PrivateKey: make(hexutil.Bytes, 32),
		}
		msg, err := st.toMessage(post, newBigInt(1))
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid private key")
		assert.Nil(t, msg)
	})
	t.Run("error address", func(t *testing.T) {
		post := stPost{
			RootHash:        common.HexToHash("0x1234"),
			LogsHash:        common.HexToHash("0xabcd"),
			TxBytes:         hexutil.Bytes{0x01, 0x02},
			ExpectException: "err",
			Indexes:         Index{Data: 0, Gas: 0, Value: 0},
		}
		st := stTransaction{
			To: "1234",
		}
		msg, err := st.toMessage(post, newBigInt(1))
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid to address")
		assert.Nil(t, msg)
	})

	t.Run("error out of bound data", func(t *testing.T) {
		post := stPost{
			RootHash:        common.HexToHash("0x1234"),
			LogsHash:        common.HexToHash("0xabcd"),
			TxBytes:         hexutil.Bytes{0x01, 0x02},
			ExpectException: "err",
			Indexes:         Index{Data: 1, Gas: 0, Value: 0},
		}
		st := stTransaction{}
		msg, err := st.toMessage(post, newBigInt(1))
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "out of bounds")
		assert.Nil(t, msg)
	})

	t.Run("error out of bound value", func(t *testing.T) {
		post := stPost{
			RootHash:        common.HexToHash("0x1234"),
			LogsHash:        common.HexToHash("0xabcd"),
			TxBytes:         hexutil.Bytes{0x01, 0x02},
			ExpectException: "err",
			Indexes:         Index{Data: 0, Gas: 1, Value: 0},
		}
		st := stTransaction{}
		msg, err := st.toMessage(post, newBigInt(1))
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "out of bounds")
		assert.Nil(t, msg)
	})

	t.Run("error out of bound gas", func(t *testing.T) {
		post := stPost{
			RootHash:        common.HexToHash("0x1234"),
			LogsHash:        common.HexToHash("0xabcd"),
			TxBytes:         hexutil.Bytes{0x01, 0x02},
			ExpectException: "err",
			Indexes:         Index{Data: 0, Gas: 0, Value: 1},
		}
		st := stTransaction{}
		msg, err := st.toMessage(post, newBigInt(1))
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "out of bounds")
		assert.Nil(t, msg)
	})

	t.Run("error invalid tx value", func(t *testing.T) {
		post := stPost{
			RootHash:        common.HexToHash("0x1234"),
			LogsHash:        common.HexToHash("0xabcd"),
			TxBytes:         hexutil.Bytes{0x01, 0x02},
			ExpectException: "err",
			Indexes:         Index{Data: 0, Gas: 0, Value: 0},
		}
		to := common.HexToAddress("0x1234")
		bytesTo, err := to.MarshalText()
		if err != nil {
			t.Fatalf("failed to marshal address: %v", err)
		}
		st := stTransaction{
			Data:                 []string{"abcdef"},
			Value:                []string{"abcdef"},
			GasLimit:             []*BigInt{newBigInt(1000000)},
			Nonce:                newBigInt(1),
			MaxFeePerGas:         newBigInt(1),
			MaxPriorityFeePerGas: newBigInt(1),
			BlobGasFeeCap:        newBigInt(1),
			To:                   string(bytesTo),
		}
		msg, err := st.toMessage(post, newBigInt(1))
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid tx value")
		assert.Nil(t, msg)
	})

	t.Run("error invalid tx data", func(t *testing.T) {
		post := stPost{
			RootHash:        common.HexToHash("0x1234"),
			LogsHash:        common.HexToHash("0xabcd"),
			TxBytes:         hexutil.Bytes{0x01, 0x02},
			ExpectException: "err",
			Indexes:         Index{Data: 0, Gas: 0, Value: 0},
		}
		to := common.HexToAddress("0x1234")
		bytesTo, err := to.MarshalText()
		if err != nil {
			t.Fatalf("failed to marshal address: %v", err)
		}
		st := stTransaction{
			Data:                 []string{"-1234"},
			Value:                []string{"1234"},
			GasLimit:             []*BigInt{newBigInt(1000000)},
			Nonce:                newBigInt(1),
			MaxFeePerGas:         newBigInt(1),
			MaxPriorityFeePerGas: newBigInt(1),
			BlobGasFeeCap:        newBigInt(1),
			To:                   string(bytesTo),
		}
		msg, err := st.toMessage(post, newBigInt(1))
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid tx data")
		assert.Nil(t, msg)
	})

	t.Run("error no gas price", func(t *testing.T) {
		post := stPost{
			RootHash:        common.HexToHash("0x1234"),
			LogsHash:        common.HexToHash("0xabcd"),
			TxBytes:         hexutil.Bytes{0x01, 0x02},
			ExpectException: "err",
			Indexes:         Index{Data: 0, Gas: 0, Value: 0},
		}
		to := common.HexToAddress("0x1234")
		bytesTo, err := to.MarshalText()
		if err != nil {
			t.Fatalf("failed to marshal address: %v", err)
		}
		st := stTransaction{
			Data:                 []string{"0x1234"},
			Value:                []string{"1234"},
			GasLimit:             []*BigInt{newBigInt(1000000)},
			Nonce:                newBigInt(1),
			MaxFeePerGas:         newBigInt(1),
			MaxPriorityFeePerGas: newBigInt(1),
			BlobGasFeeCap:        newBigInt(1),
			To:                   string(bytesTo),
		}
		msg, err := st.toMessage(post, nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no gas price provided")
		assert.Nil(t, msg)
	})

}

func TestEthTest_bigMin(t *testing.T) {
	a := big.NewInt(1)
	b := big.NewInt(2)
	assert.Equal(t, a, bigMin(a, b))
	assert.Equal(t, a, bigMin(b, a))
	assert.Equal(t, b, bigMin(b, b))
}

func TestStTransaction_MarshallAndUnmarshall(t *testing.T) {
	sender := common.HexToAddress("0x5678")
	st := stTransaction{
		Data:                 []string{"0x1234"},
		Value:                []string{"1234"},
		GasLimit:             []*BigInt{newBigInt(1000000)},
		GasPrice:             newBigInt(1),
		Nonce:                newBigInt(2),
		MaxFeePerGas:         newBigInt(3),
		MaxPriorityFeePerGas: newBigInt(4),
		BlobGasFeeCap:        newBigInt(5),
		To:                   "0x1234",
		AccessLists: []*types.AccessList{
			{
				{
					Address:     common.HexToAddress("0x1234"),
					StorageKeys: []common.Hash{common.HexToHash("0x5678")},
				},
			},
		},
		PrivateKey: hexutil.MustDecode("0x1234"),
		Sender:     &sender,
		AuthorizationList: []*stAuthorization{
			{
				ChainID: big.NewInt(1),
				Address: common.HexToAddress("0x9abc"),
				Nonce:   1,
				V:       27,
				R:       big.NewInt(123456789),
				S:       big.NewInt(987654321),
			},
		},
		BlobHashes: []common.Hash{
			common.HexToHash("0xabcdef"),
		},
	}

	marshal, err := json.Marshal(st)
	assert.NoError(t, err)
	var unmarshalled stTransaction
	err = json.Unmarshal(marshal, &unmarshalled)
	assert.NoError(t, err)

	assert.Equal(t, st, unmarshalled)
}
