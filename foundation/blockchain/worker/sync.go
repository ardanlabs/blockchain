package worker

// CORE NOTE: On startup or when reorganizing the chain, the node needs to be
// in sync with the rest of the network. This includes the mempool and
// blockchain database. This operation needs to finish before the node can
// participate in the network.

// Sync updates the peer list, mempool and blocks.
func (w *Worker) Sync() {
	w.evHandler("worker: sync: started")
	defer w.evHandler("worker: sync: completed")

	for _, peer := range w.state.KnownExternalPeers() {

		// Retrieve the status of this peer.
		peerStatus, err := w.state.NetRequestPeerStatus(peer)
		if err != nil {
			w.evHandler("worker: sync: queryPeerStatus: %s: ERROR: %s", peer.Host, err)
		}

		// Add new peers to this nodes list.
		w.addNewPeers(peerStatus.KnownPeers)

		// Retrieve the mempool from the peer.
		pool, err := w.state.NetRequestPeerMempool(peer)
		if err != nil {
			w.evHandler("worker: sync: retrievePeerMempool: %s: ERROR: %s", peer.Host, err)
		}
		for _, tx := range pool {
			w.evHandler("worker: sync: retrievePeerMempool: %s: Add Tx: %s", peer.Host, tx.SignatureString()[:16])
			w.state.UpsertMempool(tx)
		}

		// If this peer has blocks we don't have, we need to add them.
		if peerStatus.LatestBlockNumber > w.state.LatestBlock().Header.Number {
			w.evHandler("worker: sync: retrievePeerBlocks: %s: latestBlockNumber[%d]", peer.Host, peerStatus.LatestBlockNumber)

			if err := w.state.NetRequestPeerBlocks(peer); err != nil {
				w.evHandler("worker: sync: retrievePeerBlocks: %s: ERROR %s", peer.Host, err)
			}
		}
	}

	// Share with peers this node is available to participate in the network.
	w.state.NetSendNodeAvailableToPeers()
}
