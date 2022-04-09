package worker

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/ardanlabs/blockchain/foundation/blockchain/state"
	"github.com/ardanlabs/blockchain/foundation/blockchain/storage"
)

// miningOperations handles mining.
func (w *Worker) miningOperations() {
	w.evHandler("worker: miningOperations: G started")
	defer w.evHandler("worker: miningOperations: G completed")

	for {
		select {
		case <-w.startMining:
			if !w.isShutdown() {
				w.runMiningOperation()
			}
		case <-w.shut:
			w.evHandler("worker: miningOperations: received shut signal")
			return
		}
	}
}

// runMiningOperation takes all the transactions from the mempool and writes a
// new block to the database.
func (w *Worker) runMiningOperation() {
	w.evHandler("worker: runMiningOperation: MINING: started")
	defer w.evHandler("worker: runMiningOperation: MINING: completed")

	genesis := w.state.RetrieveGenesis()

	// Make sure there are at least transPerBlock in the mempool.
	length := w.state.QueryMempoolLength()
	if length < genesis.TransPerBlock {
		w.evHandler("worker: runMiningOperation: MINING: not enough transactions to mine: Txs[%d]", length)
		return
	}

	// After running a mining operation, check if a new operation should
	// be signaled again.
	defer func() {
		length := w.state.QueryMempoolLength()
		if length >= genesis.TransPerBlock {
			w.evHandler("worker: runMiningOperation: MINING: signal new mining operation: Txs[%d]", length)
			w.SignalStartMining()
		}
	}()

	// If mining is signalled to be cancelled by the WriteNextBlock function,
	// this G can't terminate until it is told it can.
	var wait chan struct{}
	defer func() {
		if wait != nil {
			w.evHandler("worker: runMiningOperation: MINING: termination signal: waiting")
			<-wait
			w.evHandler("worker: runMiningOperation: MINING: termination signal: received")
		}
	}()

	// Drain the cancel mining channel before starting.
	select {
	case <-w.cancelMining:
		w.evHandler("worker: runMiningOperation: MINING: drained cancel channel")
	default:
	}

	// Create a context so mining can be cancelled.
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Can't return from this function until these G's are complete.
	var wg sync.WaitGroup
	wg.Add(2)

	// This G exists to cancel the mining operation.
	go func() {
		defer func() {
			cancel()
			wg.Done()
		}()

		select {
		case wait = <-w.cancelMining:
			w.evHandler("worker: runMiningOperation: MINING: cancel mining requested")
		case <-ctx.Done():
		}
	}()

	// This G is performing the mining.
	go func() {
		defer func() {
			cancel()
			wg.Done()
		}()

		t := time.Now()
		block, err := w.state.MineNewBlock(ctx)
		duration := time.Since(t)

		w.evHandler("worker: runMiningOperation: MINING: mining duration[%v]", duration)

		if err != nil {
			switch {
			case errors.Is(err, state.ErrNotEnoughTransactions):
				w.evHandler("worker: runMiningOperation: MINING: WARNING: not enough transactions in mempool")
			case ctx.Err() != nil:
				w.evHandler("worker: runMiningOperation: MINING: CANCELLED: by request")
			default:
				w.evHandler("worker: runMiningOperation: MINING: ERROR: %s", err)
			}
			return
		}

		// WOW, we mined a block. Send the new block to the network.
		// Log the error, but that's it.
		if err := w.sendBlockToPeers(block); err != nil {
			w.evHandler("worker: runMiningOperation: MINING: sendBlockToPeers: WARNING %s", err)
		}
	}()

	// Wait for both G's to terminate.
	wg.Wait()
}

// sendBlockToPeers takes the new mined block and sends it to all know peers.
func (w *Worker) sendBlockToPeers(block storage.Block) error {
	w.evHandler("worker: runMiningOperation: MINING: sendBlockToPeers: started")
	defer w.evHandler("worker: runMiningOperation: MINING: sendBlockToPeers: completed")

	for _, peer := range w.state.RetrieveKnownPeers() {
		url := fmt.Sprintf("%s/block/next", fmt.Sprintf(w.baseURL, peer.Host))

		var status struct {
			Status string        `json:"status"`
			Block  storage.Block `json:"block"`
		}

		if err := send(http.MethodPost, url, block, &status); err != nil {
			return fmt.Errorf("%s: %s", peer.Host, err)
		}

		w.evHandler("worker: runMiningOperation: MINING: sendBlockToPeers: sent to peer[%s]", peer)
	}

	return nil
}
