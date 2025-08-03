package state

import (
	"testing"

	"github.com/0xsoniclabs/aida/txcontext"
	"github.com/ethereum/go-ethereum/common"
	"github.com/holiman/uint256"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestState_NewOffTheChainStateDB(t *testing.T) {
	d := NewOffTheChainStateDB()
	assert.NotNil(t, d)

}
func TestState_MakeOffTheChainStateDB(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockWs := txcontext.NewMockWorldState(ctrl)
	mockAcc := txcontext.NewMockAccount(ctrl)
	chainConduit := &ChainConduit{}
	mockAcc.EXPECT().GetCode().Return([]byte{})
	mockAcc.EXPECT().GetNonce().Return(uint64(1))
	mockAcc.EXPECT().GetBalance().Return(uint256.NewInt(100))
	mockAcc.EXPECT().ForEachStorage(gomock.Any()).Do(func(ff txcontext.StorageHandler) {
		ff(common.HexToHash("0x1234"), common.HexToHash("0x5678"))
	})
	mockWs.EXPECT().ForEachAccount(gomock.Any()).Do(func(ff txcontext.AccountHandler) {
		ff(common.HexToAddress("0x1234"), mockAcc)
	})
	d, err := MakeOffTheChainStateDB(mockWs, 0, chainConduit)
	assert.NoError(t, err)
	assert.NotNil(t, d)
}
