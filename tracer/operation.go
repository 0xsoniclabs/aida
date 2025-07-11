// Copyright 2024 Fantom Foundation
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

package tracer

import (
	"fmt"
	"log"

	"github.com/0xsoniclabs/aida/stochastic/statistics"
)

// IDs of StateDB Operations
const (
	AddBalanceID = iota
	BeginBlockID
	BeginSyncPeriodID
	BeginTransactionID
	CreateAccountID
	EmptyID
	EndBlockID
	EndSyncPeriodID
	EndTransactionID
	ExistID
	GetBalanceID
	GetCodeHashID
	GetCodeID
	GetCodeSizeID
	GetCommittedStateID
	GetNonceID
	GetStateID
	HasSelfDestructedID
	RevertToSnapshotID
	SetCodeID
	SetNonceID
	SetStateID
	SnapshotID
	SubBalanceID
	SelfDestructID
	CreateContractID
	GetStorageRootID
	GetTransientStateID
	SetTransientStateID
	SelfDestruct6780ID
	SubRefundID
	GetRefundID
	// Add new operations below this line

	NumOps
)

// numArgOps gives the number of operations with encoded argument classes
const numArgOps = NumOps * statistics.NumClasses * statistics.NumClasses * statistics.NumClasses

// opText translates IDs to operation's text
var opText = map[int]string{
	AddBalanceID:        "AddBalance",
	BeginBlockID:        "BeginBlock",
	BeginSyncPeriodID:   "BeginSyncPeriod",
	BeginTransactionID:  "BeginTransaction",
	CreateAccountID:     "CreateAccount",
	CreateContractID:    "CreateContract",
	EmptyID:             "Empty",
	EndBlockID:          "EndBlock",
	EndSyncPeriodID:     "EndSyncPeriod",
	EndTransactionID:    "EndTransaction",
	ExistID:             "Exist",
	GetBalanceID:        "GetBalance",
	GetCodeHashID:       "GetCodeHash",
	GetCodeID:           "GetCode",
	GetCodeSizeID:       "GetCodeSize",
	GetCommittedStateID: "GetCommittedState",
	GetNonceID:          "GetNonce",
	GetStateID:          "GetState",
	GetStorageRootID:    "GetStorageRoot",
	GetTransientStateID: "GetTransientState",
	HasSelfDestructedID: "HasSelfDestructed",
	RevertToSnapshotID:  "RevertToSnapshot",
	SelfDestructID:      "SelfDestruct",
	SelfDestruct6780ID:  "SelfDestruct6780",
	SetCodeID:           "SetCode",
	SetNonceID:          "SetNonce",
	SetStateID:          "SetState",
	SnapshotID:          "Snapshot",
	SubBalanceID:        "SubBalance",
	SetTransientStateID: "SetTransientState",
}

// opMnemo is a mnemonics table for operations.
var opMnemo = map[uint16]string{
	AddBalanceID:        "AB",
	BeginBlockID:        "BB",
	BeginSyncPeriodID:   "BS",
	BeginTransactionID:  "BT",
	CreateAccountID:     "CA",
	CreateContractID:    "CC",
	EmptyID:             "EM",
	EndBlockID:          "EB",
	EndSyncPeriodID:     "ES",
	EndTransactionID:    "ET",
	ExistID:             "EX",
	GetBalanceID:        "GB",
	GetCodeHashID:       "GH",
	GetCodeID:           "GC",
	GetCodeSizeID:       "GZ",
	GetCommittedStateID: "GM",
	GetNonceID:          "GN",
	GetStateID:          "GS",
	GetStorageRootID:    "GR",
	GetTransientStateID: "GT",
	HasSelfDestructedID: "HS",
	RevertToSnapshotID:  "RS",
	SelfDestructID:      "SU",
	SelfDestruct6780ID:  "S6",
	SetCodeID:           "SC",
	SetNonceID:          "SO",
	SetStateID:          "SS",
	SnapshotID:          "SN",
	SubBalanceID:        "SB",
	SetTransientStateID: "ST",
}

// opNumArgs is an argument number table for operations.
var opNumArgs = map[uint16]int{
	AddBalanceID:        1,
	BeginBlockID:        0,
	BeginSyncPeriodID:   0,
	BeginTransactionID:  0,
	CreateAccountID:     1,
	CreateContractID:    1,
	EmptyID:             1,
	EndBlockID:          0,
	EndSyncPeriodID:     0,
	EndTransactionID:    0,
	ExistID:             1,
	GetBalanceID:        1,
	GetCodeHashID:       1,
	GetCodeID:           1,
	GetCodeSizeID:       1,
	GetCommittedStateID: 2,
	GetNonceID:          1,
	GetStateID:          2,
	GetStorageRootID:    1,
	GetTransientStateID: 2,
	HasSelfDestructedID: 1,
	RevertToSnapshotID:  0,
	SelfDestructID:      1,
	SelfDestruct6780ID:  1,
	SetCodeID:           1,
	SetNonceID:          1,
	SetStateID:          3,
	SnapshotID:          0,
	SubBalanceID:        1,
	SetTransientStateID: 3,
}

// opId is an operation ID table.
var opId = map[string]uint16{
	"AB": AddBalanceID,
	"BB": BeginBlockID,
	"BS": BeginSyncPeriodID,
	"BT": BeginTransactionID,
	"CA": CreateAccountID,
	"CC": CreateContractID,
	"EM": EmptyID,
	"EB": EndBlockID,
	"ES": EndSyncPeriodID,
	"ET": EndTransactionID,
	"EX": ExistID,
	"GB": GetBalanceID,
	"GH": GetCodeHashID,
	"GC": GetCodeID,
	"GZ": GetCodeSizeID,
	"GM": GetCommittedStateID,
	"GN": GetNonceID,
	"GS": GetStateID,
	"GR": GetStorageRootID,
	"GT": GetTransientStateID,
	"HS": HasSelfDestructedID,
	"RS": RevertToSnapshotID,
	"SU": SelfDestructID,
	"S6": SelfDestruct6780ID,
	"SC": SetCodeID,
	"SO": SetNonceID,
	"SN": SnapshotID,
	"SB": SubBalanceID,
	"SS": SetStateID,
	"ST": SetTransientStateID,
}

// argMnemo is the argument-class mnemonics table.
var argMnemo = map[uint8]string{
	NoArgID:         "",
	ZeroValueID:     "z",
	NewValueID:      "n",
	PreviousValueID: "p",
	RecentValueID:   "q",
}

// argId is the argument-class id table.
var argId = map[byte]uint8{
	'z': ZeroValueID,
	'n': NewValueID,
	'p': PreviousValueID,
	'q': RecentValueID,
}

// OpMnemo returns the mnemonic code for an operation.
func OpMnemo(op uint16) string {
	if op < 0 || op >= NumOps {
		panic("opcode is out of range")
	}
	return opMnemo[op]
}

// checkArgOp checks whether op/argument combination is valid.
func checkArgOp(op uint16, contract uint8, key uint8, value uint8) bool {
	if op < 0 || op >= NumOps {
		return false
	}
	if contract < 0 || contract >= NumClasses {
		return false
	}
	if key < 0 || key >= NumClasses {
		return false
	}
	if value < 0 || value >= NumClasses {
		return false
	}
	switch opNumArgs[op] {
	case 0:
		return contract == NoArgID &&
			key == NoArgID &&
			value == NoArgID
	case 1:
		return contract != NoArgID &&
			key == NoArgID &&
			value == NoArgID
	case 2:
		return contract != NoArgID &&
			key != NoArgID &&
			value == NoArgID
	case 3:
		return contract != NoArgID &&
			key != NoArgID &&
			value != NoArgID
	default:
		return false
	}
}

// IsValidArgOp returns true if the encoding is valid.
func IsValidArgOp(argop uint16) bool {
	if argop < 0 || argop >= numArgOps {
		return false
	}
	op, contract, key, value := DecodeArgOp(argop)
	return checkArgOp(op, contract, key, value)
}

// EncodeArgOp encodes operation and argument classes via Horner's scheme to a single value.
func EncodeArgOp(op uint16, addr uint8, key uint8, value uint8) (uint16, error) {
	if !checkArgOp(op, addr, key, value) {
		return 0, fmt.Errorf("EncodeArgOp: invalid operation/arguments\naddr: %d, key: %d, value: %d, op: %d", addr, key, value, op)
	}
	return (((op*uint16(NumClasses))+uint16(addr))*uint16(NumClasses)+uint16(key))*uint16(NumClasses) + uint16(value), nil
}

// DecodeArgOp decodes operation with arguments using Honer's scheme
func DecodeArgOp(argop uint16) (uint16, uint8, uint8, uint8) {
	if argop < 0 || argop >= numArgOps {
		log.Fatalf("DecodeArgOp: invalid op range")
	}

	value := argop % uint16(NumClasses)
	argop = argop / uint16(NumClasses)

	key := argop % uint16(NumClasses)
	argop = argop / uint16(NumClasses)

	addr := argop % uint16(NumClasses)
	argop = argop / uint16(NumClasses)

	op := argop

	return op, uint8(addr), uint8(key), uint8(value)
}

// EncodeOpcode generates the opcode for an operation and its argument classes.
func EncodeOpcode(op uint16, addr uint8, key uint8, value uint8) string {
	if !checkArgOp(op, addr, key, value) {
		log.Fatalf("EncodeOpcode: invalid operation/arguments")
	}
	code := fmt.Sprintf("%v%v%v%v", opMnemo[op], argMnemo[addr], argMnemo[key], argMnemo[value])
	if len(code) != 2+opNumArgs[op] {
		log.Fatalf("EncodeOpcode: wrong opcode length for opcode %v", code)
	}
	return code
}

// validateArg checks whether argument mnemonics exists.
func validateArg(argMnemo byte) bool {
	_, ok := argId[argMnemo]
	return ok
}

// DecodeOpcode decodes opcode producing the operation id and its argument classes
func DecodeOpcode(opc string) (uint16, uint8, uint8, uint8, error) {
	mnemo := opc[:2]
	op, ok := opId[mnemo]
	if !ok {
		return 0, 0, 0, 0, fmt.Errorf("DecodeOpcode: lookup failed for %v", mnemo)
	}
	if len(opc) != 2+opNumArgs[op] {
		return 0, 0, 0, 0, fmt.Errorf("DecodeOpcode: wrong opcode length for %v", opc)
	}
	var contract, key, value uint8
	switch len(opc) - 2 {
	case 0:
		contract, key, value = NoArgID, NoArgID, NoArgID
	case 1:
		if !validateArg(opc[2]) {
			return 0, 0, 0, 0, fmt.Errorf("DecodeOpcode: wrong argument code")
		}
		contract, key, value = argId[opc[2]], NoArgID, NoArgID
	case 2:
		if !validateArg(opc[2]) || !validateArg(opc[3]) {
			return 0, 0, 0, 0, fmt.Errorf("DecodeOpcode: wrong argument code")
		}
		contract, key, value = argId[opc[2]], argId[opc[3]], NoArgID
	case 3:
		if !validateArg(opc[2]) || !validateArg(opc[3]) || !validateArg(opc[4]) {
			return 0, 0, 0, 0, fmt.Errorf("DecodeOpcode: wrong argument code")
		}
		contract, key, value = argId[opc[2]], argId[opc[3]], argId[opc[4]]
	}
	return op, contract, key, value, nil
}
