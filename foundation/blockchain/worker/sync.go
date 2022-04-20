package worker

import (
	"fmt"
	"net/http"

	"github.com/ardanlabs/blockchain/foundation/blockchain/database"
	"github.com/ardanlabs/blockchain/foundation/blockchain/peer"
)

// Sync updates the peer list, mempool and blocks.
func (w *Worker) Sync() {
	w.evHandler("worker: sync: started")
	defer w.evHandler("worker: sync: completed")

	for _, peer := range w.state.RetrieveKnownPeers() {

		// Retrieve the status of this peer.
		peerStatus, err := w.queryPeerStatus(peer)
		if err != nil {
			w.evHandler("worker: sync: queryPeerStatus: %s: ERROR: %s", peer.Host, err)
		}

		// Add new peers to this nodes list.
		w.addNewPeers(peerStatus.KnownPeers)

		// Retrieve the mempool from the peer.
		pool, err := w.retrievePeerMempool(peer)
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

			if err := w.retrievePeerBlocks(peer); err != nil {
				w.evHandler("worker: sync: retrievePeerBlocks: %s: ERROR %s", peer.Host, err)
			}
		}
	}
}

// retrievePeerMempool asks the peer for the transactions in their mempool.
func (w *Worker) retrievePeerMempool(pr peer.Peer) ([]database.BlockTx, error) {
	w.evHandler("worker: sync: retrievePeerMempool: started: %s", pr)
	defer w.evHandler("worker: sync: retrievePeerMempool: completed: %s", pr)

	url := fmt.Sprintf("%s/tx/list", fmt.Sprintf(w.baseURL, pr.Host))

	var mempool []database.BlockTx
	if err := send(http.MethodGet, url, nil, &mempool); err != nil {
		return nil, err
	}

	w.evHandler("worker: sync: retrievePeerMempool: len[%d]", len(mempool))

	return mempool, nil
}

// retrievePeerBlocks queries the specified node asking for blocks this node does
// not have, then writes them to disk.
func (w *Worker) retrievePeerBlocks(pr peer.Peer) error {
	w.evHandler("worker: sync: retrievePeerBlocks: started: %s", pr)
	defer w.evHandler("worker: sync: retrievePeerBlocks: completed: %s", pr)

	from := w.state.RetrieveLatestBlock().Header.Number + 1
	url := fmt.Sprintf("%s/block/list/%d/latest", fmt.Sprintf(w.baseURL, pr.Host), from)

	var blocksFS []database.BlockFS
	if err := send(http.MethodGet, url, nil, &blocksFS); err != nil {
		return err
	}

	w.evHandler("worker: sync: retrievePeerBlocks: found blocks[%d]", len(blocksFS))

	for _, blockFS := range blocksFS {
		block, err := database.ToBlock(blockFS)
		if err != nil {
			return err
		}

		w.evHandler("worker: sync: retrievePeerBlocks: prevBlk[%s]: newBlk[%s]: numTrans[%d]", block.Header.ParentHash, block.Hash(), len(block.Trans.Values()))
		if err := w.state.ProcessProposedBlock(block); err != nil {
			return err
		}
	}

	return nil
}
