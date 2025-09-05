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

package operations

import (
	"encoding/binary"
	"fmt"
	"log"

	"github.com/0xsoniclabs/aida/stochastic/statistics/classifier"
	"github.com/ethereum/go-ethereum/common"
)

// IDs of StateDB Operations
const (
	AddBalanceID = iota
	BeginBlockID
	BeginSyncPeriodID
	BeginTransactionID
	CreateAccountID
	CreateContractID
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
	GetStorageRootID
	GetTransientStateID
	HasSelfDestructedID
	RevertToSnapshotID
	SelfDestructID
	SelfDestruct6780ID
	SetCodeID
	SetNonceID
	SetStateID
	SetTransientStateID
	SnapshotID
	SubBalanceID

	// Add new operations below this line

	NumOps
)

// NumArgOps gives the number of operations with encoded argument kinds
const NumArgOps = NumOps * classifier.NumArgKinds * classifier.NumArgKinds * classifier.NumArgKinds

// OpText translates IDs to operation's text
var OpText = map[int]string{
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
	SetTransientStateID: "SetTransientState",
	SnapshotID:          "Snapshot",
	SubBalanceID:        "SubBalance",
}

// opMnemo is a mnemonics table for operations.
var opMnemo = map[int]string{
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
	SetTransientStateID: "ST",
	SnapshotID:          "SN",
	SubBalanceID:        "SB",
}

// OpNumArgs is an argument number table for operations.
var OpNumArgs = map[int]int{
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
	SetTransientStateID: 3,
	SnapshotID:          0,
	SubBalanceID:        1,
}

// opId is an operation ID table.
var opId = map[string]int{
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
	"SS": SetStateID,
	"ST": SetTransientStateID,
	"SN": SnapshotID,
	"SB": SubBalanceID,
}

// argMnemo is the argument-class mnemonics table.
var argMnemo = map[int]string{
	classifier.NoArgID:     "",
	classifier.ZeroArgID:   "z",
	classifier.NewArgID:    "n",
	classifier.PrevArgID:   "p",
	classifier.RecentArgID: "q",
	classifier.RandArgID:   "r",
}

// argId is the argument-class id table.
var argId = map[byte]int{
	'z': classifier.ZeroArgID,
	'n': classifier.NewArgID,
	'p': classifier.PrevArgID,
	'q': classifier.RecentArgID,
	'r': classifier.RandArgID,
}

// OpMnemo returns the mnemonic code for an operation.
func OpMnemo(op int) string {
	if op < 0 || op >= NumOps {
		panic("opcode is out of range")
	}
	return opMnemo[op]
}

// checkArgOp checks whether op/argument combination is valid.
func checkArgOp(op int, contract int, key int, value int) bool {
	if op < 0 || op >= NumOps {
		return false
	}
	if contract < 0 || contract >= classifier.NumArgKinds {
		return false
	}
	if key < 0 || key >= classifier.NumArgKinds {
		return false
	}
	if value < 0 || value >= classifier.NumArgKinds {
		return false
	}
	switch OpNumArgs[op] {
	case 0:
		return contract == classifier.NoArgID &&
			key == classifier.NoArgID &&
			value == classifier.NoArgID
	case 1:
		return contract != classifier.NoArgID &&
			key == classifier.NoArgID &&
			value == classifier.NoArgID
	case 2:
		return contract != classifier.NoArgID &&
			key != classifier.NoArgID &&
			value == classifier.NoArgID
	case 3:
		return contract != classifier.NoArgID &&
			key != classifier.NoArgID &&
			value != classifier.NoArgID
	default:
		return false
	}
}

// IsValidArgOp returns true if the encoding is valid.
func IsValidArgOp(argop int) bool {
	if argop < 0 || argop >= NumArgOps {
		return false
	}
	op, contract, key, value := DecodeArgOp(argop)
	return checkArgOp(op, contract, key, value)
}

// EncodeArgOp encodes operation and argument classes via Horner's scheme to a single value.
func EncodeArgOp(op int, addr int, key int, value int) int {
	if !checkArgOp(op, addr, key, value) {
		log.Fatalf("EncodeArgOp: invalid operation/arguments")
	}
	return (((int(op)*classifier.NumArgKinds)+addr)*classifier.NumArgKinds+key)*classifier.NumArgKinds + value
}

// DecodeArgOp decodes operation with arguments using Honer's scheme
func DecodeArgOp(argop int) (int, int, int, int) {
	if argop < 0 || argop >= NumArgOps {
		log.Fatalf("DecodeArgOp: invalid op range")
	}

	value := argop % classifier.NumArgKinds
	argop = argop / classifier.NumArgKinds

	key := argop % classifier.NumArgKinds
	argop = argop / classifier.NumArgKinds

	addr := argop % classifier.NumArgKinds
	argop = argop / classifier.NumArgKinds

	op := argop

	return op, addr, key, value
}

// EncodeOpcode generates the opcode for an operation and its argument classes.
func EncodeOpcode(op int, addr int, key int, value int) string {
	if !checkArgOp(op, addr, key, value) {
		log.Fatalf("EncodeOpcode: invalid operation/arguments")
	}
	code := fmt.Sprintf("%v%v%v%v", opMnemo[op], argMnemo[addr], argMnemo[key], argMnemo[value])
	if len(code) != 2+OpNumArgs[op] {
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
func DecodeOpcode(opc string) (int, int, int, int) {
	mnemo := opc[:2]
	op, ok := opId[mnemo]
	if !ok {
		log.Fatalf("DecodeOpcode: lookup failed for %v", mnemo)
	}
	if len(opc) != 2+OpNumArgs[op] {
		log.Fatalf("DecodeOpcode: wrong opcode length for %v", opc)
	}
	var contract, key, value int
	switch len(opc) - 2 {
	case 0:
		contract, key, value = classifier.NoArgID, classifier.NoArgID, classifier.NoArgID
	case 1:
		if !validateArg(opc[2]) {
			log.Fatalf("DecodeOpcode: wrong argument code")
		}
		contract, key, value = argId[opc[2]], classifier.NoArgID, classifier.NoArgID
	case 2:
		if !validateArg(opc[2]) || !validateArg(opc[3]) {
			log.Fatalf("DecodeOpcode: wrong argument code")
		}
		contract, key, value = argId[opc[2]], argId[opc[3]], classifier.NoArgID
	case 3:
		if !validateArg(opc[2]) || !validateArg(opc[3]) || !validateArg(opc[4]) {
			log.Fatalf("DecodeOpcode: wrong argument code")
		}
		contract, key, value = argId[opc[2]], argId[opc[3]], argId[opc[4]]
	}
	return op, contract, key, value
}

// ToAddress converts an address index to a contract address.
func ToAddress(idx int64) common.Address {
	var a common.Address
	if idx < 0 {
		panic("invalid index")
	} else if idx != 0 {
		arr := make([]byte, 8)
		binary.LittleEndian.PutUint64(arr, uint64(idx))
		a.SetBytes(arr)
	}
	return a
}

// ToHash converts a key/value index to a hash
func ToHash(idx int64) common.Hash {
	var h common.Hash
	if idx < 0 {
		panic("invalid index")
	} else if idx != 0 {
		// TODO: Improve encoding so that index conversion becomes sparse.
		arr := make([]byte, 32)
		binary.LittleEndian.PutUint64(arr, uint64(idx))
		h.SetBytes(arr)
	}
	return h
}
