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

package validator

import (
	"bytes"
	"fmt"
	"sync/atomic"

	"github.com/0xsoniclabs/aida/executor"
	"github.com/0xsoniclabs/aida/executor/extension"
	"github.com/0xsoniclabs/aida/logger"
	"github.com/0xsoniclabs/aida/state"
	"github.com/0xsoniclabs/aida/txcontext"
	"github.com/0xsoniclabs/aida/utils"
	cc "github.com/0xsoniclabs/carmen/go/common"
	"github.com/0xsoniclabs/substate/substate"
	stypes "github.com/0xsoniclabs/substate/types"
	"github.com/ethereum/go-ethereum/common"
)

// MakeLiveDbValidator creates an extension which validates LIVE StateDb
func MakeLiveDbValidator(cfg *utils.Config, target ValidateTxTarget) executor.Extension[txcontext.TxContext] {
	if !cfg.ValidateTxState {
		return extension.NilExtension[txcontext.TxContext]{}
	}

	log := logger.NewLogger(cfg.LogLevel, "Tx-Verifier")

	return makeLiveDbValidator(cfg, log, target)
}

func makeLiveDbValidator(cfg *utils.Config, log logger.Logger, target ValidateTxTarget) *liveDbTxValidator {
	return &liveDbTxValidator{
		makeStateDbValidator(cfg, log, target),
	}
}

type liveDbTxValidator struct {
	*stateDbValidator
}

// PreTransaction validates InputSubstate in given substate
func (v *liveDbTxValidator) PreTransaction(state executor.State[txcontext.TxContext], ctx *executor.Context) error {
	return v.runPreTxValidation("live-db-validator", ctx.State, state, ctx.ErrorInput)
}

// PostTransaction validates OutputAlloc in given substate
func (v *liveDbTxValidator) PostTransaction(state executor.State[txcontext.TxContext], ctx *executor.Context) error {
	return v.runPostTxValidation("live-db-validator", ctx.State, state, ctx.ExecutionResult, ctx.ErrorInput)
}

// MakeArchiveDbValidator creates an extension which validates ARCHIVE StateDb
func MakeArchiveDbValidator(cfg *utils.Config, target ValidateTxTarget) executor.Extension[txcontext.TxContext] {
	if !cfg.ValidateTxState {
		return extension.NilExtension[txcontext.TxContext]{}
	}

	log := logger.NewLogger(cfg.LogLevel, "Tx-Verifier")

	return makeArchiveDbValidator(cfg, log, target)
}

func makeArchiveDbValidator(cfg *utils.Config, log logger.Logger, target ValidateTxTarget) *archiveDbValidator {
	return &archiveDbValidator{
		makeStateDbValidator(cfg, log, target),
	}
}

type archiveDbValidator struct {
	*stateDbValidator
}

// PreTransaction validates the input WorldState before transaction is executed.
func (v *archiveDbValidator) PreTransaction(state executor.State[txcontext.TxContext], ctx *executor.Context) error {
	return v.runPreTxValidation("archive-db-validator", ctx.Archive, state, ctx.ErrorInput)
}

// PostTransaction validates the resulting WorldState after transaction is executed.
func (v *archiveDbValidator) PostTransaction(state executor.State[txcontext.TxContext], ctx *executor.Context) error {
	return v.runPostTxValidation("archive-db-validator", ctx.Archive, state, ctx.ExecutionResult, ctx.ErrorInput)
}

// makeStateDbValidator creates an extension that validates StateDb.
// stateDbValidator should always be inherited depending on what
// type of StateDb we are working with
func makeStateDbValidator(cfg *utils.Config, log logger.Logger, target ValidateTxTarget) *stateDbValidator {
	return &stateDbValidator{
		cfg:            cfg,
		log:            log,
		numberOfErrors: new(atomic.Int32),
		target:         target,
	}
}

type stateDbValidator struct {
	extension.NilExtension[txcontext.TxContext]
	cfg            *utils.Config
	log            logger.Logger
	numberOfErrors *atomic.Int32
	target         ValidateTxTarget
}

// ValidateTxTarget serves for the validator to determine what type of validation to run
type ValidateTxTarget struct {
	WorldState bool // validate state before and after processing a transaction
	Receipt    bool // validate content of transaction receipt
}

// PreRun informs the user that stateDbValidator is enabled and that they should expect slower processing speed.
func (v *stateDbValidator) PreRun(executor.State[txcontext.TxContext], *executor.Context) error {
	v.log.Warning("Transaction verification is enabled, this may slow down the block processing.")

	if v.cfg.ContinueOnFailure {
		v.log.Warningf("Continue on Failure for transaction validation is enabled, yet "+
			"block processing will stop after %v encountered issues. (0 is endless)", v.cfg.MaxNumErrors)
	}

	return nil
}

func (v *stateDbValidator) runPreTxValidation(tool string, db state.VmStateDB, state executor.State[txcontext.TxContext], errOutput chan error) error {
	if !v.target.WorldState {
		return nil
	}

	if v.cfg.OverwritePreWorldState {
		return overwriteWorldState(v.cfg, state.Data.GetInputState(), db)
	}
	err := validateWorldState(v.cfg, db, state.Data.GetInputState(), v.log)
	if err == nil {
		return nil
	}

	err = fmt.Errorf("%v err:\nblock %v tx %v\n world-state input is not contained in the state-db\n %v\n", tool, state.Block, state.Transaction, err)

	if v.isErrFatal(err, errOutput) {
		return err
	}

	return nil
}

func (v *stateDbValidator) runPostTxValidation(tool string, db state.VmStateDB, state0 executor.State[txcontext.TxContext], res txcontext.Result, errOutput chan error) error {
	if v.target.WorldState {
		if err := validateWorldState(v.cfg, db, state0.Data.GetOutputState(), v.log); err != nil {
			err = fmt.Errorf("%v err:\nworld-state0 output error at block %v tx %v; %v", tool, state0.Block, state0.Transaction, err)
			if v.isErrFatal(err, errOutput) {
				return err
			}
		}

		err := doDbValidation(db, state0.Data.GetOutputState())
		if err != nil {
			return err
		}
	}

	// ethereumLfvmBlockExceptions needs to skip receipt validation
	_, skipEthereumException := ethereumLfvmBlockExceptions[state0.Block]
	if skipEthereumException {
		// skip should only happen if we are on Ethereum chain and using lfvm
		skipEthereumException = v.cfg.VmImpl == "lfvm" && v.cfg.ChainID == utils.EthereumChainID
	}

	// TODO remove state0.Transaction < 99999 after patch aida-db
	if v.target.Receipt && state0.Transaction < utils.PseudoTx && !skipEthereumException {
		if err := v.validateReceipt(res.GetReceipt(), state0.Data.GetResult().GetReceipt()); err != nil {
			err = fmt.Errorf("%v err:\nvm-result error at block %v tx %v; %v", tool, state0.Block, state0.Transaction, err)
			if v.isErrFatal(err, errOutput) {
				return err
			}
		}
	}

	return nil
}

func doDbValidation(db state.VmStateDB, alloc2 txcontext.WorldState) error {
	gethDb, ok := db.(*state.GethStateDB)
	if !ok {
		return nil
	}

	generateDbPostAlloc(gethDb)
	alloc := gethDb.SubstatePostAlloc

	var err string

	accountCount := 0
	alloc2.ForEachAccount(func(addr common.Address, acc txcontext.Account) {
		accountCount++
		acc2, ok := alloc[stypes.Address(addr)]
		if !ok {
			err += fmt.Sprintf("  Account %v does not exist\n", addr.Hex())
			return
		}

		accBalance := acc.GetBalance()

		if accBalance.ToBig().Cmp(acc2.Balance) != 0 {
			err += fmt.Sprintf("  Failed to validate balance for account %v\n"+
				"    have %v\n"+
				"    want %v\n",
				addr.Hex(), acc2.Balance, accBalance)
		}
		if nonce := acc2.Nonce; nonce != acc.GetNonce() {
			err += fmt.Sprintf("  Failed to validate nonce for account %v\n"+
				"    have %v\n"+
				"    want %v\n",
				addr.Hex(), nonce, acc.GetNonce())
		}
		if code := acc2.Code; bytes.Compare(code, acc.GetCode()) != 0 {
			err += fmt.Sprintf("  Failed to validate code for account %v\n"+
				"    have len %v\n"+
				"    want len %v\n",
				addr.Hex(), len(code), len(acc.GetCode()))
		}

		// validate Storage
		storageCount := 0
		acc.ForEachStorage(func(keyHash common.Hash, valueHash common.Hash) {
			value, _ := acc2.Storage[stypes.Hash(keyHash)]
			if value != stypes.Hash(valueHash) {
				err += fmt.Sprintf("  Failed to validate storage for account %v, key %v\n"+
					"    have %v\n"+
					"    want %v\n",
					addr.Hex(), keyHash.Hex(), value, valueHash.Hex())
			}
			storageCount++
		})

		if storageCount != len(acc2.Storage) {
			err += fmt.Sprintf("  Failed to validate storage count for account %v\n"+
				"    have %v\n"+
				"    want %v\n",
				addr.Hex(), storageCount, len(acc2.Storage))
		}
	})

	if accountCount != alloc2.Len() {
		err += fmt.Sprintf("  Failed to validate account count\n"+
			"    have %v\n"+
			"    want %v\n",
			accountCount, alloc2.Len())
	}

	if len(err) > 0 {
		return fmt.Errorf(err)
	}
	return nil
}

func generateDbPostAlloc(gethDb *state.GethStateDB) {
	dirtyAddresses := make(map[cc.Address]struct{})

	// copy original storage values to Prestate and Poststate
	for addr, sa := range gethDb.SubstatePreAlloc {
		if sa == nil {
			dirtyAddresses[cc.Address(addr)] = struct{}{}
			delete(gethDb.SubstatePreAlloc, addr)
			continue
		}

		comAddr := common.BytesToAddress(addr.Bytes())
		ac, found := gethDb.AccessedStorage[comAddr]
		if found {
			for key := range ac {
				value := gethDb.GetCommittedState(comAddr, key)

				//if value != valueM {
				//	panic("value mismatch")
				//}

				sa.Storage[stypes.Hash(key)] = stypes.Hash(value)

			}
		}
		gethDb.SubstatePostAlloc[addr] = sa.Copy()
	}

	for address := range dirtyAddresses {
		if gethDb.Db.Exist(common.Address(address)) {
			s := make(map[stypes.Hash]stypes.Hash)
			for key := range gethDb.AccessedStorage[common.Address(address)] {
				s[stypes.Hash(key)] = stypes.Hash{}
			}
			gethDb.SubstatePostAlloc[stypes.Address(address)] = &substate.Account{Storage: s}
		}
	}

	gethDb.AccessedStorage = nil

	toDelete := make([]stypes.Address, 0)
	for address, acc := range gethDb.SubstatePostAlloc {
		if gethDb.Db.HasSelfDestructed(common.Address(address)) {
			toDelete = append(toDelete, address)
			continue
		}

		// update the account in StateDB.SubstatePostAlloc
		acc.Balance = gethDb.Db.GetBalance(common.Address(address)).ToBig()
		acc.Nonce = gethDb.Db.GetNonce(common.Address(address))
		acc.Code = gethDb.Db.GetCode(common.Address(address))
		storageToUpdate := make(map[stypes.Hash]stypes.Hash)
		for key := range acc.Storage {
			storageToUpdate[key] = stypes.Hash(gethDb.Db.GetState(common.Address(address), common.Hash(key)))
		}
		acc.Storage = storageToUpdate
	}

	for _, address := range toDelete {
		delete(gethDb.SubstatePostAlloc, address)
	}
}

// isErrFatal decides whether given error should stop the program or not depending on ContinueOnFailure and MaxNumErrors.
func (v *stateDbValidator) isErrFatal(err error, ch chan error) bool {
	// ContinueOnFailure is disabled, return the error and exit the program
	if !v.cfg.ContinueOnFailure {
		return true
	}

	ch <- err
	v.numberOfErrors.Add(1)

	// endless run
	if v.cfg.MaxNumErrors == 0 {
		return false
	}

	// too many errors
	if int(v.numberOfErrors.Load()) >= v.cfg.MaxNumErrors {
		return true
	}

	return false
}

// validateReceipt compares result from vm against the expected one.
// Error is returned if any mismatch is found.
func (v *stateDbValidator) validateReceipt(got, want txcontext.Receipt) error {
	if !got.Equal(want) {
		return fmt.Errorf(
			"\ngot:\n"+
				"\tstatus: %v\n"+
				"\tbloom: %v\n"+
				"\tlogs: %v\n"+
				"\tcontract address: %v\n"+
				"\tgas used: %v\n"+
				"\nwant:\n"+
				"\tstatus: %v\n"+
				"\tbloom: %v\n"+
				"\tlogs: %v\n"+
				"\tcontract address: %v\n"+
				"\tgas used: %v\n",
			got.GetStatus(),
			got.GetBloom().Big().Uint64(),
			got.GetLogs(),
			got.GetContractAddress(),
			got.GetGasUsed(),
			want.GetStatus(),
			want.GetBloom().Big().Uint64(),
			want.GetLogs(),
			want.GetContractAddress(),
			want.GetGasUsed())
	}

	return nil
}
