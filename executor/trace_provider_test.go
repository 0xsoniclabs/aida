package executor

import (
	"github.com/0xsoniclabs/aida/tracer"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"testing"
)

func TestTraceProvider_Run(t *testing.T) {
	ctrl := gomock.NewController(t)
	fr := tracer.NewMockFileReader(ctrl)
	consumer := NewMockOperationConsumer(ctrl)
	p := &traceProvider{
		file: fr,
	}

	beginBlockEncoded, err := tracer.EncodeArgOp(tracer.BeginBlockID, tracer.NoArgID, tracer.NoArgID, tracer.NoArgID)
	require.NoError(t, err)
	beginTransactionEncoded, err := tracer.EncodeArgOp(tracer.BeginTransactionID, tracer.NoArgID, tracer.NoArgID, tracer.NoArgID)
	endTransactionEncoded, err := tracer.EncodeArgOp(tracer.EndTransactionID, tracer.NoArgID, tracer.NoArgID, tracer.NoArgID)
	require.NoError(t, err)

	gomock.InOrder(
		// Block 4 is getting skipped
		fr.EXPECT().ReadUint16().Return(beginBlockEncoded, nil),
		fr.EXPECT().ReadUint64().Return(uint64(4), nil),
		// Block 5 is the first one in the block range
		fr.EXPECT().ReadUint16().Return(beginBlockEncoded, nil),
		fr.EXPECT().ReadUint64().Return(uint64(5), nil),

		consumer.EXPECT().Consume(5, 0, gomock.Any()),
		fr.EXPECT().ReadUint16().Return(beginTransactionEncoded, nil),
		fr.EXPECT().ReadUint32().Return(uint32(3), nil),
		consumer.EXPECT().Consume(5, 3, gomock.Any()),

		fr.EXPECT().ReadUint16().Return(endTransactionEncoded, nil),
		consumer.EXPECT().Consume(5, 3, gomock.Any()),
		// Block 6
		fr.EXPECT().ReadUint16().Return(beginBlockEncoded, nil),
		fr.EXPECT().ReadUint64().Return(uint64(6), nil),
		consumer.EXPECT().Consume(6, 0, gomock.Any()),
		// Block 7 is out of range - bailout
		fr.EXPECT().ReadUint16().Return(beginBlockEncoded, nil),
		fr.EXPECT().ReadUint64().Return(uint64(7), nil),
		fr.EXPECT().Close().Return(nil),
	)

	err = p.Run(5, 6, toOperationConsumer(consumer))
	require.NoError(t, err)
}

func TestTraceProvider_readOperation(t *testing.T) {
	type operation struct {
		op                   uint8
		addrCl, keyCl, valCl uint8
	}
	addr1 := common.Address{0x1}
	addr2 := common.Address{0x2}
	key1 := common.Hash{0x1}
	key2 := common.Hash{0x2}
	val1 := common.Hash{0x4}
	val2 := common.Hash{0x5}
	tests := []struct {
		name       string
		operations []operation
		expectedOp tracer.Operation
		setup      func(*tracer.MockFileReader)
	}{
		{
			name: "NoArgs",
			operations: []operation{
				{op: tracer.BeginBlockID, addrCl: tracer.NoArgID, keyCl: tracer.NoArgID, valCl: tracer.NoArgID},
			},
			expectedOp: tracer.Operation{
				Op:    tracer.BeginBlockID,
				Addr:  common.Address{},
				Key:   common.Hash{},
				Value: common.Hash{},
			},
			setup: func(m *tracer.MockFileReader) {
				// nothing
			},
		},
		{
			name: "NewArgs",
			operations: []operation{
				// only one operation with new arguments
				{op: tracer.SetStateID, addrCl: tracer.NewValueID, keyCl: tracer.NewValueID, valCl: tracer.NewValueID},
			},
			expectedOp: tracer.Operation{
				Op:    tracer.SetStateID,
				Addr:  addr1,
				Key:   key1,
				Value: val1,
			},
			setup: func(m *tracer.MockFileReader) {

				gomock.InOrder(
					// Three new values
					m.EXPECT().ReadAddr().Return(addr1, nil),
					m.EXPECT().ReadHash().Return(key1, nil),
					m.EXPECT().ReadHash().Return(val1, nil),
				)
			},
		},
		{
			name: "PrevArgs",
			operations: []operation{
				// two operations, one with new arguments, another with previous arguments
				{op: tracer.SetStateID, addrCl: tracer.NewValueID, keyCl: tracer.NewValueID, valCl: tracer.NewValueID},
				{op: tracer.SetStateID, addrCl: tracer.PreviousValueID, keyCl: tracer.PreviousValueID, valCl: tracer.PreviousValueID},
			},
			expectedOp: tracer.Operation{
				Op:    tracer.SetStateID,
				Addr:  addr2,
				Key:   key2,
				Value: val2,
			},
			setup: func(m *tracer.MockFileReader) {
				gomock.InOrder(
					// Three new values
					m.EXPECT().ReadAddr().Return(addr2, nil),
					m.EXPECT().ReadHash().Return(key2, nil),
					m.EXPECT().ReadHash().Return(val2, nil),
					// Previous values do not touch the file
				)
			},
		},
		{
			name: "RecArgs",
			operations: []operation{
				// three operations, two with new arguments, another with recent arguments
				{op: tracer.SetStateID, addrCl: tracer.NewValueID, keyCl: tracer.NewValueID, valCl: tracer.NewValueID},
				{op: tracer.SetStateID, addrCl: tracer.NewValueID, keyCl: tracer.NewValueID, valCl: tracer.NewValueID},
				{op: tracer.SetStateID, addrCl: tracer.RecentValueID, keyCl: tracer.RecentValueID, valCl: tracer.RecentValueID},
			},
			expectedOp: tracer.Operation{
				Op:    tracer.SetStateID,
				Addr:  addr1,
				Key:   key1,
				Value: val1,
			},
			setup: func(m *tracer.MockFileReader) {
				gomock.InOrder(
					m.EXPECT().ReadAddr().Return(addr1, nil),
					m.EXPECT().ReadHash().Return(key1, nil),
					m.EXPECT().ReadHash().Return(val1, nil),
					m.EXPECT().ReadAddr().Return(addr2, nil),
					m.EXPECT().ReadHash().Return(key2, nil),
					m.EXPECT().ReadHash().Return(val2, nil),
					// Lastly indexes for the addresses, keys, and values are read
					m.EXPECT().ReadUint8().Return(uint8(1), nil), // addr1
					m.EXPECT().ReadUint8().Return(uint8(1), nil), // key1
					m.EXPECT().ReadUint8().Return(uint8(1), nil), // val1
				)
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			fr := tracer.NewMockFileReader(ctrl)
			p := &traceProvider{
				file:      fr,
				contracts: tracer.NewQueue[common.Address](),
				keys:      tracer.NewQueue[common.Hash](),
				values:    tracer.NewQueue[common.Hash](),
			}
			test.setup(fr)
			var (
				finalOp tracer.Operation
			)
			for _, op := range test.operations {
				argOp, err := tracer.EncodeArgOp(op.op, op.addrCl, op.keyCl, op.valCl)
				require.NoError(t, err)
				finalOp, err = p.readOperation(argOp)
				require.NoError(t, err)
			}
			require.Equal(t, test.expectedOp, finalOp)
		})
	}
}

func TestTraceProvider_readOperation_IncorrectOperation(t *testing.T) {
	ctrl := gomock.NewController(t)
	fr := tracer.NewMockFileReader(ctrl)
	p := traceProvider{file: fr}

	incorrect := uint16(tracer.NumOps) * uint16(tracer.NumClasses) * uint16(tracer.NumClasses) * uint16(tracer.NumClasses)
	_, err := p.readOperation(incorrect)
	require.ErrorContains(t, err, "cannot decode operation")
}
