package trace

import (
	"github.com/0xsoniclabs/aida/executor"
	"github.com/0xsoniclabs/aida/state"
	"github.com/0xsoniclabs/aida/tracer"
	"github.com/0xsoniclabs/aida/txcontext"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/tracing"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/params"
	"github.com/holiman/uint256"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	big "math/big"
	"testing"
)

func TestTraceProcessor_Process_AllCases(t *testing.T) {
	addr := common.HexToAddress("0x1")
	key := common.HexToHash("0x2")
	value := common.HexToHash("0x3")
	uint64Data := uint64(125)
	uint32Data := uint32(42)
	boolData := true
	hash := common.HexToHash("0xabc")
	code := []byte{0x1, 0x2}

	tests := []struct {
		name  string
		op    tracer.Operation
		setup func(reader *tracer.MockFileReader, state *state.MockStateDB)
	}{
		{
			name: "AddBalance",
			op:   tracer.Operation{Op: tracer.AddBalanceID, Addr: addr},
			setup: func(reader *tracer.MockFileReader, state *state.MockStateDB) {
				balance := uint256.NewInt(100)
				reader.EXPECT().ReadBalanceChange().Return(balance, tracing.BalanceChangeTransfer, nil)
				state.EXPECT().AddBalance(addr, balance, tracing.BalanceChangeTransfer)
			},
		},
		{
			name: "BeginBlock",
			op:   tracer.Operation{Op: tracer.BeginBlockID},
			setup: func(reader *tracer.MockFileReader, state *state.MockStateDB) {
				state.EXPECT().BeginBlock(uint64Data).Return(nil)
			},
		},
		{
			name: "BeginSyncPeriod",
			op:   tracer.Operation{Op: tracer.BeginSyncPeriodID},
			setup: func(reader *tracer.MockFileReader, state *state.MockStateDB) {
				reader.EXPECT().ReadUint64().Return(uint64Data, nil)
				state.EXPECT().BeginSyncPeriod(uint64Data)
			},
		},
		{
			name: "BeginTransaction",
			op:   tracer.Operation{Op: tracer.BeginTransactionID},
			setup: func(reader *tracer.MockFileReader, state *state.MockStateDB) {
				state.EXPECT().BeginTransaction(uint32Data).Return(nil)
			},
		},
		{
			name: "CreateAccount",
			op:   tracer.Operation{Op: tracer.CreateAccountID, Addr: addr},
			setup: func(reader *tracer.MockFileReader, state *state.MockStateDB) {
				state.EXPECT().CreateAccount(addr)
			},
		},
		{
			name: "Empty",
			op:   tracer.Operation{Op: tracer.EmptyID, Addr: addr},
			setup: func(reader *tracer.MockFileReader, state *state.MockStateDB) {
				state.EXPECT().Empty(addr)
			},
		},
		{
			name: "EndBlock",
			op:   tracer.Operation{Op: tracer.EndBlockID},
			setup: func(reader *tracer.MockFileReader, state *state.MockStateDB) {
				state.EXPECT().EndBlock().Return(nil)
			},
		},
		{
			name: "EndSyncPeriod",
			op:   tracer.Operation{Op: tracer.EndSyncPeriodID},
			setup: func(reader *tracer.MockFileReader, state *state.MockStateDB) {
				state.EXPECT().EndSyncPeriod()
			},
		},
		{
			name: "EndTransaction",
			op:   tracer.Operation{Op: tracer.EndTransactionID},
			setup: func(reader *tracer.MockFileReader, state *state.MockStateDB) {
				state.EXPECT().EndTransaction().Return(nil)
			},
		},
		{
			name: "Exist",
			op:   tracer.Operation{Op: tracer.ExistID, Addr: addr},
			setup: func(reader *tracer.MockFileReader, state *state.MockStateDB) {
				state.EXPECT().Exist(addr)
			},
		},
		{
			name: "GetBalance",
			op:   tracer.Operation{Op: tracer.GetBalanceID, Addr: addr},
			setup: func(reader *tracer.MockFileReader, state *state.MockStateDB) {
				state.EXPECT().GetBalance(addr)
			},
		},
		{
			name: "GetCodeHash",
			op:   tracer.Operation{Op: tracer.GetCodeHashID, Addr: addr},
			setup: func(reader *tracer.MockFileReader, state *state.MockStateDB) {
				state.EXPECT().GetCodeHash(addr)
			},
		},
		{
			name: "GetCode",
			op:   tracer.Operation{Op: tracer.GetCodeID, Addr: addr},
			setup: func(reader *tracer.MockFileReader, state *state.MockStateDB) {
				state.EXPECT().GetCode(addr)
			},
		},
		{
			name: "GetCodeSize",
			op:   tracer.Operation{Op: tracer.GetCodeSizeID, Addr: addr},
			setup: func(reader *tracer.MockFileReader, state *state.MockStateDB) {
				state.EXPECT().GetCodeSize(addr)
			},
		},
		{
			name: "GetCommittedState",
			op:   tracer.Operation{Op: tracer.GetCommittedStateID, Addr: addr, Key: key},
			setup: func(reader *tracer.MockFileReader, state *state.MockStateDB) {
				state.EXPECT().GetCommittedState(addr, key)
			},
		},
		{
			name: "GetNonce",
			op:   tracer.Operation{Op: tracer.GetNonceID, Addr: addr},
			setup: func(reader *tracer.MockFileReader, state *state.MockStateDB) {
				state.EXPECT().GetNonce(addr)
			},
		},
		{
			name: "GetState",
			op:   tracer.Operation{Op: tracer.GetStateID, Addr: addr, Key: key},
			setup: func(reader *tracer.MockFileReader, state *state.MockStateDB) {
				state.EXPECT().GetState(addr, key)
			},
		},
		{
			name: "HasSelfDestructed",
			op:   tracer.Operation{Op: tracer.HasSelfDestructedID, Addr: addr},
			setup: func(reader *tracer.MockFileReader, state *state.MockStateDB) {
				state.EXPECT().HasSelfDestructed(addr)
			},
		},
		{
			name: "RevertToSnapshot",
			op:   tracer.Operation{Op: tracer.RevertToSnapshotID},
			setup: func(reader *tracer.MockFileReader, state *state.MockStateDB) {
				reader.EXPECT().ReadUint32().Return(uint32Data, nil)
				state.EXPECT().RevertToSnapshot(int(uint32Data))
			},
		},
		{
			name: "SetCode",
			op:   tracer.Operation{Op: tracer.SetCodeID, Addr: addr},
			setup: func(reader *tracer.MockFileReader, state *state.MockStateDB) {
				reader.EXPECT().ReadVariableSizeData().Return(code, nil)
				state.EXPECT().SetCode(addr, code)
			},
		},
		{
			name: "SetNonce",
			op:   tracer.Operation{Op: tracer.SetNonceID, Addr: addr},
			setup: func(reader *tracer.MockFileReader, state *state.MockStateDB) {
				reader.EXPECT().ReadNonceChange().Return(uint64Data, tracing.NonceChangeRevert, nil)
				state.EXPECT().SetNonce(addr, uint64Data, tracing.NonceChangeRevert)
			},
		},
		{
			name: "SetState",
			op:   tracer.Operation{Op: tracer.SetStateID, Addr: addr, Key: key, Value: value},
			setup: func(reader *tracer.MockFileReader, state *state.MockStateDB) {
				state.EXPECT().SetState(addr, key, value)
			},
		},
		{
			name: "Snapshot",
			op:   tracer.Operation{Op: tracer.SnapshotID},
			setup: func(reader *tracer.MockFileReader, state *state.MockStateDB) {
				state.EXPECT().Snapshot()
			},
		},
		{
			name: "SubBalance",
			op:   tracer.Operation{Op: tracer.SubBalanceID, Addr: addr},
			setup: func(reader *tracer.MockFileReader, state *state.MockStateDB) {
				balance := uint256.NewInt(50)
				reader.EXPECT().ReadBalanceChange().Return(balance, tracing.BalanceChangeRevert, nil)
				state.EXPECT().SubBalance(addr, balance, tracing.BalanceChangeRevert)
			},
		},
		{
			name: "SelfDestruct",
			op:   tracer.Operation{Op: tracer.SelfDestructID, Addr: addr},
			setup: func(reader *tracer.MockFileReader, state *state.MockStateDB) {
				state.EXPECT().SelfDestruct(addr)
			},
		},
		{
			name: "CreateContract",
			op:   tracer.Operation{Op: tracer.CreateContractID, Addr: addr},
			setup: func(reader *tracer.MockFileReader, state *state.MockStateDB) {
				state.EXPECT().CreateContract(addr)
			},
		},
		{
			name: "GetStorageRoot",
			op:   tracer.Operation{Op: tracer.GetStorageRootID, Addr: addr},
			setup: func(reader *tracer.MockFileReader, state *state.MockStateDB) {
				state.EXPECT().GetStorageRoot(addr)
			},
		},
		{
			name: "GetTransientState",
			op:   tracer.Operation{Op: tracer.GetTransientStateID, Addr: addr, Key: key},
			setup: func(reader *tracer.MockFileReader, state *state.MockStateDB) {
				state.EXPECT().GetTransientState(addr, key)
			},
		},
		{
			name: "SetTransientState",
			op:   tracer.Operation{Op: tracer.SetTransientStateID, Addr: addr, Value: value},
			setup: func(reader *tracer.MockFileReader, state *state.MockStateDB) {
				state.EXPECT().GetTransientState(addr, value)
			},
		},
		{
			name: "SelfDestruct6780",
			op:   tracer.Operation{Op: tracer.SelfDestruct6780ID, Addr: addr},
			setup: func(reader *tracer.MockFileReader, state *state.MockStateDB) {
				state.EXPECT().SelfDestruct6780(addr)
			},
		},
		{
			name: "SubRefund",
			op:   tracer.Operation{Op: tracer.SubRefundID},
			setup: func(reader *tracer.MockFileReader, state *state.MockStateDB) {
				reader.EXPECT().ReadUint64().Return(uint64Data, nil)
				state.EXPECT().SubRefund(uint64Data)
			},
		},
		{
			name: "GetRefund",
			op:   tracer.Operation{Op: tracer.GetRefundID},
			setup: func(reader *tracer.MockFileReader, state *state.MockStateDB) {
				state.EXPECT().GetRefund()
			},
		},
		{
			name: "AddRefund",
			op:   tracer.Operation{Op: tracer.AddRefundID},
			setup: func(reader *tracer.MockFileReader, state *state.MockStateDB) {
				reader.EXPECT().ReadUint64().Return(uint64Data, nil)
				state.EXPECT().AddRefund(uint64Data)
			},
		},
		{
			name: "Prepare",
			op:   tracer.Operation{Op: tracer.PrepareID},
			setup: func(reader *tracer.MockFileReader, state *state.MockStateDB) {
				rules := params.Rules{
					IsPrague: true,
				}
				accessList := types.AccessList{
					{
						Address:     addr,
						StorageKeys: []common.Hash{key},
					},
				}

				reader.EXPECT().ReadRules().Return(rules, nil)
				reader.EXPECT().ReadAddr().Return(addr, nil).Times(3)
				reader.EXPECT().ReadUint32().Return(uint32Data, nil)
				reader.EXPECT().ReadAddr().Return(addr, nil).Times(int(uint32Data))
				reader.EXPECT().ReadAccessList().Return(accessList, nil)
				state.EXPECT().Prepare(rules, addr, addr, &addr, gomock.Any(), accessList)
			},
		},
		{
			name: "AddAddressToAccessList",
			op:   tracer.Operation{Op: tracer.AddAddressToAccessListID, Addr: addr},
			setup: func(reader *tracer.MockFileReader, state *state.MockStateDB) {
				state.EXPECT().AddAddressToAccessList(addr)
			},
		},
		{
			name: "AddressInAccessList",
			op:   tracer.Operation{Op: tracer.AddressInAccessListID, Addr: addr},
			setup: func(reader *tracer.MockFileReader, state *state.MockStateDB) {
				state.EXPECT().AddressInAccessList(addr)
			},
		},
		{
			name: "SlotInAccessList",
			op:   tracer.Operation{Op: tracer.SlotInAccessListID, Addr: addr, Key: key},
			setup: func(reader *tracer.MockFileReader, state *state.MockStateDB) {
				state.EXPECT().SlotInAccessList(addr, key)
			},
		},
		{
			name: "AddSlotToAccessList",
			op:   tracer.Operation{Op: tracer.AddSlotToAccessListID, Addr: addr, Key: key},
			setup: func(reader *tracer.MockFileReader, state *state.MockStateDB) {
				state.EXPECT().AddSlotToAccessList(addr, key)
			},
		},
		{
			name: "AddLog",
			op:   tracer.Operation{Op: tracer.AddLogID},
			setup: func(reader *tracer.MockFileReader, state *state.MockStateDB) {
				log := &types.Log{
					Address: addr,
					Topics:  []common.Hash{key},
					Data:    code,
				}
				reader.EXPECT().ReadLog().Return(log, nil)
				state.EXPECT().AddLog(log)
			},
		},
		{
			name: "GetLogs",
			op:   tracer.Operation{Op: tracer.GetLogsID},
			setup: func(reader *tracer.MockFileReader, state *state.MockStateDB) {
				reader.EXPECT().ReadHash().Return(hash, nil)
				reader.EXPECT().ReadUint64().Return(uint64Data, nil)
				reader.EXPECT().ReadHash().Return(hash, nil)
				reader.EXPECT().ReadUint64().Return(uint64Data, nil)
				state.EXPECT().GetLogs(hash, uint64Data, hash, uint64Data)
			},
		},
		{
			name: "PointCache",
			op:   tracer.Operation{Op: tracer.PointCacheID},
			setup: func(reader *tracer.MockFileReader, state *state.MockStateDB) {
				state.EXPECT().PointCache()
			},
		},
		{
			name: "Witness",
			op:   tracer.Operation{Op: tracer.WitnessID},
			setup: func(reader *tracer.MockFileReader, state *state.MockStateDB) {
				state.EXPECT().Witness()
			},
		},
		{
			name: "AddPreimage",
			op:   tracer.Operation{Op: tracer.AddPreimageID},
			setup: func(reader *tracer.MockFileReader, state *state.MockStateDB) {
				reader.EXPECT().ReadHash().Return(hash, nil)
				reader.EXPECT().ReadVariableSizeData().Return(code, nil)
				state.EXPECT().AddPreimage(hash, code)
			},
		},
		{
			name: "SetTxContext",
			op:   tracer.Operation{Op: tracer.SetTxContextID},
			setup: func(reader *tracer.MockFileReader, state *state.MockStateDB) {
				reader.EXPECT().ReadHash().Return(hash, nil)
				reader.EXPECT().ReadUint32().Return(uint32Data, nil)
				state.EXPECT().SetTxContext(hash, int(uint32Data))
			},
		},
		{
			name: "Finalise",
			op:   tracer.Operation{Op: tracer.FinaliseID},
			setup: func(reader *tracer.MockFileReader, state *state.MockStateDB) {
				reader.EXPECT().ReadBool().Return(boolData, nil)
				state.EXPECT().Finalise(boolData)
			},
		},
		{
			name: "IntermediateRoot",
			op:   tracer.Operation{Op: tracer.IntermediateRootID},
			setup: func(reader *tracer.MockFileReader, state *state.MockStateDB) {
				reader.EXPECT().ReadBool().Return(boolData, nil)
				state.EXPECT().IntermediateRoot(boolData)
			},
		},
		{
			name: "Commit",
			op:   tracer.Operation{Op: tracer.CommitID},
			setup: func(reader *tracer.MockFileReader, state *state.MockStateDB) {
				reader.EXPECT().ReadBool().Return(boolData, nil)
				reader.EXPECT().ReadUint64().Return(uint64Data, nil)
				state.EXPECT().Commit(uint64Data, boolData).Return(common.Hash{0x23}, nil)
			},
		},
		{
			name: "Close",
			op:   tracer.Operation{Op: tracer.CloseID},
			setup: func(reader *tracer.MockFileReader, state *state.MockStateDB) {
				state.EXPECT().Close().Return(nil)
			},
		},
		{
			name: "AccessEvents",
			op:   tracer.Operation{Op: tracer.AccessEventsID},
			setup: func(reader *tracer.MockFileReader, state *state.MockStateDB) {
				state.EXPECT().AccessEvents()
			},
		},
		{
			name: "GetHash",
			op:   tracer.Operation{Op: tracer.GetHashID},
			setup: func(reader *tracer.MockFileReader, state *state.MockStateDB) {
				state.EXPECT().GetHash().Return(common.Hash{0x12}, nil)
			},
		},
		{
			name: "GetSubstatePostAlloc",
			op:   tracer.Operation{Op: tracer.GetSubstatePostAllocID},
			setup: func(reader *tracer.MockFileReader, state *state.MockStateDB) {
				state.EXPECT().GetSubstatePostAlloc()
			},
		},
		{
			name: "PrepareSubstate",
			op:   tracer.Operation{Op: tracer.PrepareSubstateID},
			setup: func(reader *tracer.MockFileReader, state *state.MockStateDB) {
				ws := txcontext.AidaWorldState{
					addr: txcontext.NewAccount(
						[]byte{0x22},
						map[common.Hash]common.Hash{{0x1}: {0x3}},
						big.NewInt(22),
						12,
					),
				}
				reader.EXPECT().ReadWorldState().Return(ws, nil)
				state.EXPECT().PrepareSubstate(ws, uint64Data)
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			fr := tracer.NewMockFileReader(ctrl)
			stateDb := state.NewMockStateDB(ctrl)
			tp := traceProcessor{fr}

			test.setup(fr, stateDb)

			err := tp.Process(executor.State[tracer.Operation]{
				Data:        test.op,
				Block:       int(uint64Data),
				Transaction: int(uint32Data),
			}, &executor.Context{State: stateDb})
			assert.NoError(t, err)
		})
	}
}
