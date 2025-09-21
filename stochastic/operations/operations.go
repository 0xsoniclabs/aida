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

package operations

import (
	"encoding/binary"
	"fmt"

	"github.com/0xsoniclabs/aida/stochastic"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
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
	GetStateAndCommittedStateID
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
const NumArgOps = NumOps * stochastic.NumArgKinds * stochastic.NumArgKinds * stochastic.NumArgKinds

// OpText translates IDs to operation's text
var OpText = map[int]string{
	AddBalanceID:                "AddBalance",
	BeginBlockID:                "BeginBlock",
	BeginSyncPeriodID:           "BeginSyncPeriod",
	BeginTransactionID:          "BeginTransaction",
	CreateAccountID:             "CreateAccount",
	CreateContractID:            "CreateContract",
	EmptyID:                     "Empty",
	EndBlockID:                  "EndBlock",
	EndSyncPeriodID:             "EndSyncPeriod",
	EndTransactionID:            "EndTransaction",
	ExistID:                     "Exist",
	GetBalanceID:                "GetBalance",
	GetCodeHashID:               "GetCodeHash",
	GetCodeID:                   "GetCode",
	GetCodeSizeID:               "GetCodeSize",
	GetCommittedStateID:         "GetCommittedState",
	GetNonceID:                  "GetNonce",
	GetStateID:                  "GetState",
	GetStateAndCommittedStateID: "GetCommittedStateAndState",
	GetStorageRootID:            "GetStorageRoot",
	GetTransientStateID:         "GetTransientState",
	HasSelfDestructedID:         "HasSelfDestructed",
	RevertToSnapshotID:          "RevertToSnapshot",
	SelfDestructID:              "SelfDestruct",
	SelfDestruct6780ID:          "SelfDestruct6780",
	SetCodeID:                   "SetCode",
	SetNonceID:                  "SetNonce",
	SetStateID:                  "SetState",
	SetTransientStateID:         "SetTransientState",
	SnapshotID:                  "Snapshot",
	SubBalanceID:                "SubBalance",
}

// opMnemo is a mnemonics table for operations.
var opMnemo = map[int]string{
	AddBalanceID:                "AB",
	BeginBlockID:                "BB",
	BeginSyncPeriodID:           "BS",
	BeginTransactionID:          "BT",
	CreateAccountID:             "CA",
	CreateContractID:            "CC",
	EmptyID:                     "EM",
	EndBlockID:                  "EB",
	EndSyncPeriodID:             "ES",
	EndTransactionID:            "ET",
	ExistID:                     "EX",
	GetBalanceID:                "GB",
	GetCodeHashID:               "GH",
	GetCodeID:                   "GC",
	GetCodeSizeID:               "GZ",
	GetCommittedStateID:         "GM",
	GetNonceID:                  "GN",
	GetStateID:                  "GS",
	GetStateAndCommittedStateID: "CS",
	GetStorageRootID:            "GR",
	GetTransientStateID:         "GT",
	HasSelfDestructedID:         "HS",
	RevertToSnapshotID:          "RS",
	SelfDestructID:              "SU",
	SelfDestruct6780ID:          "S6",
	SetCodeID:                   "SC",
	SetNonceID:                  "SO",
	SetStateID:                  "SS",
	SetTransientStateID:         "ST",
	SnapshotID:                  "SN",
	SubBalanceID:                "SB",
}

// OpNumArgs is an argument number table for operations.
var OpNumArgs = map[int]int{
	AddBalanceID:                1,
	BeginBlockID:                0,
	BeginSyncPeriodID:           0,
	BeginTransactionID:          0,
	CreateAccountID:             1,
	CreateContractID:            1,
	EmptyID:                     1,
	EndBlockID:                  0,
	EndSyncPeriodID:             0,
	EndTransactionID:            0,
	ExistID:                     1,
	GetBalanceID:                1,
	GetCodeHashID:               1,
	GetCodeID:                   1,
	GetCodeSizeID:               1,
	GetCommittedStateID:         2,
	GetNonceID:                  1,
	GetStateID:                  2,
	GetStateAndCommittedStateID: 2,
	GetStorageRootID:            1,
	GetTransientStateID:         2,
	HasSelfDestructedID:         1,
	RevertToSnapshotID:          0,
	SelfDestructID:              1,
	SelfDestruct6780ID:          1,
	SetCodeID:                   1,
	SetNonceID:                  1,
	SetStateID:                  3,
	SetTransientStateID:         3,
	SnapshotID:                  0,
	SubBalanceID:                1,
}

// opId is an operation ID table.
var opId = map[string]int{
	"AB": AddBalanceID,
	"BB": BeginBlockID,
	"BS": BeginSyncPeriodID,
	"BT": BeginTransactionID,
	"CA": CreateAccountID,
	"CC": CreateContractID,
	"CS": GetStateAndCommittedStateID,
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
	stochastic.NoArgID:     "",
	stochastic.ZeroArgID:   "z",
	stochastic.NewArgID:    "n",
	stochastic.PrevArgID:   "p",
	stochastic.RecentArgID: "q",
	stochastic.RandArgID:   "r",
}

// argId is the argument-class id table.
var argId = map[byte]int{
	'z': stochastic.ZeroArgID,
	'n': stochastic.NewArgID,
	'p': stochastic.PrevArgID,
	'q': stochastic.RecentArgID,
	'r': stochastic.RandArgID,
}

// OpMnemo returns the mnemonic code for an operation.
func OpMnemo(op int) string {
	if op < 0 || op >= NumOps {
		panic("opcode is out of range")
	}
	return opMnemo[op]
}

// checkArgOp checks whether op/argument combination is valid.
func checkArgOp(op int, contract int, key int, value int) error {
	if op < 0 || op >= NumOps {
		return fmt.Errorf("checkArgOp: operations out of range %v not in [0,%v]", op, NumOps)
	}
	if contract < 0 || contract >= stochastic.NumArgKinds {
		return fmt.Errorf("checkArgOp: contract arg out of range %v not in [0,%v]", contract, stochastic.NumArgKinds)
	}
	if key < 0 || key >= stochastic.NumArgKinds {
		return fmt.Errorf("checkArgOp: key arg out of range %v not in [0,%v]", key, stochastic.NumArgKinds)
	}
	if value < 0 || value >= stochastic.NumArgKinds {
		return fmt.Errorf("checkArgOp: value arg out of range %v not in [0,%v]", value, stochastic.NumArgKinds)
	}
	switch OpNumArgs[op] {
	case 0:
		if !(contract == stochastic.NoArgID &&
			key == stochastic.NoArgID &&
			value == stochastic.NoArgID) {
			return fmt.Errorf("checkArgOp: op %v takes no arguments", op)
		}
	case 1:
		if !(contract != stochastic.NoArgID &&
			key == stochastic.NoArgID &&
			value == stochastic.NoArgID) {
			return fmt.Errorf("checkArgOp: op %v takes one contract argument", op)
		}
	case 2:
		if !(contract != stochastic.NoArgID &&
			key != stochastic.NoArgID &&
			value == stochastic.NoArgID) {
			return fmt.Errorf("checkArgOp: op %v takes contract and key arguments", op)
		}
	case 3:
		if !(contract != stochastic.NoArgID &&
			key != stochastic.NoArgID &&
			value != stochastic.NoArgID) {
			return fmt.Errorf("checkArgOp: op %v takes contract, key and value arguments", op)
		}
	default:
		return fmt.Errorf("checkArgOp: op %v has invalid number of arguments", op)
	}
	return nil
}

// IsValidArgOp returns true if the encoding of an operation with its argument is valid.
func IsValidArgOp(argop int) bool {
	if argop < 0 || argop >= NumArgOps {
		return false
	}
	_, _, _, _, err := DecodeArgOp(argop)
	return err == nil
}

// EncodeArgOp encodes operation and argument classes via Horner's scheme to a single value.
func EncodeArgOp(op int, addr int, key int, value int) (int, error) {
	if err := checkArgOp(op, addr, key, value); err != nil {
		return 0, fmt.Errorf("EncodeArgOp: invalid operation/arguments. Error: %v", err)
	}
	return (((int(op)*stochastic.NumArgKinds)+addr)*stochastic.NumArgKinds+key)*stochastic.NumArgKinds + value, nil
}

// DecodeArgOp decodes operation with arguments using Horner's scheme
func DecodeArgOp(argop int) (int, int, int, int, error) {
	if argop < 0 || argop >= NumArgOps {
		return 0, 0, 0, 0, fmt.Errorf("DecodeArgOp: argument out of range %v not in [0,%v]", argop, NumArgOps)
	}

	value := argop % stochastic.NumArgKinds
	argop = argop / stochastic.NumArgKinds

	key := argop % stochastic.NumArgKinds
	argop = argop / stochastic.NumArgKinds

	addr := argop % stochastic.NumArgKinds
	argop = argop / stochastic.NumArgKinds

	op := argop

	if err := checkArgOp(op, addr, key, value); err != nil {
		return 0, 0, 0, 0, fmt.Errorf("DecodeArgOp: invalid operation/arguments. Error: %v", err)
	}

	return op, addr, key, value, nil
}

// EncodeOpcode generates the opcode for an operation and its argument classes.
func EncodeOpcode(op int, addr int, key int, value int) (string, error) {
	if err := checkArgOp(op, addr, key, value); err != nil {
		return "", fmt.Errorf("EncodeOpcode: invalid operation/arguments. Error: %v", err)
	}
	code := fmt.Sprintf("%v%v%v%v", opMnemo[op], argMnemo[addr], argMnemo[key], argMnemo[value])
	if len(code) != 2+OpNumArgs[op] {
		return "", fmt.Errorf("EncodeOpcode: wrong opcode length for opcode %v (%v). Expected", code, 2+OpNumArgs[op])
	}
	return code, nil
}

// validateArg checks whether argument mnemonics exists.
func validateArg(argMnemo byte) bool {
	_, ok := argId[argMnemo]
	return ok
}

// DecodeOpcode decodes an string opcode encoding operation and its argument to a tuple representation
func DecodeOpcode(opc string) (int, int, int, int, error) {
	mnemo := opc[:2]
	op, ok := opId[mnemo]
	if !ok {
		return 0, 0, 0, 0, fmt.Errorf("DecodeOpcode: lookup failed for %v.", mnemo)
	}
	if len(opc) != 2+OpNumArgs[op] {
		return 0, 0, 0, 0, fmt.Errorf("DecodeOpcode: wrong opcode length for %v", opc)
	}
	var contract, key, value int
	switch len(opc) - 2 {
	case 0:
		contract, key, value = stochastic.NoArgID, stochastic.NoArgID, stochastic.NoArgID
	case 1:
		if !validateArg(opc[2]) {
			return 0, 0, 0, 0, fmt.Errorf("DecodeOpcode: wrong argument code %v", opc)
		}
		contract, key, value = argId[opc[2]], stochastic.NoArgID, stochastic.NoArgID
	case 2:
		if !validateArg(opc[2]) || !validateArg(opc[3]) {
			return 0, 0, 0, 0, fmt.Errorf("DecodeOpcode: wrong argument code %v", opc)
		}
		contract, key, value = argId[opc[2]], argId[opc[3]], stochastic.NoArgID
	case 3:
		if !validateArg(opc[2]) || !validateArg(opc[3]) || !validateArg(opc[4]) {
			return 0, 0, 0, 0, fmt.Errorf("DecodeOpcode: wrong argument code %v", opc)
		}
		contract, key, value = argId[opc[2]], argId[opc[3]], argId[opc[4]]
	}
	return op, contract, key, value, nil
}

// ToAddress converts an address index to a contract address.
func ToAddress(idx int64) (common.Address, error) {
	var a common.Address
	if idx < 0 {
		return a, fmt.Errorf("invalid index (%v)", idx)
	}
	if idx != 0 {
		arr := make([]byte, binary.MaxVarintLen64)
		binary.PutVarint(arr, -idx)
		a.SetBytes(crypto.Keccak256(arr))
	}
	return a, nil
}

// ToHash converts a key/value index to a hash
func ToHash(idx int64) (common.Hash, error) {
	var h common.Hash
	if idx < 0 {
		return h, fmt.Errorf("invalid index (%v)", idx)
	}
	if idx != 0 {
		arr := make([]byte, binary.MaxVarintLen64)
		binary.PutVarint(arr, -idx)
		h = crypto.Keccak256Hash(arr)
	}
	return h, nil
}
