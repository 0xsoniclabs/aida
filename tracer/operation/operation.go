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
	"fmt"
	"io"
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
	label    string                             // operation's Label
	readfunc func(io.Reader) (Operation, error) // operation's read-function
}

// opDict relates an operation's id with its label and read-function.
var opDict = map[byte]OperationDictionary{
	AddBalanceID:                    {label: "AddBalance", readfunc: ReadPanic},
	BeginBlockID:                    {label: "BeginBlock", readfunc: ReadPanic},
	BeginSyncPeriodID:               {label: "BeginSyncPeriod", readfunc: ReadPanic},
	BeginTransactionID:              {label: "BeginTransaction", readfunc: ReadPanic},
	CommitID:                        {label: "Commit", readfunc: ReadPanic},
	CreateAccountID:                 {label: "CreateAccount", readfunc: ReadPanic},
	EmptyID:                         {label: "Empty", readfunc: ReadPanic},
	EndBlockID:                      {label: "EndBlock", readfunc: ReadPanic},
	EndSyncPeriodID:                 {label: "EndSyncPeriod", readfunc: ReadPanic},
	EndTransactionID:                {label: "EndTransaction", readfunc: ReadPanic},
	ExistID:                         {label: "Exist", readfunc: ReadPanic},
	FinaliseID:                      {label: "Finalise", readfunc: ReadPanic},
	GetBalanceID:                    {label: "GetBalance", readfunc: ReadPanic},
	GetCodeHashID:                   {label: "GetCodeHash", readfunc: ReadPanic},
	GetCodeHashLcID:                 {label: "GetCodeLcHash", readfunc: ReadPanic},
	GetCodeID:                       {label: "GetCode", readfunc: ReadPanic},
	GetCodeSizeID:                   {label: "GetCodeSize", readfunc: ReadPanic},
	GetCommittedStateID:             {label: "GetCommittedState", readfunc: ReadPanic},
	GetCommittedStateLclsID:         {label: "GetCommittedStateLcls", readfunc: ReadPanic},
	GetStateAndCommittedStateID:     {label: "GetStateAndCommittedState", readfunc: ReadPanic},
	GetStateAndCommittedStateLclsID: {label: "GetStateAndCommittedStateLcls", readfunc: ReadPanic},
	GetNonceID:                      {label: "GetNonce", readfunc: ReadPanic},
	GetStateID:                      {label: "GetState", readfunc: ReadPanic},
	GetStateLcID:                    {label: "GetStateLc", readfunc: ReadPanic},
	GetStateLccsID:                  {label: "GetStateLccs", readfunc: ReadPanic},
	GetStateLclsID:                  {label: "GetStateLcls", readfunc: ReadPanic},
	HasSelfDestructedID:             {label: "HasSelfDestructed", readfunc: ReadPanic},
	RevertToSnapshotID:              {label: "RevertToSnapshot", readfunc: ReadPanic},
	SetCodeID:                       {label: "SetCode", readfunc: ReadPanic},
	SetNonceID:                      {label: "SetNonce", readfunc: ReadPanic},
	SetStateID:                      {label: "SetState", readfunc: ReadPanic},
	SetStateLclsID:                  {label: "SetStateLcls", readfunc: ReadPanic},
	SnapshotID:                      {label: "Snapshot", readfunc: ReadPanic},
	SubBalanceID:                    {label: "SubBalance", readfunc: ReadPanic},
	SelfDestructID:                  {label: "SelfDestruct", readfunc: ReadPanic},
	SelfDestruct6780ID:              {label: "SelfDestruct", readfunc: ReadPanic},
	CreateContractID:                {label: "CreateContract", readfunc: ReadPanic},
	GetStorageRootID:                {label: "GetStorageRoot", readfunc: ReadPanic},

	// for testing
	AddAddressToAccessListID: {label: "AddAddressToAccessList", readfunc: ReadPanic},
	AddLogID:                 {label: "AddLog", readfunc: ReadPanic},
	AddPreimageID:            {label: "AddPreimage", readfunc: ReadPanic},
	AddRefundID:              {label: "AddRefund", readfunc: ReadPanic},
	AddressInAccessListID:    {label: "AddressInAccessList", readfunc: ReadPanic},
	AddSlotToAccessListID:    {label: "AddSlotToAccessList", readfunc: ReadPanic},
	CloseID:                  {label: "Close", readfunc: ReadPanic},
	GetLogsID:                {label: "GetLogs", readfunc: ReadPanic},
	GetRefundID:              {label: "GetRefund", readfunc: ReadPanic},
	IntermediateRootID:       {label: "IntermediateRoot", readfunc: ReadPanic},
	PrepareID:                {label: "Prepare", readfunc: ReadPanic},
	SetTxContextID:           {label: "SetTxContext", readfunc: ReadPanic},
	SlotInAccessListID:       {label: "SlotInAccessList", readfunc: ReadPanic},
	SubRefundID:              {label: "SubRefund", readfunc: ReadPanic},
	PointCacheID:             {label: "PointCache", readfunc: ReadPanic},
	WitnessID:                {label: "Witness", readfunc: ReadPanic},

	// Transient Storage
	GetTransientStateID:     {label: "GetTransientState", readfunc: ReadPanic},
	GetTransientStateLcID:   {label: "GetTransientStateLc", readfunc: ReadPanic},
	GetTransientStateLccsID: {label: "GetTransientStateLccs", readfunc: ReadPanic},
	GetTransientStateLclsID: {label: "GetTransientStateLcls", readfunc: ReadPanic},
	SetTransientStateID:     {label: "SetTransientState", readfunc: ReadPanic},
	SetTransientStateLclsID: {label: "SetTransientStateLcls", readfunc: ReadPanic},
}

// GetLabel retrieves a label of a state operation.
func GetLabel(i byte) string {
	if _, ok := opDict[i]; !ok {
		log.Fatalf("GetLabel failed; operation is not defined")
	}

	return opDict[i].label
}

// Operation interface.
type Operation interface {
	GetId() byte           // get operation identifier
	Write(io.Writer) error // write operation to a file
}

func ReadPanic(f io.Reader) (Operation, error) {
	return nil, fmt.Errorf("operation not implemented")
}

// CreateIdLabelMap returns a map of opcode ID and opcode name
func CreateIdLabelMap() map[byte]string {
	ret := make(map[byte]string)
	for id := byte(0); id < NumOperations; id++ {
		ret[id] = GetLabel(id)
	}
	return ret
}
