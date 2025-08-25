package generate

import (
	"testing"

	"github.com/0xsoniclabs/aida/state/proxy"
	"github.com/0xsoniclabs/aida/utils"
	"github.com/0xsoniclabs/substate/db"
	"github.com/0xsoniclabs/substate/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"
)

func TestDeletedAccounts_writeDeletedAccounts(t *testing.T) {
	ddbPath := t.TempDir() + "/ddb"
	ddb, err := db.NewDefaultDestroyedAccountDB(ddbPath)
	require.NoError(t, err)
	ss := utils.GetTestSubstate("pb")
	ch := make(chan proxy.ContractLiveliness, 3)

	// delete 2 accounts
	ch <- proxy.ContractLiveliness{
		Addr:      common.Address{0x1},
		IsDeleted: true,
	}
	ch <- proxy.ContractLiveliness{
		Addr:      common.Address{0x2},
		IsDeleted: true,
	}
	// resurrect one of them
	ch <- proxy.ContractLiveliness{
		Addr:      common.Address{0x1},
		IsDeleted: false,
	}

	close(ch)
	deleteHistory := make(map[common.Address]bool)
	err = writeDeletedAccounts(ddb, ss, ch, &deleteHistory)
	require.NoError(t, err)
	destroyed, resurrected, err := ddb.GetDestroyedAccounts(ss.Block, ss.Transaction)
	require.NoError(t, err)
	require.Contains(t, resurrected, types.Address{0x1})
	require.Contains(t, destroyed, types.Address{0x2})
}
