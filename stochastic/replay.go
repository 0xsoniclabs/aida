package stochastic

import (
	"encoding/binary"
	"fmt"
	"log"
	"math/big"
	"math/rand"

	"github.com/Fantom-foundation/Aida/state"
	"github.com/ethereum/go-ethereum/common"
)

// Simulation constants
// TODO: convert constants to CLI parameters so that they can be changed without recompiling.
const (
	AddBalanceRange = 100000  // balance range for adding value to an account
	SetNonceRange   = 1000000 // nonce range
	MaxCodeSize     = 24576   // fixed upper limit by EIP-170
	FinaliseFlag    = true    // flag for Finalise() StateDB operation
)

// stochasticAccount keeps necessary account information for the simulation in memory
type stochasticAccount struct {
	balance     int64 // current balance of account
	hasSuicided bool  // flag whether suicide has been invoked
}

// stochasticState keeps the execution state for the stochastic simulation
type stochasticState struct {
	db             state.StateDB                // StateDB database
	contracts      *IndirectAccess              // index access generator for contracts
	keys           *RandomAccess                // index access generator for keys
	values         *RandomAccess                // index access generator for values
	snapshotLambda float64                      // lambda parameter for snapshot delta distribution
	txNum          uint32                       // current transaction number
	blockNum       uint64                       // current block number
	epochNum       uint64                       // current epoch number
	snapshot       []int                        // stack of active snapshots
	accounts       map[int64]*stochasticAccount // account information using address index as key
	balanceLog     map[int64][]int64            // balance log keeping track of balances for snapshots
	verbose        bool                         // verbose flag
}

// RunStochasticReplay runs the stochastic simulation for StateDB operations.
// It requires the simulation model and simulation length. The verbose enables/disables
// the printing of StateDB operations and their arguments on the screen.
func RunStochasticReplay(db state.StateDB, e *EstimationModelJSON, simLength int, verbose bool) {

	// retrieve operations and stochastic matrix from simulation object
	operations := e.Operations
	A := e.StochasticMatrix

	// produce random access generators for contract addresses,
	// storage-keys, and storage addresses.
	// (NB: Contracts need an indirect access wrapper because
	// contract addresses can be deleted by suicide.)
	contracts := NewIndirectAccess(NewRandomAccess(
		e.Contracts.NumKeys,
		e.Contracts.Lambda,
		e.Contracts.QueueDistribution,
	))
	keys := NewRandomAccess(
		e.Keys.NumKeys,
		e.Keys.Lambda,
		e.Keys.QueueDistribution,
	)
	values := NewRandomAccess(
		e.Values.NumKeys,
		e.Values.Lambda,
		e.Values.QueueDistribution,
	)

	// setup state
	ss := NewStochasticState(db, contracts, keys, values, e.SnapshotLambda, verbose)

	// create accounts in StateDB
	ss.prime()

	// set initial state to BeginEpoch
	state := initialState(operations, "BE")
	if state == -1 {
		panic("BeginEpoch cannot be observed in stochastic matrix/recording failed.")
	}

	blocks := 0
	for {

		// decode opcode
		op, addrCl, keyCl, valueCl := DecodeOpcode(operations[state])

		// execute operation with its argument classes
		ss.execute(op, addrCl, keyCl, valueCl)

		// check for end of simulation
		if op == EndBlockID {
			blocks++
			if blocks >= simLength {
				break
			}
		}

		// transit to next state in Markovian process
		state = nextState(A, state)
	}
}

// NewStochasticState creates a new state for execution StateDB operations
func NewStochasticState(db state.StateDB, contracts *IndirectAccess, keys *RandomAccess, values *RandomAccess, snapshotLambda float64, verbose bool) stochasticState {

	// retrieve number of contracts
	n := contracts.NumElem()

	// initialise accounts in memory with balances greater than zero
	accounts := make(map[int64]*stochasticAccount, n+1)
	for i := int64(0); i <= n; i++ {
		accounts[i] = &stochasticAccount{
			balance:     rand.Int63n(AddBalanceRange),
			hasSuicided: false,
		}
	}

	// return stochastic state
	return stochasticState{
		db:             db,
		accounts:       accounts,
		contracts:      contracts,
		keys:           keys,
		values:         values,
		snapshotLambda: snapshotLambda,
		verbose:        verbose,
		balanceLog:     make(map[int64][]int64),
	}
}

// prime StateDB accounts using account information
func (ss *stochasticState) prime() {
	db := ss.db
	for addrIdx, detail := range ss.accounts {
		addr := toAddress(addrIdx)
		db.CreateAccount(addr)
		if detail.balance > 0 {
			db.AddBalance(addr, big.NewInt(detail.balance))
		}
	}
}

// execute StateDB operations on a stochastic state.
func (ss *stochasticState) execute(op int, addrCl int, keyCl int, valueCl int) {
	var (
		addr  common.Address
		key   common.Hash
		value common.Hash
		db    state.StateDB = ss.db
	)

	// fetch indexes from index access generators
	addrIdx := ss.contracts.NextIndex(addrCl)
	keyIdx := ss.keys.NextIndex(keyCl)
	valueIdx := ss.values.NextIndex(valueCl)

	// convert index to address/hashes
	if addrCl != noArgID {
		if addrCl == newValueID {
			// create a new internal representation of an account
			// but don't create an account in StateDB; this is done
			// by CreateAccount.
			ss.accounts[addrIdx] = &stochasticAccount{
				balance:     0,
				hasSuicided: false,
			}
		}
		addr = toAddress(addrIdx)
	}
	if keyCl != noArgID {
		key = toHash(keyIdx)
	}
	if valueCl != noArgID {
		value = toHash(valueIdx)
	}

	// print opcode and its arguments
	if ss.verbose {
		// print operation
		fmt.Printf("opcode:%v (%v)", opText[op], EncodeOpcode(op, addrCl, keyCl, valueCl))

		// print indexes of contract address, storage key, and storage value.
		if addrCl != noArgID {
			fmt.Printf(" addr-idx: %v", addrIdx)
		}
		if keyCl != noArgID {
			fmt.Printf(" key-idx: %v", keyIdx)
		}
		if valueCl != noArgID {
			fmt.Printf(" value-idx: %v", valueIdx)
		}
	}

	switch op {
	case AddBalanceID:
		value := rand.Int63n(AddBalanceRange)
		if ss.verbose {
			fmt.Printf(" value: %v", value)
		}
		ss.updateBalanceLog(addrIdx, value)
		db.AddBalance(addr, big.NewInt(value))

	case BeginBlockID:
		if ss.verbose {
			fmt.Printf(" id: %v", ss.blockNum)
		}
		db.BeginBlock(ss.blockNum)
		ss.txNum = 0

	case BeginEpochID:
		if ss.verbose {
			fmt.Printf(" id: %v", ss.epochNum)
		}
		db.BeginEpoch(ss.epochNum)

	case BeginTransactionID:
		if ss.verbose {
			fmt.Printf(" id: %v", ss.txNum)
		}
		db.BeginTransaction(ss.txNum)
		ss.snapshot = []int{}

	case CreateAccountID:
		db.CreateAccount(addr)

	case EmptyID:
		db.Empty(addr)

	case EndBlockID:
		db.EndBlock()
		ss.blockNum++

	case EndEpochID:
		db.EndEpoch()
		ss.epochNum++

	case EndTransactionID:
		db.EndTransaction()
		ss.txNum++
		ss.commitBalanceLog()

	case ExistID:
		db.Exist(addr)

	case FinaliseID:
		db.Finalise(FinaliseFlag)
		ss.deleteAccounts()

	case GetBalanceID:
		db.GetBalance(addr)

	case GetCodeHashID:
		db.GetCodeHash(addr)

	case GetCodeID:
		db.GetCode(addr)

	case GetCodeSizeID:
		db.GetCodeSize(addr)

	case GetCommittedStateID:
		db.GetCommittedState(addr, key)

	case GetNonceID:
		db.GetNonce(addr)

	case GetStateID:
		db.GetState(addr, key)

	case HasSuicidedID:
		db.HasSuicided(addr)

	case RevertToSnapshotID:
		snapshotNum := len(ss.snapshot)
		if snapshotNum > 0 {
			// TODO: consider a more realistic distribution
			// rather than the uniform distribution.
			snapshotIdx := snapshotNum - int(randIndex(ss.snapshotLambda, int64(snapshotNum))) - 1
			snapshot := ss.snapshot[snapshotIdx]
			if ss.verbose {
				fmt.Printf(" id: %v", snapshot)
			}
			db.RevertToSnapshot(snapshot)

			// update active snapshots and perform a rollback in balance log
			ss.snapshot = ss.snapshot[0:snapshotIdx]
			ss.rollbackBalanceLog(snapshotIdx)
		}

	case SetCodeID:
		sz := rand.Intn(MaxCodeSize-1) + 1
		if ss.verbose {
			fmt.Printf(" code-size: %v", sz)
		}
		code := make([]byte, sz)
		_, err := rand.Read(code)
		if err != nil {
			log.Fatalf("error producing a random byte slice. Error: %v", err)
		}
		db.SetCode(addr, code)

	case SetNonceID:
		value := uint64(rand.Intn(SetNonceRange))
		db.SetNonce(addr, value)

	case SetStateID:
		db.SetState(addr, key, value)

	case SnapshotID:
		id := db.Snapshot()
		if ss.verbose {
			fmt.Printf(" id: %v", id)
		}
		ss.snapshot = append(ss.snapshot, id)

	case SubBalanceID:
		balance := ss.getBalanceLog(addrIdx)
		if balance > 0 {
			// get a delta that does not exceed current balance
			// in the current snapshot
			value := rand.Int63n(balance)
			if ss.verbose {
				fmt.Printf(" value: %v", value)
			}
			db.SubBalance(addr, big.NewInt(value))
			ss.updateBalanceLog(addrIdx, -value)
		}

	case SuicideID:
		db.Suicide(addr)
		ss.accounts[addrIdx].hasSuicided = true

	default:
		panic("invalid operation")
	}
	if ss.verbose {
		fmt.Println()
	}
}

// initialState returns the row/column index of the first state in the stochastic matrix.
func initialState(operations []string, opcode string) int {
	for i, opc := range operations {
		if opc == opcode {
			return i
		}
	}
	return -1
}

// nextState produces the next state in the Markovian process.
func nextState(A [][]float64, i int) int {
	// Retrieve a random number in [0,1.0).
	r := rand.Float64()

	// Use Kahan's sum for summing values
	// in case we have a combination of very small
	// and very large values.
	sum := float64(0.0)
	c := float64(0.0)
	k := -1
	for j := 0; j < len(A); j++ {
		y := A[i][j] - c
		t := sum + y
		c = (t - sum) - y
		sum = t
		if r <= sum {
			return j
		}
		// If we have a numerical unstable cumulative
		// distribution (large and small numbers that cancel
		// each other out when summing up), we can take the last
		// non-zero entry as a solution. It also detects
		// stochastic matrices with a row whose row
		// sum is not zero (return value is -1 for such a case).
		if A[i][j] > 0.0 {
			k = j
		}
	}
	return k
}

// toAddress converts an address index to a contract address.
// TODO: Improve encoding so that index conversion becomes sparse.
func toAddress(idx int64) common.Address {
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

// toHash converts a key/value index to a hash
func toHash(idx int64) common.Hash {
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

// getBalanceLog computes the actual balance for the current snapshot
func (ss *stochasticState) getBalanceLog(addrIdx int64) int64 {
	balance := ss.accounts[addrIdx].balance
	for _, v := range ss.balanceLog[addrIdx] {
		balance += v
	}
	return balance
}

// updateBalanceLog adds a delta balance for an contract for the current snapshot.
func (ss *stochasticState) updateBalanceLog(addrIdx int64, delta int64) {
	snapshotNum := len(ss.snapshot) // retrieve number of active snapshots
	if snapshotNum > 0 {
		logLen := len(ss.balanceLog[addrIdx]) // retrieve number of log entries for addrIdx
		if logLen < snapshotNum {
			// fill log entry if too short with zeros
			ss.balanceLog[addrIdx] = append(ss.balanceLog[addrIdx], make([]int64, snapshotNum-logLen)...)
		} else if logLen != snapshotNum {
			panic("log wasn't rolled black")
		}
		// update delta of address for current snapshot
		ss.balanceLog[addrIdx][snapshotNum-1] += delta
	} else {
		// if no snapshot exists, just add delta to balance directly
		ss.accounts[addrIdx].balance += delta
	}
}

// commitBalanceLog updates the balances in the account and
// deletes the balance log.
func (ss *stochasticState) commitBalanceLog() {
	// update balances with balance log
	for idx, log := range ss.balanceLog {
		balance := ss.accounts[idx].balance
		for _, value := range log {
			balance += value
		}
		ss.accounts[idx].balance = balance
	}

	// destroy old log for next transaction
	ss.balanceLog = make(map[int64][]int64)
}

// rollbackBalanceLog rollbacks balance log to the k-th snapshot
func (ss *stochasticState) rollbackBalanceLog(k int) {
	// delete deltas of prior snapshots in balance log
	for idx, log := range ss.balanceLog {
		if len(log) > k {
			ss.balanceLog[idx] = ss.balanceLog[idx][0:k]
		}
	}
}

// delete account information when suicide was invoked
func (ss *stochasticState) deleteAccounts() {
	// remove account information when suicide was invoked in the block.
	for addrIdx, detail := range ss.accounts {
		if detail.hasSuicided {
			delete(ss.accounts, addrIdx)
			ss.contracts.DeleteIndex(addrIdx)
		}
	}
}
