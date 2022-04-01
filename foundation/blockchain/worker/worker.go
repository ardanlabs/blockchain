// Package worker implements mining, peer updates, and transaction sharing for
// the blockchain.
package worker

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/ardanlabs/blockchain/foundation/blockchain/state"
	"github.com/ardanlabs/blockchain/foundation/blockchain/storage"
)

// peerUpdateInterval represents the interval of finding new peer nodes
// and updating the blockchain on disk with missing blocks.
const peerUpdateInterval = time.Minute

// =============================================================================

// Worker manages the POW workflows for the blockchain.
type Worker struct {
	state        *state.State
	wg           sync.WaitGroup
	ticker       time.Ticker
	shut         chan struct{}
	startMining  chan bool
	cancelMining chan chan struct{}
	txSharing    chan storage.BlockTx
	evHandler    state.EventHandler
	baseURL      string
}

// Run creates a worker, registers the worker with the state package, and
// starts up all the background processes.
func Run(state *state.State, evHandler state.EventHandler) {
	w := Worker{
		state:        state,
		ticker:       *time.NewTicker(peerUpdateInterval),
		shut:         make(chan struct{}),
		startMining:  make(chan bool, 1),
		cancelMining: make(chan chan struct{}, 1),
		txSharing:    make(chan storage.BlockTx, maxTxShareRequests),
		evHandler:    evHandler,
		baseURL:      "http://%s/v1/node",
	}

	// Register this worker with the state package.
	state.Worker = &w

	// Update this node before starting any support G's.
	w.sync()

	// Load the set of operations we need to run.
	operations := []func(){
		w.peerOperations,
		w.miningOperations,
		w.shareTxOperations,
	}

	// Set waitgroup to match the number of G's we need for the set
	// of operations we have.
	g := len(operations)
	w.wg.Add(g)

	// We don't want to return until we know all the G's are up and running.
	hasStarted := make(chan bool)

	// Start all the operational G's.
	for _, op := range operations {
		go func(op func()) {
			defer w.wg.Done()
			hasStarted <- true
			op()
		}(op)
	}

	// Wait for the G's to report they are running.
	for i := 0; i < g; i++ {
		<-hasStarted
	}
}

// =============================================================================
// These methods implement the state.Worker interface.

// Shutdown terminates the goroutine performing work.
func (w *Worker) Shutdown() {
	w.evHandler("worker: shutdown: started")
	defer w.evHandler("worker: shutdown: completed")

	w.evHandler("worker: shutdown: stop ticker")
	w.ticker.Stop()

	w.evHandler("worker: shutdown: signal cancel mining")
	done := w.SignalCancelMining()
	done()

	w.evHandler("worker: shutdown: terminate goroutines")
	close(w.shut)
	w.wg.Wait()
}

// SignalStartMining starts a mining operation. If there is already a signal
// pending in the channel, just return since a mining operation will start.
func (w *Worker) SignalStartMining() {
	select {
	case w.startMining <- true:
	default:
	}
	w.evHandler("worker: SignalStartMining: mining signaled")
}

// SignalCancelMining signals the G executing the runMiningOperation function
// to stop immediately. That G will not return from the function until done
// is called. This allows the caller to complete any state changes before a new
// mining operation takes place.
func (w *Worker) SignalCancelMining() (done func()) {
	wait := make(chan struct{})

	select {
	case w.cancelMining <- wait:
	default:
	}
	w.evHandler("worker: SignalCancelMining: cancel mining signaled")

	return func() { close(wait) }
}

// SignalShareTx signals a share transaction operation. If
// maxTxShareRequests signals exist in the channel, we won't send these.
func (w *Worker) SignalShareTx(blockTx storage.BlockTx) {
	select {
	case w.txSharing <- blockTx:
		w.evHandler("worker: SignalShareTx: share Tx signaled")
	default:
		w.evHandler("worker: SignalShareTx: queue full, transactions won't be shared.")
	}
}

// =============================================================================

// isShutdown is used to test if a shutdown has been signaled.
func (w *Worker) isShutdown() bool {
	select {
	case <-w.shut:
		return true
	default:
		return false
	}
}

// send is a helper function to send an HTTP request to a node.
func send(method string, url string, dataSend interface{}, dataRecv interface{}) error {
	var req *http.Request

	switch {
	case dataSend != nil:
		data, err := json.Marshal(dataSend)
		if err != nil {
			return err
		}
		req, err = http.NewRequest(method, url, bytes.NewReader(data))
		if err != nil {
			return err
		}

	default:
		var err error
		req, err = http.NewRequest(method, url, nil)
		if err != nil {
			return err
		}
	}

	var client http.Client
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNoContent {
		return nil
	}

	if resp.StatusCode != http.StatusOK {
		msg, err := io.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		return errors.New(string(msg))
	}

	if dataRecv != nil {
		if err := json.NewDecoder(resp.Body).Decode(dataRecv); err != nil {
			return err
		}
	}

	return nil
}
