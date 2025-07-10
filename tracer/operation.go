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
	"errors"
	"fmt"
	"log"

	"github.com/ethereum/go-ethereum/common"
)

type Operation struct {
	Op    uint8
	Addr  common.Address
	Key   common.Hash
	Value common.Hash
}

// IDs of StateDB Operations

// numArgOps gives the number of operations with encoded argument classes
const numArgOps = uint16(NumOps) * uint16(NumClasses) * uint16(NumClasses) * uint16(NumClasses)

// OpText translates IDs to operation's text
var OpText = map[uint8]string{
	AddBalanceID:             "AddBalance",
	BeginBlockID:             "BeginBlock",
	BeginSyncPeriodID:        "BeginSyncPeriod",
	BeginTransactionID:       "BeginTransaction",
	CreateAccountID:          "CreateAccount",
	CreateContractID:         "CreateContract",
	EmptyID:                  "Empty",
	EndBlockID:               "EndBlock",
	EndSyncPeriodID:          "EndSyncPeriod",
	EndTransactionID:         "EndTransaction",
	ExistID:                  "Exist",
	GetBalanceID:             "GetBalance",
	GetCodeHashID:            "GetCodeHash",
	GetCodeID:                "GetCode",
	GetCodeSizeID:            "GetCodeSize",
	GetCommittedStateID:      "GetCommittedState",
	GetNonceID:               "GetNonce",
	GetStateID:               "GetState",
	GetStorageRootID:         "GetStorageRoot",
	GetTransientStateID:      "GetTransientState",
	HasSelfDestructedID:      "HasSelfDestructed",
	RevertToSnapshotID:       "RevertToSnapshot",
	SelfDestructID:           "SelfDestruct",
	SelfDestruct6780ID:       "SelfDestruct6780",
	SetCodeID:                "SetCode",
	SetNonceID:               "SetNonce",
	SetStateID:               "SetState",
	SnapshotID:               "Snapshot",
	SubBalanceID:             "SubBalance",
	SetTransientStateID:      "SetTransientState",
	GetRefundID:              "GetRefundID",
	SubRefundID:              "SubRefund",
	AddRefundID:              "AddRefund",
	PrepareID:                "Prepare",
	AddAddressToAccessListID: "AddAddressToAccessList",
	AddressInAccessListID:    "AddressInAccessList",
	SlotInAccessListID:       "SlotInAccessList",
	AddSlotToAccessListID:    "AddSlotToAccessList",
	AddLogID:                 "AddLog",
	GetLogsID:                "GetLogs",
	PointCacheID:             "PointCache",
	WitnessID:                "Witness",
	AddPreimageID:            "AddPreimage",
	SetTxContextID:           "SetTxContext",
	FinaliseID:               "Finalise",
	IntermediateRootID:       "IntermediateRoot",
	CommitID:                 "Commit",
	CloseID:                  "Close",
	AccessEventsID:           "AccessEvents",
	GetHashID:                "GetHash",
	GetSubstatePostAllocID:   "GetSubstatePostAllocID",
	PrepareSubstateID:        "PrepareSubstateID",
	GetArchiveStateID:        "GetArchiveStateID",
	GetArchiveBlockHeightID:  "GetArchiveBlockHeightID",
}

// opMnemo is a mnemonics table for operations.
var opMnemo = map[uint8]string{
	AddBalanceID:             "AB",
	BeginBlockID:             "BB",
	BeginSyncPeriodID:        "BS",
	BeginTransactionID:       "BT",
	CreateAccountID:          "CA",
	CreateContractID:         "CC",
	EmptyID:                  "EM",
	EndBlockID:               "EB",
	EndSyncPeriodID:          "ES",
	EndTransactionID:         "ET",
	ExistID:                  "EX",
	GetBalanceID:             "GB",
	GetCodeHashID:            "GH",
	GetCodeID:                "GC",
	GetCodeSizeID:            "GZ",
	GetCommittedStateID:      "GM",
	GetNonceID:               "GN",
	GetStateID:               "GS",
	GetStorageRootID:         "GR",
	GetTransientStateID:      "GT",
	HasSelfDestructedID:      "HS",
	RevertToSnapshotID:       "RS",
	SelfDestructID:           "SU",
	SelfDestruct6780ID:       "S6",
	SetCodeID:                "SC",
	SetNonceID:               "SO",
	SetStateID:               "SS",
	SnapshotID:               "SN",
	SubBalanceID:             "SB",
	SetTransientStateID:      "ST",
	GetRefundID:              "GF",
	SubRefundID:              "SF",
	AddRefundID:              "AF",
	PrepareID:                "PR",
	AddAddressToAccessListID: "AA",
	AddressInAccessListID:    "AI",
	SlotInAccessListID:       "SI",
	AddSlotToAccessListID:    "AS",
	AddLogID:                 "AL",
	GetLogsID:                "GL",
	PointCacheID:             "PC",
	WitnessID:                "WI",
	AddPreimageID:            "AP",
	SetTxContextID:           "TX",
	FinaliseID:               "FI",
	IntermediateRootID:       "IR",
	CommitID:                 "CM",
	CloseID:                  "CL",
	AccessEventsID:           "AE",
	GetHashID:                "GA",
	GetSubstatePostAllocID:   "GP",
	PrepareSubstateID:        "PS",
	GetArchiveStateID:        "AR",
	GetArchiveBlockHeightID:  "BH",
}

// opNumArgs is an argument number table for operations.
var opNumArgs = map[uint8]int{
	AddBalanceID:             1,
	BeginBlockID:             0,
	BeginSyncPeriodID:        0,
	BeginTransactionID:       0,
	CreateAccountID:          1,
	EmptyID:                  1,
	EndBlockID:               0,
	EndSyncPeriodID:          0,
	EndTransactionID:         0,
	ExistID:                  1,
	GetBalanceID:             1,
	GetCodeHashID:            1,
	GetCodeID:                1,
	GetCodeSizeID:            1,
	GetCommittedStateID:      2,
	GetNonceID:               1,
	GetStateID:               2,
	HasSelfDestructedID:      1,
	RevertToSnapshotID:       0,
	SetCodeID:                1,
	SetNonceID:               1,
	SetStateID:               3,
	SnapshotID:               0,
	SubBalanceID:             1,
	SelfDestructID:           1,
	CreateContractID:         1,
	GetStorageRootID:         1,
	GetTransientStateID:      2,
	SetTransientStateID:      3,
	SelfDestruct6780ID:       1,
	SubRefundID:              0,
	GetRefundID:              0,
	AddRefundID:              0,
	PrepareID:                0,
	AddAddressToAccessListID: 1,
	AddressInAccessListID:    1,
	SlotInAccessListID:       2,
	AddSlotToAccessListID:    2,
	AddLogID:                 0,
	GetLogsID:                0,
	PointCacheID:             0,
	WitnessID:                0,
	AddPreimageID:            0,
	SetTxContextID:           0,
	FinaliseID:               0,
	IntermediateRootID:       0,
	CommitID:                 0,
	CloseID:                  0,
	AccessEventsID:           0,
	GetHashID:                0,
	GetSubstatePostAllocID:   0,
	PrepareSubstateID:        0,
	GetArchiveStateID:        0,
	GetArchiveBlockHeightID:  0,
}

// opId is an operation ID table.
var opId = map[string]uint8{
	"AB": AddBalanceID,
	"BB": BeginBlockID,
	"BS": BeginSyncPeriodID,
	"BT": BeginTransactionID,
	"CA": CreateAccountID,
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
	"HS": HasSelfDestructedID,
	"RS": RevertToSnapshotID,
	"SC": SetCodeID,
	"SO": SetNonceID,
	"SS": SetStateID,
	"SN": SnapshotID,
	"SB": SubBalanceID,
	"SU": SelfDestructID,
	"CC": CreateContractID,
	"GR": GetStorageRootID,
	"GT": GetTransientStateID,
	"S6": SelfDestruct6780ID,
	"ST": SetTransientStateID,
	"GF": GetRefundID,
	"SF": SubRefundID,
	"AF": AddRefundID,
	"PR": PrepareID,
	"AA": AddAddressToAccessListID,
	"AI": AddressInAccessListID,
	"SI": SlotInAccessListID,
	"AS": AddSlotToAccessListID,
	"AL": AddLogID,
	"GL": GetLogsID,
	"PC": PointCacheID,
	"WI": WitnessID,
	"AP": AddPreimageID,
	"TX": SetTxContextID,
	"FI": FinaliseID,
	"IR": IntermediateRootID,
	"CM": CommitID,
	"CL": CloseID,
	"AE": AccessEventsID,
	"GA": GetHashID,
	"GP": GetSubstatePostAllocID,
	"PS": PrepareSubstateID,
	"AR": GetArchiveStateID,
	"BH": GetArchiveBlockHeightID,
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
func OpMnemo(op uint8) string {
	if op >= NumOps {
		panic("opcode is out of range")
	}
	return opMnemo[op]
}

// isValidOp checks whether op/argument combination is valid.
func isValidOp(op uint8, contract uint8, key uint8, value uint8) bool {
	if op >= NumOps {
		return false
	}
	if contract >= NumClasses {
		return false
	}
	if key >= NumClasses {
		return false
	}
	if value >= NumClasses {
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

// EncodeArgOp encodes operation and argument classes via Horner's scheme to a single value.
func EncodeArgOp(op uint8, addr uint8, key uint8, value uint8) (uint16, error) {
	if !isValidOp(op, addr, key, value) {
		return 0, fmt.Errorf("encodeArgOp: invalid operation/arguments\naddr: %d, key: %d, value: %d, op: %d", addr, key, value, op)
	}
	numClasses := uint16(NumClasses)
	return (((uint16(op)*numClasses)+uint16(addr))*numClasses+uint16(key))*numClasses + uint16(value), nil
}

// DecodeArgOp decodes operation with arguments using Honer's scheme
func DecodeArgOp(argop uint16) (uint8, uint8, uint8, uint8, error) {
	if argop >= numArgOps {
		return 0, 0, 0, 0, errors.New("DecodeArgOp: invalid op range")
	}

	value := argop % uint16(NumClasses)
	argop = argop / uint16(NumClasses)

	key := argop % uint16(NumClasses)
	argop = argop / uint16(NumClasses)

	addr := argop % uint16(NumClasses)
	argop = argop / uint16(NumClasses)

	op := argop

	return uint8(op), uint8(addr), uint8(key), uint8(value), nil
}

// EncodeOpcode generates the opcode for an operation and its argument classes.
func EncodeOpcode(op uint8, addr uint8, key uint8, value uint8) (string, error) {
	if !isValidOp(op, addr, key, value) {
		return "", fmt.Errorf("EncodeOpcode: invalid operation/arguments")
	}
	code := fmt.Sprintf("%v%v%v%v", opMnemo[op], argMnemo[addr], argMnemo[key], argMnemo[value])
	if len(code) != 2+opNumArgs[op] {
		return "", fmt.Errorf("EncodeOpcode: wrong opcode length for opcode %v", code)
	}
	return code, nil
}

// validateArg checks whether argument mnemonics exists.
func validateArg(argMnemo byte) bool {
	_, ok := argId[argMnemo]
	return ok
}

// DecodeOpcode decodes opcode producing the operation id and its argument classes
func DecodeOpcode(opc string) (uint8, uint8, uint8, uint8, error) {
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
