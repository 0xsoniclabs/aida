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

package executor

//go:generate mockgen -source executor.go -destination executor_mock.go -package executor

import (
	"errors"
	"fmt"
	"runtime/debug"
	"sync"
	"sync/atomic"

	"github.com/0xsoniclabs/aida/logger"
	"github.com/0xsoniclabs/aida/state"
	"github.com/0xsoniclabs/aida/txcontext"
	"github.com/0xsoniclabs/aida/utils"
	"github.com/0xsoniclabs/substate/db"
)

// ----------------------------------------------------------------------------
//                             Interfaces
// ----------------------------------------------------------------------------

// Executor is an entity coordinating the execution of transactions within a
// requested block range. It implements the decorator pattern, allowing
// extensions to monitor and annotate the execution at various hook-in points.
//
// When running sequentially, the general execution is structured as follows:
//
//	PreRun()
//	for each block {
//	   PreBlock()
//	   for each transaction {
//	       PreTransaction()
//	       Processor.Process(transaction)
//	       PostTransaction()
//	   }
//	   PostBlock()
//	}
//	PostRun()
//
// When running with multiple workers on TransactionLevel granularity, the execution is structures like this:
//
//	PreRun()
//	for transaction in parallel {
//	    PreTransaction()
//	    Processor.Process(transaction)
//	    PostTransaction()
//	}
//	PostRun()
//
// Note that there are no block boundary events in the parallel mode.
//
// When running with multiple workers on BlockLevel granularity, the execution is structures like this:
//
//	PreRun()
//	for block in parallel {
//	   PreBlock()
//	   for each transaction {
//	       PreTransaction()
//	       Processor.Process(transaction)
//	       PostTransaction()
//	   }
//	   PostBlock()
//	}
//	PostRun()
//
// Note that every worker has its own Context so any manipulation with this variable does not need to be thread safe.
//
// Each PreXXX() and PostXXX() is a hook-in point at which extensions may
// track information and/or interfere with the execution. For more details on
// the specific call-backs see the Extension interface below.
type Executor[T any] interface {
	// Run feeds all transactions of the given block range [from,to) to the
	// provided processor and performs the needed call-backs on the provided
	// extensions. If a processor or an extension returns an error, execution
	// stops with the reported error.
	// PreXXX events are delivered to the extensions in the given order, while
	// PostXXX events are delivered in reverse order. If any of the extensions
	// reports an error during processing of an event, the same event is still
	// delivered to the remaining extensions before processing is aborted.
	Run(params Params, processor Processor[T], extensions []Extension[T], aidaDb db.BaseDB) error
}

// NewExecutor creates a new executor based on the given provider.
func NewExecutor[T any](provider Provider[T], logLevel string) Executor[T] {
	return newExecutor[T](provider, logger.NewLogger(logLevel, "Executor"))
}

func newExecutor[T any](provider Provider[T], log logger.Logger) Executor[T] {
	return &executor[T]{
		provider: provider,
		log:      log,
	}
}

// ParallelismGranularity determines isolation level if same archive is kept for all transactions in block or for each is created new one
type ParallelismGranularity byte

const (
	TransactionLevel ParallelismGranularity = iota // Post and Pre Transactions() need to be Thread-Safe
	BlockLevel
)

// Params summarizes input parameters for a run of the executor.
type Params struct {
	// From is the beginning of the range of blocks to be processed (inclusive).
	From int
	// To is the end of the range of blocks to be processed (exclusive).
	To int
	// State is an optional StateDB instance to be made available to the
	// processor and extensions during execution.
	State state.StateDB
	// NumWorkers is the number of concurrent goroutines to be used to
	// process blocks. If the number of workers is 1, transactions are
	// guranteed to be processed in-order. If it is > 1 no fixed order
	// is guranteed. Any number <= 1 is considered to be 1, thus the default
	// value of 0 is valid.
	NumWorkers int
	// ParallelismGranularity determines whether parallelism is done on block or transaction level
	ParallelismGranularity ParallelismGranularity
}

// Processor is an interface for the entity to which an executor is feeding
// transactions to.
type Processor[T any] interface {
	// Process is called on each transaction in the range of blocks covered
	// by an Executor run. When running with multiple workers, the Process
	// function is required to be thread safe.
	Process(State[T], *Context) error
}

// Extension is an interface for modulare annotations to the execution of
// a range of transactions. During various stages, methods of extensions are
// called, enabling them to monitor and/or interfere with the execution.
// Since blocks may be processed in parallel, callbacks are generally
// required to be thread safe (with the exception of the Pre-/ and PostRun)
// callback.
type Extension[T any] interface {
	// PreRun is called before the begin of the execution of a block range,
	// even if the range is empty. The provided state lists the initial block
	// of the range. For every run, PreRun is only called once, before any
	// other call-back. If an error is reported, execution will abort after
	// PreRun has been called on all registered Extensions.
	PreRun(State[T], *Context) error

	// PostRun is guranteed to be called at the end of each execution. An
	// execution may end successfully, if no exception has been produced by
	// the Processor or any Extension, or in a failure state, if errors
	// have been produced. In case of a successful execution, the provided
	// state lists the first non-executiond block, while in an error case
	// it references the last transaction attempted to be processed. Also,
	// the second parameter contains the error causing the abort.
	PostRun(State[T], *Context, error) error

	// PreBlock is called once before the begin of processing a block with
	// the state containing the number of the Block. This function is not
	// called when running with multiple workers.
	PreBlock(State[T], *Context) error

	// PostBlock is called once after the end of processing a block with
	// the state containing the number of the Block and the last transaction
	// processed in the block. This function is not called when running with
	// multiple workers.
	PostBlock(State[T], *Context) error

	// PreTransaction is called once before each transaction with the state
	// listing the block number, the transaction number, and the substate data
	// providing the input for the subsequent execution of the transaction.
	// When running with multiple workers, this function may be called
	// concurrently, and must thus be thread safe.
	PreTransaction(State[T], *Context) error

	// PostTransaction is called once after each transaction with the state
	// listing the block number, the transaction number, and the substate data
	// providing the input for the subsequent execution of the transaction.
	// When running with multiple workers, this function may be called
	// concurrently, and must thus be thread safe.
	PostTransaction(State[T], *Context) error
}

// State summarizes the current state of an execution and is passed to
// Processors and Extensions as an input for their actions.
type State[T any] struct {
	// Block the current block number, valid for all call-backs.
	Block int

	// Transaction is the transaction number of the current transaction within
	// its respective block. It is only valid for PreTransaction, PostTransaction,
	// PostBlock, and for PostRun events in case of an abort.
	Transaction int

	// Data is the input required for processing the current transaction. It is
	// only valid for Pre- and PostTransaction events.
	Data T
}

// Context summarizes context data for the current execution and is passed
// as a mutable object to Processors and Extensions. Either max decide to
// modify its content to implement their respective features.
type Context struct {
	// State is an optional StateDB instance manipulated during by the processor
	// and extensions of a block-range execution.
	State state.StateDB

	// Archive represents a historical State
	Archive state.NonCommittableStateDB

	// StateDbPath contains path to working stateDb directory
	StateDbPath string

	// AidaDb is an optional LevelDb readonly database containing data for testing StateDb (i.e. state hashes).
	AidaDb db.BaseDB

	// ErrorInput is used if continue-on-failure is enabled or if log-file is definer so that at the end
	// of the run, we log all errors into a file. This chanel should be only used for processing errors,
	// hence any non-fatal errors. Any fatal should still be returned so that the app ends.
	ErrorInput chan error

	// ExecutionResult is set after the execution.
	// It is used for validation and gas measurements.
	ExecutionResult txcontext.Result
}

// ----------------------------------------------------------------------------
//                               Implementations
// ----------------------------------------------------------------------------

type executor[T any] struct {
	provider Provider[T]
	log      logger.Logger
}

func (e *executor[T]) Run(params Params, processor Processor[T], extensions []Extension[T], aidaDb db.BaseDB) (err error) {
	state := State[T]{}
	ctx := Context{State: params.State, AidaDb: aidaDb}

	defer func() {
		// Skip PostRun actions if a panic occurred. In such a case there is no guarantee
		// on the state of anything, and PostRun operations may deadlock or cause damage.
		if r := recover(); r != nil {
			msg := fmt.Sprintf("%v\n%s", r, string(debug.Stack()))
			panic(msg)
		}
		err = errors.Join(
			err,
			signalPostRun(state, &ctx, err, extensions),
		)
	}()

	state.Block = params.From
	if err = signalPreRun(state, &ctx, extensions); err != nil {
		return err
	}

	if params.NumWorkers <= 1 {
		params.NumWorkers = 1
	}

	switch params.ParallelismGranularity {
	case TransactionLevel:
		return e.runTransactions(params, processor, extensions, &state, &ctx)
	case BlockLevel:
		return e.runBlocks(params, processor, extensions, &state, &ctx)
	default:
		return fmt.Errorf("incorrect parallelism type: %v", params.ParallelismGranularity)
	}
}

// runBlock runs transaction execution in a block
func runBlock[T any](
	workerNumber int,
	blocks chan []*TransactionInfo[T],
	wg *sync.WaitGroup,
	abort utils.Event,
	workerErrs []error,
	processor Processor[T],
	extensions []Extension[T],
	ctx *Context,
	cachedPanic *atomic.Value,
) {

	// channel panics back to the main thread.
	defer func() {
		if r := recover(); r != nil {
			abort.Signal() // stop forwarder and other workers too
			msg := fmt.Sprintf("worker %v recovered panic; %v\n%s", workerNumber, r, string(debug.Stack()))
			cachedPanic.Store(msg)
		}
		wg.Done()
	}()

	var localState State[T]
	for {
		select {
		case blockTransactions := <-blocks:
			if blockTransactions == nil || len(blockTransactions) == 0 {
				return // reached an end without abort
			}

			localState.Block = blockTransactions[0].Block
			localState.Data = blockTransactions[0].Data
			localCtx := *ctx

			if err := signalPreBlock(localState, &localCtx, extensions); err != nil {
				workerErrs[workerNumber] = err
				abort.Signal()
				return
			}

			for _, tx := range blockTransactions {
				localState.Data = tx.Data
				localState.Transaction = tx.Transaction

				if err := runTransaction(localState, &localCtx, tx.Data, processor, extensions); err != nil {
					workerErrs[workerNumber] = err
					abort.Signal()
					return
				}

				// listen for possible abort between the transactions
				select {
				case <-abort.Wait():
					return
				default:
					continue
				}
			}

			if err := signalPostBlock(localState, &localCtx, extensions); err != nil {
				workerErrs[workerNumber] = err
				abort.Signal()
				return
			}
		case <-abort.Wait():
			return
		}
	}
}

// forwardBlocks is a worker that unites transactions by block and forwards them to execution.
func (e *executor[T]) forwardBlocks(params Params, abort utils.Event) (chan []*TransactionInfo[T], *atomic.Pointer[error]) {
	blocks := make(chan []*TransactionInfo[T], 10*params.NumWorkers)
	forwardErr := new(atomic.Pointer[error])

	go func() {
		defer close(blocks)
		abortErr := errors.New("aborted")

		previousBlock := params.From
		first := true

		block := make([]*TransactionInfo[T], 0)
		err := e.provider.Run(params.From, params.To, func(tx TransactionInfo[T]) error {
			if first {
				previousBlock = tx.Block
				first = false
			}

			if tx.Block != previousBlock {
				previousBlock = tx.Block
				select {
				case blocks <- block:
					// clean block for reuse
					block = make([]*TransactionInfo[T], 0)
				case <-abort.Wait():
					return abortErr
				}
			}

			block = append(block, &tx)

			return nil
		})

		// send last block to the queue
		if err == nil {
			select {
			case blocks <- block:
			case <-abort.Wait():
				err = abortErr
			}
		}

		if err != abortErr {
			forwardErr.Store(&err)
		}
	}()

	return blocks, forwardErr
}

func (e *executor[T]) runTransactions(params Params, processor Processor[T], extensions []Extension[T], state *State[T], ctx *Context) error {
	numWorkers := params.NumWorkers

	// An event for signaling an abort of the execution.
	abort := utils.MakeEvent()

	var wg sync.WaitGroup
	// Start one go-routine forwarding transactions from the provider to a local channel.
	var forwardErr error
	transactions := make(chan *TransactionInfo[T], 10*numWorkers)
	wg.Add(1)
	go func() {
		defer func() {
			close(transactions)
			wg.Done()
		}()
		abortErr := errors.New("aborted")
		err := e.provider.Run(params.From, params.To, func(tx TransactionInfo[T]) error {
			select {
			case transactions <- &tx:
				return nil
			case <-abort.Wait():
				return abortErr
			}
		})
		if err != abortErr {
			forwardErr = err
		}
	}()

	// Start numWorkers go-routines processing transactions in parallel.
	var cachedPanic atomic.Value

	wg.Add(numWorkers)
	workerErrs := make([]error, numWorkers)
	e.log.Debugf("Starting %v workers run on Transaction granularity...", numWorkers)
	for i := 0; i < numWorkers; i++ {
		go func(i int) {
			// channel panics back to the main thread.
			defer func() {
				if r := recover(); r != nil {
					abort.Signal() // stop forwarder and other workers too
					msg := fmt.Sprintf("worker %v recovered panic; %v\n%s", i, r, string(debug.Stack()))
					cachedPanic.Store(msg)
				}
				wg.Done()
			}()
			for {
				select {
				case tx := <-transactions:
					if tx == nil {
						return // reached an end without abort
					}
					localState := *state
					localState.Block = tx.Block
					localState.Transaction = tx.Transaction
					localCtx := *ctx
					if err := runTransaction(localState, &localCtx, tx.Data, processor, extensions); err != nil {
						workerErrs[i] = err
						abort.Signal()
						return
					}
				case <-abort.Wait():
					return
				}
			}
		}(i)
	}

	wg.Wait()

	if r := cachedPanic.Load(); r != nil {
		panic(r)
	}

	err := errors.Join(
		forwardErr,
		errors.Join(workerErrs...),
	)
	if err == nil {
		state.Block = params.To
	}
	return err
}

func runTransaction[T any](state State[T], ctx *Context, data T, processor Processor[T], extensions []Extension[T]) error {
	state.Data = data
	if err := signalPreTransaction(state, ctx, extensions); err != nil {
		return err
	}
	if err := processor.Process(state, ctx); err != nil {
		return err
	}
	if err := signalPostTransaction(state, ctx, extensions); err != nil {
		return err
	}
	return nil
}
func (e *executor[T]) runBlocks(params Params, processor Processor[T], extensions []Extension[T], state *State[T], ctx *Context) error {
	numWorkers := params.NumWorkers

	// An event for signaling an abort of the execution.
	abort := utils.MakeEvent()

	// Start one go-routine forwarding blocks from the provider to a local channel.
	blocks, forwardErr := e.forwardBlocks(params, abort)

	// Start numWorkers go-routines processing blocks in parallel.
	wg := new(sync.WaitGroup)
	workerErrs := make([]error, numWorkers)

	cachedPanic := new(atomic.Value)

	wg.Add(numWorkers)
	e.log.Debugf("Starting %v workers run on Block granularity...", numWorkers)
	for i := 0; i < numWorkers; i++ {
		go runBlock(i, blocks, wg, abort, workerErrs, processor, extensions, ctx, cachedPanic)
	}

	wg.Wait()

	if r := cachedPanic.Load(); r != nil {
		panic(r)
	}

	err := errors.Join(workerErrs...)
	// append error from blocks forwarding if there is any
	if errPtr := forwardErr.Load(); errPtr != nil {
		err = errors.Join(err, *errPtr)
	}

	if err == nil {
		state.Block = params.To
	}
	return err
}

func RunUtilPrimer[T any](params Params, extensions []Extension[T], aidaDb db.BaseDB) (err error) {
	state := State[T]{}
	ctx := Context{State: params.State, AidaDb: aidaDb}

	state.Block = params.To
	if err = signalPreRun(state, &ctx, extensions); err != nil {
		return err
	}

	return errors.Join(
		err,
		signalPostRun(state, &ctx, err, extensions),
	)
}

func signalPreRun[T any](state State[T], ctx *Context, extensions []Extension[T]) error {
	defer func() {
		if r := recover(); r != nil {
			p := fmt.Sprintf("sending forward recovered panic from PreRun; %v\n%s", r, string(debug.Stack()))
			panic(p)
		}

	}()
	return forEachForward(extensions, func(extension Extension[T]) error {
		return extension.PreRun(state, ctx)
	})
}

func signalPostRun[T any](state State[T], ctx *Context, err error, extensions []Extension[T]) (recoveredPanic error) {
	defer func() {
		if r := recover(); r != nil {
			recoveredPanic = fmt.Errorf("sending forward recovered panic from PostRun; %v\n%s", r, string(debug.Stack()))
			return
		}

	}()
	return forEachBackward(extensions, func(extension Extension[T]) error {
		return extension.PostRun(state, ctx, err)
	})
}

func signalPreBlock[T any](state State[T], ctx *Context, extensions []Extension[T]) error {
	defer func() {
		if r := recover(); r != nil {
			p := fmt.Sprintf("sending forward recovered panic from PreBlock; %v\n%s", r, string(debug.Stack()))
			panic(p)
		}

	}()
	return forEachForward(extensions, func(extension Extension[T]) error {
		return extension.PreBlock(state, ctx)
	})
}

func signalPostBlock[T any](state State[T], ctx *Context, extensions []Extension[T]) error {
	defer func() {
		if r := recover(); r != nil {
			p := fmt.Sprintf("sending forward recovered panic from PostBlock; %v\n%s", r, string(debug.Stack()))
			panic(p)
		}

	}()
	return forEachBackward(extensions, func(extension Extension[T]) error {
		return extension.PostBlock(state, ctx)
	})
}

func signalPreTransaction[T any](state State[T], ctx *Context, extensions []Extension[T]) error {
	defer func() {
		if r := recover(); r != nil {
			p := fmt.Sprintf("sending forward recovered panic from PreTransaction; %v\n%s", r, string(debug.Stack()))
			panic(p)
		}

	}()
	return forEachForward(extensions, func(extension Extension[T]) error {
		return extension.PreTransaction(state, ctx)
	})
}

func signalPostTransaction[T any](state State[T], ctx *Context, extensions []Extension[T]) error {
	defer func() {
		if r := recover(); r != nil {
			p := fmt.Sprintf("sending forward recovered panic from PostTransaction; %v\n%s", r, string(debug.Stack()))
			panic(p)
		}

	}()
	return forEachBackward(extensions, func(extension Extension[T]) error {
		return extension.PostTransaction(state, ctx)
	})
}

func forEachForward[T any](extensions []Extension[T], op func(extension Extension[T]) error) error {
	errs := []error{}
	for _, extension := range extensions {
		if err := op(extension); err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}

func forEachBackward[T any](extensions []Extension[T], op func(extension Extension[T]) error) error {
	errs := []error{}
	for i := len(extensions) - 1; i >= 0; i-- {
		if err := op(extensions[i]); err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}
