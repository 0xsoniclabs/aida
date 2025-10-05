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
