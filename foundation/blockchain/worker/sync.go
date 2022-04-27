package worker

// Sync updates the peer list, mempool and blocks.
func (w *Worker) Sync() {
	w.evHandler("worker: sync: started")
	defer w.evHandler("worker: sync: completed")

	for _, peer := range w.state.RetrieveKnownPeers() {

		// Retrieve the status of this peer.
		peerStatus, err := w.state.NetQueryPeerStatus(peer)
		if err != nil {
			w.evHandler("worker: sync: queryPeerStatus: %s: ERROR: %s", peer.Host, err)
		}

		// Add new peers to this nodes list.
		w.addNewPeers(peerStatus.KnownPeers)

		// Retrieve the mempool from the peer.
		pool, err := w.state.NetRetrievePeerMempool(peer)
		if err != nil {
			w.evHandler("worker: sync: retrievePeerMempool: %s: ERROR: %s", peer.Host, err)
		}
		for _, tx := range pool {
			w.evHandler("worker: sync: retrievePeerMempool: %s: Add Tx: %s", peer.Host, tx.SignatureString()[:16])
			w.state.UpsertMempool(tx)
		}

		// If this peer has blocks we don't have, we need to add them.
		if peerStatus.LatestBlockNumber > w.state.RetrieveLatestBlock().Header.Number {
			w.evHandler("worker: sync: retrievePeerBlocks: %s: latestBlockNumber[%d]", peer.Host, peerStatus.LatestBlockNumber)

			if err := w.state.NetRetrievePeerBlocks(peer); err != nil {
				w.evHandler("worker: sync: retrievePeerBlocks: %s: ERROR %s", peer.Host, err)
			}
		}
	}
}
