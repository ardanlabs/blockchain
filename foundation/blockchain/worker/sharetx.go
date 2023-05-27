package worker

import (
	"context"
	"sync"
)

// CORE NOTE: Sharing new transactions received directly by a wallet is
// performed by this goroutine. When a wallet transaction is received,
// the request goroutine shares it with this goroutine to send it over the
// p2p network. Up to 100 transactions can be pending to be sent before new
// transactions are dropped and not sent.

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

	var wg sync.WaitGroup
	wg.Add(1)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		defer wg.Done()

		for {
			select {
			case tx := <-w.txSharing:
				if !w.isShutdown() {
					w.state.NetSendTxToPeers(ctx, tx)
				}
			case <-w.shut:
				return
			}
		}
	}()

	<-w.shut
	w.evHandler("worker: shareTxOperations: received shut signal")

	cancel()
	wg.Wait()
}
