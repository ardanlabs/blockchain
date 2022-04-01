package worker

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/ardanlabs/blockchain/foundation/blockchain/peer"
)

// peerOperations handles finding new peers.
func (w *Worker) peerOperations() {
	w.evHandler("worker: peerOperations: G started")
	defer w.evHandler("worker: peerOperations: G completed")

	for {
		select {
		case <-w.ticker.C:
			if !w.isShutdown() {
				w.runPeersOperation()
			}
		case <-w.shut:
			w.evHandler("worker: peerOperations: received shut signal")
			return
		}
	}
}

// runPeersOperation updates the peer list.
func (w *Worker) runPeersOperation() {
	w.evHandler("worker: runPeersOperation: started")
	defer w.evHandler("worker: runPeersOperation: completed")

	for _, peer := range w.state.RetrieveKnownPeers() {

		// Retrieve the status of this peer.
		peerStatus, err := w.queryPeerStatus(peer)
		if err != nil {
			w.evHandler("worker: runPeersOperation: queryPeerStatus: %s: ERROR: %s", peer.Host, err)
		}

		// Add new peers to this nodes list.
		w.addNewPeers(peerStatus.KnownPeers)
	}
}

// queryPeerStatus looks for new nodes on the blockchain by asking
// known nodes for their peer list. New nodes are added to the list.
func (w *Worker) queryPeerStatus(pr peer.Peer) (peer.PeerStatus, error) {
	w.evHandler("worker: runPeerUpdatesOperation: queryPeerStatus: started: %s", pr)
	defer w.evHandler("worker: runPeerUpdatesOperation: queryPeerStatus: completed: %s", pr)

	url := fmt.Sprintf("%s/status", fmt.Sprintf(w.baseURL, pr.Host))

	var ps peer.PeerStatus
	if err := send(http.MethodGet, url, nil, &ps); err != nil {
		return peer.PeerStatus{}, err
	}

	w.evHandler("worker: runPeerUpdatesOperation: queryPeerStatus: peer-node[%s]: latest-blknum[%d]: peer-list[%s]", pr, ps.LatestBlockNumber, ps.KnownPeers)

	return ps, nil
}

// addNewPeers takes the list of known peers and makes sure they are included
// in the nodes list of know peers.
func (w *Worker) addNewPeers(knownPeers []peer.Peer) error {
	w.evHandler("worker: runPeerUpdatesOperation: addNewPeers: started")
	defer w.evHandler("worker: runPeerUpdatesOperation: addNewPeers: completed")

	for _, peer := range knownPeers {

		// Don't add this running node to the known peer list.
		if peer.Match(w.state.RetrieveHost()) {
			return errors.New("already exists")
		}

		w.evHandler("worker: runPeerUpdatesOperation: addNewPeers: add peer nodes: adding peer-node %s", peer)
		w.state.AddKnownPeer(peer)
	}

	return nil
}
