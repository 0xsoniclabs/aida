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

package delta

import (
	"context"
	"encoding/hex"
	"fmt"
	"math/big"
	"strconv"
	"strings"

	"github.com/0xsoniclabs/aida/state"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/tracing"
	"github.com/holiman/uint256"
)

// stateReplayer replays logger trace entries against a StateDB implementation.
type stateReplayer struct {
	backend      state.StateDB
	currentBlock uint64
}

// newStateReplayer constructs a replayer for the provided StateDB.
func newStateReplayer(backend state.StateDB) *stateReplayer {
	return &stateReplayer{backend: backend}
}

// Execute runs all trace operations until completion or failure.
func (r *stateReplayer) Execute(ctx context.Context, ops []TraceOp) error {
	for i, op := range ops {
		if err := ctx.Err(); err != nil {
			return err
		}
		if err := r.execute(op); err != nil {
			return fmt.Errorf("operation %d (%s): %w", i, op.Kind, err)
		}
	}
	return nil
}

func (r *stateReplayer) execute(op TraceOp) error {
	if op.Kind == "Bulk" {
		return fmt.Errorf("bulk operations are not supported in logger traces")
	}

	switch op.Kind {
	case "BeginBlock":
		block, err := parseUint64(op.Args, 0)
		if err != nil {
			return err
		}
		if err := r.backend.BeginBlock(block); err != nil {
			return err
		}
		r.currentBlock = block
	case "EndBlock":
		return r.backend.EndBlock()
	case "BeginSyncPeriod":
		number, err := parseUint64(op.Args, 0)
		if err != nil {
			return err
		}
		r.backend.BeginSyncPeriod(number)
	case "EndSyncPeriod":
		r.backend.EndSyncPeriod()
	case "CreateAccount":
		addr, err := parseAddress(op.Args, 0)
		if err != nil {
			return err
		}
		r.backend.CreateAccount(addr)
	case "CreateContract":
		addr, err := parseAddress(op.Args, 0)
		if err != nil {
			return err
		}
		r.backend.CreateContract(addr)
	case "Exist":
		addr, err := parseAddress(op.Args, 0)
		if err != nil {
			return err
		}
		r.backend.Exist(addr)
	case "Empty":
		addr, err := parseAddress(op.Args, 0)
		if err != nil {
			return err
		}
		r.backend.Empty(addr)
	case "SelfDestruct":
		addr, err := parseAddress(op.Args, 0)
		if err != nil {
			return err
		}
		r.backend.SelfDestruct(addr)
	case "SelfDestruct6780":
		addr, err := parseAddress(op.Args, 0)
		if err != nil {
			return err
		}
		r.backend.SelfDestruct6780(addr)
	case "HasSelfDestructed":
		addr, err := parseAddress(op.Args, 0)
		if err != nil {
			return err
		}
		r.backend.HasSelfDestructed(addr)
	case "GetBalance":
		addr, err := parseAddress(op.Args, 0)
		if err != nil {
			return err
		}
		r.backend.GetBalance(addr)
	case "AddBalance":
		addr, err := parseAddress(op.Args, 0)
		if err != nil {
			return err
		}
		value, err := parseUint256(op.Args, 1)
		if err != nil {
			return err
		}
		reason, err := parseBalanceReason(op.Args, 3)
		if err != nil {
			return err
		}
		r.backend.AddBalance(addr, value, reason)
	case "SubBalance":
		addr, err := parseAddress(op.Args, 0)
		if err != nil {
			return err
		}
		value, err := parseUint256(op.Args, 1)
		if err != nil {
			return err
		}
		reason, err := parseBalanceReason(op.Args, 3)
		if err != nil {
			return err
		}
		r.backend.SubBalance(addr, value, reason)
	case "GetNonce":
		addr, err := parseAddress(op.Args, 0)
		if err != nil {
			return err
		}
		r.backend.GetNonce(addr)
	case "SetNonce":
		addr, err := parseAddress(op.Args, 0)
		if err != nil {
			return err
		}
		val, err := parseUint64(op.Args, 1)
		if err != nil {
			return err
		}
		reason, err := parseNonceReason(op.Args, 2)
		if err != nil {
			return err
		}
		r.backend.SetNonce(addr, val, reason)
	case "GetCommittedState":
		addr, err := parseAddress(op.Args, 0)
		if err != nil {
			return err
		}
		key, err := parseHash(op.Args, 1)
		if err != nil {
			return err
		}
		r.backend.GetCommittedState(addr, key)
	case "GetStateAndCommittedState":
		addr, err := parseAddress(op.Args, 0)
		if err != nil {
			return err
		}
		key, err := parseHash(op.Args, 1)
		if err != nil {
			return err
		}
		r.backend.GetStateAndCommittedState(addr, key)
	case "GetState":
		addr, err := parseAddress(op.Args, 0)
		if err != nil {
			return err
		}
		key, err := parseHash(op.Args, 1)
		if err != nil {
			return err
		}
		r.backend.GetState(addr, key)
	case "SetState":
		addr, err := parseAddress(op.Args, 0)
		if err != nil {
			return err
		}
		key, err := parseHash(op.Args, 1)
		if err != nil {
			return err
		}
		value, err := parseHash(op.Args, 2)
		if err != nil {
			return err
		}
		r.backend.SetState(addr, key, value)
	case "SetTransientState":
		addr, err := parseAddress(op.Args, 0)
		if err != nil {
			return err
		}
		key, err := parseHash(op.Args, 1)
		if err != nil {
			return err
		}
		value, err := parseHash(op.Args, 2)
		if err != nil {
			return err
		}
		r.backend.SetTransientState(addr, key, value)
	case "GetTransientState":
		addr, err := parseAddress(op.Args, 0)
		if err != nil {
			return err
		}
		key, err := parseHash(op.Args, 1)
		if err != nil {
			return err
		}
		r.backend.GetTransientState(addr, key)
	case "GetCode":
		addr, err := parseAddress(op.Args, 0)
		if err != nil {
			return err
		}
		r.backend.GetCode(addr)
	case "GetCodeSize":
		addr, err := parseAddress(op.Args, 0)
		if err != nil {
			return err
		}
		r.backend.GetCodeSize(addr)
	case "GetCodeHash":
		addr, err := parseAddress(op.Args, 0)
		if err != nil {
			return err
		}
		r.backend.GetCodeHash(addr)
	case "SetCode":
		addr, err := parseAddress(op.Args, 0)
		if err != nil {
			return err
		}
		code, err := parseByteSlice(op.Args, 1)
		if err != nil {
			return err
		}
		r.backend.SetCode(addr, code, tracing.CodeChangeUnspecified)
	case "Snapshot":
		r.backend.Snapshot()
	case "RevertToSnapshot":
		id, err := parseInt(op.Args, 0)
		if err != nil {
			return err
		}
		r.backend.RevertToSnapshot(id)
	case "BeginTransaction":
		s, err := getArg(op.Args, 0)
		if err != nil {
			return err
		}
		txID, err := parseUint32(s)
		if err != nil {
			return err
		}
		return r.backend.BeginTransaction(txID)
	case "EndTransaction":
		return r.backend.EndTransaction()
	case "Finalise":
		flag, err := parseBool(op.Args, 0)
		if err != nil {
			return err
		}
		r.backend.Finalise(flag)
	case "AddRefund":
		amount, err := parseUint64(op.Args, 0)
		if err != nil {
			return err
		}
		r.backend.AddRefund(amount)
	case "SubRefund":
		amount, err := parseUint64(op.Args, 0)
		if err != nil {
			return err
		}
		r.backend.SubRefund(amount)
	case "GetRefund":
		r.backend.GetRefund()
	case "SetTxContext":
		hash, err := parseHash(op.Args, 0)
		if err != nil {
			return err
		}
		txIndex, err := parseInt(op.Args, 1)
		if err != nil {
			return err
		}
		r.backend.SetTxContext(hash, txIndex)
	case "GetStorageRoot":
		addr, err := parseAddress(op.Args, len(op.Args)-1)
		if err != nil {
			return err
		}
		r.backend.GetStorageRoot(addr)
	case "AddAddressToAccessList":
		addr, err := parseAddress(op.Args, 0)
		if err != nil {
			return err
		}
		r.backend.AddAddressToAccessList(addr)
	case "AddSlotToAccessList":
		addr, err := parseAddress(op.Args, 0)
		if err != nil {
			return err
		}
		slot, err := parseHash(op.Args, 1)
		if err != nil {
			return err
		}
		r.backend.AddSlotToAccessList(addr, slot)
	case "AddressInAccessList":
		addr, err := parseAddress(op.Args, 0)
		if err != nil {
			return err
		}
		r.backend.AddressInAccessList(addr)
	case "SlotInAccessList":
		addr, err := parseAddress(op.Args, 0)
		if err != nil {
			return err
		}
		slot, err := parseHash(op.Args, 1)
		if err != nil {
			return err
		}
		r.backend.SlotInAccessList(addr, slot)
	case "GetLogs":
		txHash, err := parseHash(op.Args, 0)
		if err != nil {
			return err
		}
		blockNumber, err := parseUint64(op.Args, 1)
		if err != nil {
			return err
		}
		blockHash, err := parseHash(op.Args, 2)
		if err != nil {
			return err
		}
		timestamp, err := parseUint64(op.Args, 3)
		if err != nil {
			return err
		}
		r.backend.GetLogs(txHash, blockNumber, blockHash, timestamp)
	case "GetHash":
		_, _ = r.backend.GetHash()
	case "IntermediateRoot":
		flag, err := parseBool(op.Args, 0)
		if err != nil {
			return err
		}
		r.backend.IntermediateRoot(flag)
	case "Commit":
		deleteEmpty, err := parseBool(op.Args, 0)
		if err != nil {
			return err
		}
		_, err = r.backend.Commit(r.currentBlock, deleteEmpty)
		return err
	case "AddPreimage":
		hash, err := parseHash(op.Args, 0)
		if err != nil {
			return err
		}
		data, err := parseByteSlice(op.Args, 1)
		if err != nil {
			return err
		}
		r.backend.AddPreimage(hash, data)
	case "AccessEvents":
		r.backend.AccessEvents()
	case "PointCache":
		r.backend.PointCache()
	case "Witness":
		r.backend.Witness()
	case "GetSubstatePostAlloc":
		r.backend.GetSubstatePostAlloc()
	case "GetArchiveBlockHeight":
		_, _, _ = r.backend.GetArchiveBlockHeight()
	case "GetCodeHashLc", "GetCodeHashLcS", "GetStateLccs", "GetStateLc", "GetStateLcls":
		return fmt.Errorf("operation %s is not supported in logger traces", op.Kind)
	case "AddLog", "Prepare", "PrepareSubstate", "Close", "Error", "Release":
		return nil
	default:
		return fmt.Errorf("unsupported operation kind %q", op.Kind)
	}

	return nil
}

func parseAddress(args []string, idx int) (common.Address, error) {
	s, err := getArg(args, idx)
	if err != nil {
		return common.Address{}, err
	}
	return common.HexToAddress(s), nil
}

func parseHash(args []string, idx int) (common.Hash, error) {
	s, err := getArg(args, idx)
	if err != nil {
		return common.Hash{}, err
	}
	return common.HexToHash(s), nil
}

func parseUint64(args []string, idx int) (uint64, error) {
	s, err := getArg(args, idx)
	if err != nil {
		return 0, err
	}
	if strings.HasPrefix(s, "0x") || strings.HasPrefix(s, "0X") {
		return strconv.ParseUint(s[2:], 16, 64)
	}
	return strconv.ParseUint(s, 10, 64)
}

func parseUint32(s string) (uint32, error) {
	v, err := strconv.ParseUint(s, 10, 32)
	if err != nil {
		return 0, err
	}
	return uint32(v), nil
}

func parseInt(args []string, idx int) (int, error) {
	s, err := getArg(args, idx)
	if err != nil {
		return 0, err
	}
	if strings.HasPrefix(s, "0x") || strings.HasPrefix(s, "0X") {
		val, err := strconv.ParseInt(s[2:], 16, 32)
		return int(val), err
	}
	val, err := strconv.ParseInt(s, 10, 32)
	return int(val), err
}

func parseBool(args []string, idx int) (bool, error) {
	s, err := getArg(args, idx)
	if err != nil {
		return false, err
	}
	return strconv.ParseBool(strings.ToLower(s))
}

func parseUint256(args []string, idx int) (*uint256.Int, error) {
	s, err := getArg(args, idx)
	if err != nil {
		return nil, err
	}
	num := new(big.Int)
	base := 10
	if strings.HasPrefix(s, "0x") || strings.HasPrefix(s, "0X") {
		base = 16
		s = s[2:]
	}
	if _, ok := num.SetString(s, base); !ok {
		return nil, fmt.Errorf("invalid uint256 value %q", s)
	}
	if num.Sign() < 0 {
		return nil, fmt.Errorf("negative uint256 value %q", s)
	}
	out := new(uint256.Int)
	if overflow := out.SetFromBig(num); overflow {
		return nil, fmt.Errorf("uint256 overflow for %q", s)
	}
	return out, nil
}

func parseByteSlice(args []string, idx int) ([]byte, error) {
	s, err := getArg(args, idx)
	if err != nil {
		return nil, err
	}
	s = strings.TrimSpace(s)
	if strings.HasPrefix(s, "0x") && len(s) >= 2 {
		return hex.DecodeString(s[2:])
	}
	if len(s) >= 2 && s[0] == '[' && s[len(s)-1] == ']' {
		s = s[1 : len(s)-1]
	}
	s = strings.TrimSpace(s)
	if s == "" {
		return []byte{}, nil
	}
	parts := strings.Fields(strings.ReplaceAll(s, ",", " "))
	out := make([]byte, len(parts))
	for i, part := range parts {
		val, err := strconv.ParseInt(part, 10, 8)
		if err != nil {
			return nil, fmt.Errorf("invalid byte value %q: %w", part, err)
		}
		out[i] = byte(val)
	}
	return out, nil
}

func parseBalanceReason(args []string, idx int) (tracing.BalanceChangeReason, error) {
	raw, err := getArg(args, idx)
	if err != nil {
		return tracing.BalanceChangeUnspecified, err
	}
	name := normalizeEnumName(raw)
	if reason, ok := balanceReasonByName[name]; ok {
		return reason, nil
	}
	if reason, ok := balanceReasonByName[strings.ToLower(name)]; ok {
		return reason, nil
	}
	if val, ok := parseEnumNumber(name); ok {
		if reason, ok := balanceReasonByValue[val]; ok {
			return reason, nil
		}
	}
	return tracing.BalanceChangeUnspecified, nil
}

func parseNonceReason(args []string, idx int) (tracing.NonceChangeReason, error) {
	raw, err := getArg(args, idx)
	if err != nil {
		return tracing.NonceChangeUnspecified, err
	}
	name := normalizeEnumName(raw)
	if reason, ok := nonceReasonByName[name]; ok {
		return reason, nil
	}
	if reason, ok := nonceReasonByName[strings.ToLower(name)]; ok {
		return reason, nil
	}
	if val, ok := parseEnumNumber(name); ok {
		if reason, ok := nonceReasonByValue[val]; ok {
			return reason, nil
		}
	}
	return tracing.NonceChangeUnspecified, nil
}

func getArg(args []string, idx int) (string, error) {
	if idx < 0 || idx >= len(args) {
		return "", fmt.Errorf("missing argument %d", idx)
	}
	return strings.TrimSpace(args[idx]), nil
}

var balanceReasonByName = func() map[string]tracing.BalanceChangeReason {
	res := make(map[string]tracing.BalanceChangeReason)
	for r := tracing.BalanceChangeReason(0); r <= tracing.BalanceChangeReason(15); r++ {
		label := r.String()
		res[label] = r
		res[strings.ToLower(label)] = r
	}
	return res
}()

var nonceReasonByName = func() map[string]tracing.NonceChangeReason {
	res := make(map[string]tracing.NonceChangeReason)
	for r := tracing.NonceChangeReason(0); r <= tracing.NonceChangeReason(6); r++ {
		label := r.String()
		res[label] = r
		res[strings.ToLower(label)] = r
	}
	return res
}()

var balanceReasonByValue = func() map[uint64]tracing.BalanceChangeReason {
	res := make(map[uint64]tracing.BalanceChangeReason)
	for r := tracing.BalanceChangeReason(0); r <= tracing.BalanceChangeReason(15); r++ {
		res[uint64(r)] = r
	}
	return res
}()

var nonceReasonByValue = func() map[uint64]tracing.NonceChangeReason {
	res := make(map[uint64]tracing.NonceChangeReason)
	for r := tracing.NonceChangeReason(0); r <= tracing.NonceChangeReason(6); r++ {
		res[uint64(r)] = r
	}
	return res
}()

func normalizeEnumName(raw string) string {
	name := strings.TrimSpace(raw)
	name = strings.Trim(name, "\"'")
	if idx := strings.IndexAny(name, " \t("); idx >= 0 {
		name = name[:idx]
	}
	name = strings.TrimSuffix(name, ",")
	name = strings.TrimSpace(name)
	return name
}

func parseEnumNumber(input string) (uint64, bool) {
	trimmed := strings.TrimSpace(input)
	if trimmed == "" {
		return 0, false
	}
	lower := strings.ToLower(trimmed)
	var (
		val uint64
		err error
	)
	if strings.HasPrefix(lower, "0x") {
		val, err = strconv.ParseUint(strings.TrimPrefix(lower, "0x"), 16, 8)
	} else {
		val, err = strconv.ParseUint(trimmed, 10, 8)
	}
	if err != nil {
		return 0, false
	}
	return val, true
}
