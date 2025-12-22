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
	"log"
)

//go:generate mockgen -source operation.go -destination operation_mock.go -package operation

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

// OperationDictionary data structure contains a Label and a read function for an operation
type OperationDictionary struct {
	label string // operation's Label
}

// opDict relates an operation's id with its label and read-function.
var opDict = map[byte]OperationDictionary{
	AddBalanceID:                    {label: "AddBalance"},
	BeginBlockID:                    {label: "BeginBlock"},
	BeginSyncPeriodID:               {label: "BeginSyncPeriod"},
	BeginTransactionID:              {label: "BeginTransaction"},
	CommitID:                        {label: "Commit"},
	CreateAccountID:                 {label: "CreateAccount"},
	EmptyID:                         {label: "Empty"},
	EndBlockID:                      {label: "EndBlock"},
	EndSyncPeriodID:                 {label: "EndSyncPeriod"},
	EndTransactionID:                {label: "EndTransaction"},
	ExistID:                         {label: "Exist"},
	FinaliseID:                      {label: "Finalise"},
	GetBalanceID:                    {label: "GetBalance"},
	GetCodeHashID:                   {label: "GetCodeHash"},
	GetCodeHashLcID:                 {label: "GetCodeLcHash"},
	GetCodeID:                       {label: "GetCode"},
	GetCodeSizeID:                   {label: "GetCodeSize"},
	GetCommittedStateID:             {label: "GetCommittedState"},
	GetCommittedStateLclsID:         {label: "GetCommittedStateLcls"},
	GetStateAndCommittedStateID:     {label: "GetStateAndCommittedState"},
	GetStateAndCommittedStateLclsID: {label: "GetStateAndCommittedStateLcls"},
	GetNonceID:                      {label: "GetNonce"},
	GetStateID:                      {label: "GetState"},
	GetStateLcID:                    {label: "GetStateLc"},
	GetStateLccsID:                  {label: "GetStateLccs"},
	GetStateLclsID:                  {label: "GetStateLcls"},
	HasSelfDestructedID:             {label: "HasSelfDestructed"},
	RevertToSnapshotID:              {label: "RevertToSnapshot"},
	SetCodeID:                       {label: "SetCode"},
	SetNonceID:                      {label: "SetNonce"},
	SetStateID:                      {label: "SetState"},
	SetStateLclsID:                  {label: "SetStateLcls"},
	SnapshotID:                      {label: "Snapshot"},
	SubBalanceID:                    {label: "SubBalance"},
	SelfDestructID:                  {label: "SelfDestruct"},
	SelfDestruct6780ID:              {label: "SelfDestruct"},
	CreateContractID:                {label: "CreateContract"},
	GetStorageRootID:                {label: "GetStorageRoot"},

	// for testing
	AddAddressToAccessListID: {label: "AddAddressToAccessList"},
	AddLogID:                 {label: "AddLog"},
	AddPreimageID:            {label: "AddPreimage"},
	AddRefundID:              {label: "AddRefund"},
	AddressInAccessListID:    {label: "AddressInAccessList"},
	AddSlotToAccessListID:    {label: "AddSlotToAccessList"},
	CloseID:                  {label: "Close"},
	GetLogsID:                {label: "GetLogs"},
	GetRefundID:              {label: "GetRefund"},
	IntermediateRootID:       {label: "IntermediateRoot"},
	PrepareID:                {label: "Prepare"},
	SetTxContextID:           {label: "SetTxContext"},
	SlotInAccessListID:       {label: "SlotInAccessList"},
	SubRefundID:              {label: "SubRefund"},
	PointCacheID:             {label: "PointCache"},
	WitnessID:                {label: "Witness"},

	// Transient Storage
	GetTransientStateID:     {label: "GetTransientState"},
	GetTransientStateLcID:   {label: "GetTransientStateLc"},
	GetTransientStateLccsID: {label: "GetTransientStateLccs"},
	GetTransientStateLclsID: {label: "GetTransientStateLcls"},
	SetTransientStateID:     {label: "SetTransientState"},
	SetTransientStateLclsID: {label: "SetTransientStateLcls"},
}

// GetLabel retrieves a label of a state operation.
func GetLabel(i byte) string {
	if _, ok := opDict[i]; !ok {
		log.Fatalf("GetLabel failed; operation is not defined")
	}

	return opDict[i].label
}

// CreateIdLabelMap returns a map of opcode ID and opcode name
func CreateIdLabelMap() map[byte]string {
	ret := make(map[byte]string)
	for id := byte(0); id < NumOperations; id++ {
		ret[id] = GetLabel(id)
	}
	return ret
}
