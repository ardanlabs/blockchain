package worker

import (
	"fmt"
	"net/http"

	"github.com/ardanlabs/blockchain/foundation/blockchain/database"
)

// maxTxShareRequests represents the max number of pending tx network share
// requests that can be outstanding before share requests are dropped. To keep
// this simple, a buffered channel of this arbitrary number is being used. If
// the channel does become full, requests for new transactions to be shared
// will not be accepted.
const maxTxShareRequests = 100

// =============================================================================

// shareTxOperations handles sharing new block transactions.
func (w *Worker) shareTxOperations() {
	w.evHandler("worker: shareTxOperations: G started")
	defer w.evHandler("worker: shareTxOperations: G completed")

	for {
		select {
		case tx := <-w.txSharing:
			if !w.isShutdown() {
				w.runShareTxOperation(tx)
			}
		case <-w.shut:
			w.evHandler("worker: shareTxOperations: received shut signal")
			return
		}
	}
}

// runShareTxOperation shares a new block transactions with the known peers.
func (w *Worker) runShareTxOperation(tx database.BlockTx) {
	w.evHandler("worker: runShareTxOperation: started")
	defer w.evHandler("worker: runShareTxOperation: completed")

	// Bitcoin does not send the full transaction immediately to save on
	// bandwidth. A node will send the transaction's mempool key first so
	// the receiving node can check if they already have the transaction or
	// not. If the receiving node doesn't have it, then it will request the
	// transaction based on the mempool key it received.

	// For now, the Ardan blockchain just sends the full transaction.
	for _, peer := range w.state.RetrieveKnownPeers() {
		url := fmt.Sprintf("%s/tx/submit", fmt.Sprintf(w.baseURL, peer.Host))
		if err := send(http.MethodPost, url, tx, nil); err != nil {
			w.evHandler("worker: runShareTxOperation: WARNING: %s", err)
		}
	}
}
