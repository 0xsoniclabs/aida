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

// Operation IDs of the StateDB interface
const (
	AddBalanceID = iota
	BeginBlockID
	BeginSyncPeriodID
	BeginTransactionID
	CreateAccountID
	CommitID
	EmptyID
	EndBlockID
	EndSyncPeriodID
	EndTransactionID
	ExistID
	FinaliseID
	GetBalanceID
	GetCodeHashID
	GetCodeHashLcID
	GetCodeID
	GetCodeSizeID
	GetCommittedStateID
	GetCommittedStateLclsID
	GetStateAndCommittedStateID
	GetStateAndCommittedStateLclsID
	GetNonceID
	GetStateID
	GetStateLccsID
	GetStateLcID
	GetStateLclsID
	HasSelfDestructedID
	RevertToSnapshotID
	SetCodeID
	SetNonceID
	SetStateID
	SetStateLclsID
	SnapshotID
	SubBalanceID
	SelfDestructID

	AddAddressToAccessListID
	AddressInAccessListID
	AddSlotToAccessListID
	PrepareID
	SlotInAccessListID

	AddLogID
	AddPreimageID
	AddRefundID
	CloseID
	GetLogsID
	GetRefundID
	IntermediateRootID
	SetTxContextID
	SubRefundID

	// statedb operatioans from Altair to Cancun
	CreateContractID
	GetStorageRootID
	GetTransientStateID
	GetTransientStateLccsID
	GetTransientStateLcID
	GetTransientStateLclsID
	SetTransientStateID
	SetTransientStateLclsID
	SelfDestruct6780ID
	PointCacheID
	WitnessID

	// WARNING: New IDs should be added here. Any change in the order of the
	// IDs above invalidates persisted data -- in particular storage traces.

	// NumOperations is number of distinct operations (must be last)
	NumOperations
)


// Operation interface.
type Operation interface {
	GetId() byte                                                   // get operation identifier
}



