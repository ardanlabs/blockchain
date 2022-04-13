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

	for _, peer := range w.state.RetrieveKnownPeers() {
		url := fmt.Sprintf("%s/tx/submit", fmt.Sprintf(w.baseURL, peer.Host))
		if err := send(http.MethodPost, url, tx, nil); err != nil {
			w.evHandler("worker: runShareTxOperation: WARNING: %s", err)
		}
	}
}
