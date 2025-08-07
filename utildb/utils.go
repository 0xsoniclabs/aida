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

package utildb

import (
	"bufio"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"math/big"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/0xsoniclabs/aida/logger"
	"github.com/0xsoniclabs/aida/utils"
	"github.com/0xsoniclabs/substate/db"
	"github.com/0xsoniclabs/substate/substate"
	"github.com/0xsoniclabs/substate/types"
	"github.com/0xsoniclabs/substate/updateset"
	"github.com/Fantom-foundation/lachesis-base/common/bigendian"
	"github.com/holiman/uint256"
	"github.com/op/go-logging"
	"github.com/stretchr/testify/assert"
)

const commandOutputLimit = 50

// OpenSourceDatabases opens all databases required for merge
func OpenSourceDatabases(sourceDbPaths []string) ([]db.BaseDB, error) {
	if len(sourceDbPaths) < 1 {
		return nil, fmt.Errorf("no source database were specified\n")
	}

	var sourceDbs []db.BaseDB
	for i := 0; i < len(sourceDbPaths); i++ {
		path := sourceDbPaths[i]
		_, err := os.Stat(path)
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("source database %s; doesn't exist\n", path)
		}
		db, err := db.NewReadOnlyBaseDB(path)
		if err != nil {
			return nil, fmt.Errorf("source database %s; error: %v", path, err)
		}
		sourceDbs = append(sourceDbs, db)
	}

	return sourceDbs, nil
}

// MustCloseDB close database safely
func MustCloseDB(db db.BaseDB) {
	if db != nil {
		err := db.Close()
		if err != nil {
			if err.Error() != "leveldb: closed" {
				fmt.Printf("could not close database; %s\n", err.Error())
			}
		}
	}
}

// runCommand wraps cmd execution to distinguish whether to display its output
func runCommand(cmd *exec.Cmd, resultChan chan string, stopChan chan struct{}, log logger.Logger) error {
	if resultChan != nil {
		defer close(resultChan)
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("unable to create StdoutPipe; %v", err)
	}
	defer stdout.Close()
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("unable to create StderrPipe; %v", err)
	}
	defer stderr.Close()

	err = cmd.Start()
	if err != nil {
		return fmt.Errorf("unable to start Command %v; %v", cmd, err)
	}

	merged := io.MultiReader(stderr, stdout)
	scanner := bufio.NewScanner(merged)

	lastOutputMessagesChan := make(chan string, commandOutputLimit)

	// scannedChan to relay command output into channel to be able to select with stopChan
	scannedChan := make(chan string)
	go func() {
		for scanner.Scan() {
			scannedChan <- scanner.Text()
		}
		close(scannedChan)
	}()

	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
	}()

	// this command expects possibility to be stopped by kill signal from aida
	for {
		select {
		case <-stopChan:
			// not returning any error other than from failure of kill signal,
			// because the command was terminated by aida intentionally
			return killCommand(cmd, log, done)
		case m, ok := <-scannedChan:
			if ok {
				processScannedCommandOutput(m, resultChan, log, lastOutputMessagesChan)
				break
			}

			close(lastOutputMessagesChan)

			// wait until command finishes or stopSignal is received
			select {
			case <-stopChan:
				return killCommand(cmd, log, done)
			case res, ok := <-done:
				return processCommandResult(res, ok, scanner, lastOutputMessagesChan, resultChan, cmd, log)
			}
		}
	}
}

// killCommand terminates command gracefully first and then forcefully
func killCommand(cmd *exec.Cmd, log logger.Logger, done chan error) error {
	// A stop signal was received; terminate the command.
	// Attempting to interrupt command gracefully first.
	// Create a timeout with a 1-minute duration.
	timeout := time.NewTimer(time.Minute)
	err := cmd.Process.Signal(syscall.SIGINT)
	if err != nil {
		// might be just race condition when process already finished
		log.Warningf("unable to send SIGINT to Command %v; %v", cmd, err)
	}

	select {
	case <-done:
		log.Noticef("Command %v terminated gracefully", cmd)
	case <-timeout.C:
		// Send a kill signal to the process
		err = cmd.Process.Signal(syscall.SIGKILL)
		if err != nil {
			return fmt.Errorf("unable to send SIGKILL to Command %v; %v", cmd, err)
		}
		// Wait for cmd.Wait() to return after termination.
		<-done
	}
	return nil
}

// processScannedCommandOutput output and send it to resultChan if it is listening and keep lastOutputMessagesChan updated
func processScannedCommandOutput(message string, resultChan chan string, log logger.Logger, lastOutputMessagesChan chan string) {
	if resultChan != nil {
		resultChan <- message
	}
	if log.IsEnabledFor(logging.DEBUG) {
		log.Debug(message)
	} else {
		// in case debugging is turned off and resultChan doesn't listen to output
		// we need to keep most recent output lines in case of error
		if resultChan == nil {
			// throw out the oldest line in case we are at limit
			if len(lastOutputMessagesChan) == commandOutputLimit {
				<-lastOutputMessagesChan
			}
			lastOutputMessagesChan <- message
		}
	}
}

// processCommandResult is used to process command result
func processCommandResult(err error, ok bool, scanner *bufio.Scanner, lastOutputMessagesChan chan string, resultChan chan string, cmd *exec.Cmd, log logger.Logger) error {
	if !ok {
		return fmt.Errorf("unexpected doneChan closed error while executing Command %v; %v", cmd, err)
	}
	// command failed
	if err != nil {
		// print out gathered output since generation failed
		for {
			m, ok := <-lastOutputMessagesChan
			if !ok {
				break
			}
			log.Error(m)
		}

		// read rest of the output - might not be needed
		for scanner.Scan() {
			m := scanner.Text()
			if resultChan != nil {
				resultChan <- m
			}
			log.Error(m)
		}
		return fmt.Errorf("error while executing Command %v; %v", cmd, err)
	}
	return nil
}

// startOperaIpc starts opera node for ipc requests
func startOperaIpc(cfg *utils.Config, stopChan chan struct{}) chan error {
	errChan := make(chan error, 1)

	log := logger.NewLogger(cfg.LogLevel, "Autogen-ipc")
	log.Noticef("Starting opera ipc %v", cfg.ClientDb)

	resChan := make(chan string, 100)
	go func() {
		defer close(errChan)

		//cleanup opera.ipc when node is stopped
		defer func(name string) {
			err := os.Remove(name)
			if !os.IsNotExist(err) && err != nil {
				log.Errorf("failed to remove ipc file %s; %v", name, err)
			}
		}(cfg.ClientDb + "/opera.ipc")

		cmd := exec.Command(getOperaBinary(cfg), "--datadir", cfg.ClientDb, "--maxpeers=0")
		err := runCommand(cmd, resChan, stopChan, log)
		if err != nil {
			errChan <- fmt.Errorf("unable run ipc opera --datadir %v; binary %v; %v", cfg.ClientDb, getOperaBinary(cfg), err)
		}
	}()

	log.Noticef("Waiting for ipc to start")
	errChanParser := make(chan error, 1)

	// wait for ipc to start
	waitDuration := 10 * time.Second
	timer := time.NewTimer(waitDuration)

	err := ipcLoadingProcessWait(resChan, errChan, timer, waitDuration, log)
	if err != nil {
		errChanParser <- err
		close(errChanParser)
		return errChanParser
	}

	go errorRelayer(resChan, errChan, errChanParser)

	return errChanParser
}

// errorRelayer non-blocking error relaying while reading from resChan to prevent deadlock
func errorRelayer(resChan chan string, errChan chan error, errChanParser chan error) {
	defer close(errChanParser)
	for {
		select {
		// since resChan was used the output still needs to be read to prevent deadlock by chan being full
		case <-resChan:
		case err, ok := <-errChan:
			if ok {
				// error happened, the opera failed after ipc initialization
				errChanParser <- fmt.Errorf("opera error after ipc initialization; %v", err)
			}
			return
		}
	}
}

// ipcLoadingProcessWait waits for opera ipc to start and returns error if it didn't start in given time
func ipcLoadingProcessWait(resChan chan string, errChan chan error, timer *time.Timer, waitDuration time.Duration, log logger.Logger) error {
	for {
		select {
		// since resChan was used the output still needs to be read to prevent deadlock by chan being full
		case res, ok := <-resChan:
			if ok {
				// waiting for opera message in output which indicates that ipc is ready for usage
				if strings.Contains(res, "IPC endpoint opened") {
					log.Noticef(res)
					return nil
				}
			}
		case err, ok := <-errChan:
			if ok {
				// errChan closed, this means that stopChan signal was called to terminate opera ipc,
				// which otherwise without an error never stops on its own

				// error happened, the opera ipc didn't start properly
				return fmt.Errorf("opera error during ipc initialization; %v", err)
			}
		case <-timer.C:
			// if ipc didn't start in given time produce an error
			return fmt.Errorf("timeout waiting for opera ipc to start after %s", waitDuration.String())
		}
	}
}

// startOperaRecording records substates
func startOperaRecording(cfg *utils.Config, syncUntilEpoch uint64) chan error {
	errChan := make(chan error, 1)
	// todo check if path to aidaDb exists otherwise create the dir

	log := logger.NewLogger(cfg.LogLevel, "autogen-recording")
	log.Noticef("Starting opera recording %v", cfg.ClientDb)

	go func() {
		defer close(errChan)

		// syncUntilEpoch +1 because command is off by one
		cmd := exec.Command(getOperaBinary(cfg), "--datadir", cfg.ClientDb, "--recording", "--substate-db", cfg.SubstateDb, "--exitwhensynced.epoch", strconv.FormatUint(syncUntilEpoch+1, 10))
		err := runCommand(cmd, nil, nil, log)
		if err != nil {
			errChan <- fmt.Errorf("unable to record opera substates %v; binary %v ; %v", cfg.ClientDb, getOperaBinary(cfg), err)
		}
	}()
	return errChan
}

// getOperaBinary returns path to opera binary
func getOperaBinary(cfg *utils.Config) string {
	var operaBin = "opera"
	if cfg.OperaBinary != "" {
		operaBin = cfg.OperaBinary
	}
	return operaBin
}

// GetDbSize retrieves database size
func GetDbSize(db db.BaseDB) uint64 {
	var count uint64
	iter := db.NewIterator(nil, nil)
	defer iter.Release()
	for iter.Next() {
		count++
	}
	return count
}

// PrintMetadata from given AidaDb
func PrintMetadata(pathToDb string) error {
	log := logger.NewLogger("INFO", "Print-Metadata")
	base, err := db.NewReadOnlyBaseDB(pathToDb)
	if err != nil {
		return err
	}

	md := utils.NewAidaDbMetadata(base, "INFO")

	log.Notice("AIDA-DB INFO:")

	if err = printDbType(md); err != nil {
		return err
	}

	lastBlock := md.GetLastBlock()

	firstBlock := md.GetFirstBlock()

	// CHAIN-ID
	chainID := md.GetChainID()

	if firstBlock == 0 && lastBlock == 0 && chainID == 0 {
		log.Error("your db does not contain metadata; please use metadata generate command")
	} else {
		log.Infof("Chain-ID: %v", chainID)

		// BLOCKS
		log.Infof("First Block: %v", firstBlock)

		log.Infof("Last Block: %v", lastBlock)

		// EPOCHS
		firstEpoch := md.GetFirstEpoch()

		log.Infof("First Epoch: %v", firstEpoch)

		lastEpoch := md.GetLastEpoch()

		log.Infof("Last Epoch: %v", lastEpoch)

		dbHash := md.GetDbHash()

		log.Infof("Db Hash: %v", hex.EncodeToString(dbHash))

		// TIMESTAMP
		timestamp := md.GetTimestamp()

		log.Infof("Created: %v", time.Unix(int64(timestamp), 0))
	}

	// UPDATE-SET
	printUpdateSetInfo(md)

	return nil
}

// printUpdateSetInfo from given AidaDb
func printUpdateSetInfo(m *utils.AidaDbMetadata) {
	log := logger.NewLogger("INFO", "Print-Metadata")

	log.Notice("UPDATE-SET INFO:")

	intervalBytes, err := m.Db.Get([]byte(db.UpdatesetIntervalKey))
	if err != nil {
		log.Warning("Value for update-set interval does not exist in given Dbs metadata")
	} else {
		log.Infof("Interval: %v blocks", bigendian.BytesToUint64(intervalBytes))
	}

	sizeBytes, err := m.Db.Get([]byte(db.UpdatesetSizeKey))
	if err != nil {
		log.Warning("Value for update-set size does not exist in given Dbs metadata")
	} else {
		u := bigendian.BytesToUint64(sizeBytes)

		log.Infof("Size: %.1f MB", float64(u)/float64(1_000_000))
	}
}

// printDbType from given AidaDb
func printDbType(m *utils.AidaDbMetadata) error {
	t := m.GetDbType()

	var typePrint string
	switch t {
	case utils.GenType:
		typePrint = "Generate"
	case utils.CloneType:
		typePrint = "Clone"
	case utils.PatchType:
		typePrint = "Patch"
	case utils.NoType:
		typePrint = "NoType"

	default:
		return errors.New("unknown db type")
	}

	logger.NewLogger("INFO", "Print-Metadata").Noticef("DB-Type: %v", typePrint)

	return nil
}

func GenerateTestAidaDb(t *testing.T) db.BaseDB {
	tmpDir := t.TempDir() + "/testAidaDb"
	database, err := db.NewDefaultBaseDB(tmpDir)
	if err != nil {
		t.Fatalf("error opening stateHash leveldb %s: %v", tmpDir, err)
	}
	md := utils.NewAidaDbMetadata(database, "ERROR")
	err = md.SetAllMetadata(1, 50, 1, 50, 250, []byte("0x0"), 1)
	assert.NoError(t, err)

	// write substates to the database
	substateDb := db.MakeDefaultSubstateDBFromBaseDB(database)
	state := substate.Substate{
		Block:       10,
		Transaction: 7,
		Env: &substate.Env{
			Number:     11,
			Difficulty: big.NewInt(1),
			GasLimit:   uint64(15),
		},
		Message: &substate.Message{
			Value:    big.NewInt(12),
			GasPrice: big.NewInt(14),
		},
		InputSubstate:  substate.WorldState{},
		OutputSubstate: substate.WorldState{},
		Result:         &substate.Result{},
	}

	for i := 0; i < 10; i++ {
		state.Block = uint64(10 + i)
		err = substateDb.PutSubstate(&state)
		if err != nil {
			t.Fatal(err)
		}
	}

	udb := db.MakeDefaultUpdateDBFromBaseDB(database)
	// write update sets to the database
	for i := 1; i <= 10; i++ {
		updateSet := &updateset.UpdateSet{
			WorldState: substate.WorldState{
				types.Address{1}: &substate.Account{
					Nonce:   1,
					Balance: new(uint256.Int).SetUint64(1),
					Code:    []byte{0x01, 0x02},
				},
			},
			Block: uint64(i),
		}
		err = udb.PutUpdateSet(updateSet, []types.Address{})
		if err != nil {
			t.Fatal(err)
		}
	}

	// write delete accounts to the database
	for i := 1; i <= 10; i++ {
		err = database.Put(db.EncodeDestroyedAccountKey(uint64(i), i), []byte("0x1234567812345678123456781234567812345678123456781234567812345678"))
		if err != nil {
			t.Fatal(err)
		}
	}

	// write state hashes to the database
	for i := 11; i <= 20; i++ {
		key := "0x" + strconv.FormatInt(int64(i), 16)
		err = utils.SaveStateRoot(database, key, "0x1234567812345678123456781234567812345678123456781234567812345678")
		if err != nil {
			t.Fatal(err)
		}
	}

	// write block hashes to the database
	for i := 21; i <= 30; i++ {
		key := "0x" + strconv.FormatInt(int64(i), 16)
		err = utils.SaveBlockHash(database, key, "0x1234567812345678123456781234567812345678123456781234567812345678")
		if err != nil {
			t.Fatal(err)
		}
	}

	// write exceptions to the database
	for i := 31; i <= 40; i++ {
		exception := &substate.Exception{
			Block: uint64(i),
			Data: substate.ExceptionBlock{
				PreBlock:  &substate.WorldState{types.Address{0x01}: &substate.Account{Nonce: 1, Balance: uint256.NewInt(100)}},
				PostBlock: &substate.WorldState{types.Address{0x02}: &substate.Account{Nonce: 2, Balance: uint256.NewInt(200)}},
			},
		}
		eDb := db.MakeDefaultExceptionDBFromBaseDB(database)
		err = eDb.PutException(exception)
		if err != nil {
			t.Fatal(err)
		}
	}

	return database
}
