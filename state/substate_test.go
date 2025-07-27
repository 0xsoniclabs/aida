package state

import (
	"testing"

	"github.com/0xsoniclabs/aida/txcontext"
	"github.com/ethereum/go-ethereum/params"
	"github.com/stretchr/testify/require"
)

func TestOffTheChainStateDb_CloseDoesNotPanicIfBackendIsNil(t *testing.T) {
	conduit := NewChainConduit(true, params.AllEthashProtocolChanges)
	db, err := MakeOffTheChainStateDB(txcontext.AidaWorldState{}, 0, conduit)
	require.NoError(t, err)
	err = db.Close()
	require.NoError(t, err)
}
