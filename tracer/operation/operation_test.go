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

package operation

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetLabel(t *testing.T) {
	tests := []struct {
		name     string
		id       byte
		expected string
	}{
		{"AddBalance", AddBalanceID, "AddBalance"},
		{"BeginBlock", BeginBlockID, "BeginBlock"},
		{"BeginSyncPeriod", BeginSyncPeriodID, "BeginSyncPeriod"},
		{"BeginTransaction", BeginTransactionID, "BeginTransaction"},
		{"CreateAccount", CreateAccountID, "CreateAccount"},
		{"Commit", CommitID, "Commit"},
		{"Empty", EmptyID, "Empty"},
		{"EndBlock", EndBlockID, "EndBlock"},
		{"EndSyncPeriod", EndSyncPeriodID, "EndSyncPeriod"},
		{"EndTransaction", EndTransactionID, "EndTransaction"},
		{"Exist", ExistID, "Exist"},
		{"Finalise", FinaliseID, "Finalise"},
		{"GetBalance", GetBalanceID, "GetBalance"},
		{"GetCodeHash", GetCodeHashID, "GetCodeHash"},
		{"GetCodeHashLc", GetCodeHashLcID, "GetCodeLcHash"},
		{"GetCode", GetCodeID, "GetCode"},
		{"GetCodeSize", GetCodeSizeID, "GetCodeSize"},
		{"GetCommittedState", GetCommittedStateID, "GetCommittedState"},
		{"GetCommittedStateLcls", GetCommittedStateLclsID, "GetCommittedStateLcls"},
		{"GetStateAndCommittedState", GetStateAndCommittedStateID, "GetStateAndCommittedState"},
		{"GetStateAndCommittedStateLcls", GetStateAndCommittedStateLclsID, "GetStateAndCommittedStateLcls"},
		{"GetNonce", GetNonceID, "GetNonce"},
		{"GetState", GetStateID, "GetState"},
		{"GetStateLc", GetStateLcID, "GetStateLc"},
		{"GetStateLccs", GetStateLccsID, "GetStateLccs"},
		{"GetStateLcls", GetStateLclsID, "GetStateLcls"},
		{"HasSelfDestructed", HasSelfDestructedID, "HasSelfDestructed"},
		{"RevertToSnapshot", RevertToSnapshotID, "RevertToSnapshot"},
		{"SetCode", SetCodeID, "SetCode"},
		{"SetNonce", SetNonceID, "SetNonce"},
		{"SetState", SetStateID, "SetState"},
		{"SetStateLcls", SetStateLclsID, "SetStateLcls"},
		{"Snapshot", SnapshotID, "Snapshot"},
		{"SubBalance", SubBalanceID, "SubBalance"},
		{"SelfDestruct", SelfDestructID, "SelfDestruct"},
		{"AddAddressToAccessList", AddAddressToAccessListID, "AddAddressToAccessList"},
		{"AddressInAccessList", AddressInAccessListID, "AddressInAccessList"},
		{"AddSlotToAccessList", AddSlotToAccessListID, "AddSlotToAccessList"},
		{"Prepare", PrepareID, "Prepare"},
		{"SlotInAccessList", SlotInAccessListID, "SlotInAccessList"},
		{"AddLog", AddLogID, "AddLog"},
		{"AddPreimage", AddPreimageID, "AddPreimage"},
		{"AddRefund", AddRefundID, "AddRefund"},
		{"Close", CloseID, "Close"},
		{"GetLogs", GetLogsID, "GetLogs"},
		{"GetRefund", GetRefundID, "GetRefund"},
		{"IntermediateRoot", IntermediateRootID, "IntermediateRoot"},
		{"SetTxContext", SetTxContextID, "SetTxContext"},
		{"SubRefund", SubRefundID, "SubRefund"},
		{"CreateContract", CreateContractID, "CreateContract"},
		{"GetStorageRoot", GetStorageRootID, "GetStorageRoot"},
		{"GetTransientState", GetTransientStateID, "GetTransientState"},
		{"GetTransientStateLc", GetTransientStateLcID, "GetTransientStateLc"},
		{"GetTransientStateLccs", GetTransientStateLccsID, "GetTransientStateLccs"},
		{"GetTransientStateLcls", GetTransientStateLclsID, "GetTransientStateLcls"},
		{"SetTransientState", SetTransientStateID, "SetTransientState"},
		{"SetTransientStateLcls", SetTransientStateLclsID, "SetTransientStateLcls"},
		{"SelfDestruct6780", SelfDestruct6780ID, "SelfDestruct"},
		{"PointCache", PointCacheID, "PointCache"},
		{"Witness", WitnessID, "Witness"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetLabel(tt.id)
			assert.Equal(t, tt.expected, result, "GetLabel(%d) should return %s", tt.id, tt.expected)
		})
	}
}

func TestCreateIdLabelMap(t *testing.T) {
	idLabelMap := CreateIdLabelMap()

	require.NotNil(t, idLabelMap)
	assert.Equal(t, int(NumOperations), len(idLabelMap))

	// Verify some mappings
	testCases := []struct {
		id       byte
		expected string
	}{
		{AddBalanceID, "AddBalance"},
		{BeginBlockID, "BeginBlock"},
		{CommitID, "Commit"},
		{GetStateID, "GetState"},
		{SetStateID, "SetState"},
		{SnapshotID, "Snapshot"},
		{AddLogID, "AddLog"},
		{GetTransientStateID, "GetTransientState"},
		{SetTransientStateID, "SetTransientState"},
		{SelfDestruct6780ID, "SelfDestruct"},
		{PointCacheID, "PointCache"},
		{WitnessID, "Witness"},
	}

	for _, tc := range testCases {
		label, exists := idLabelMap[tc.id]
		assert.True(t, exists)
		assert.Equal(t, tc.expected, label)
	}
}
