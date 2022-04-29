package worker

import (
	"errors"

	"github.com/ardanlabs/blockchain/foundation/blockchain/peer"
)

// peerOperations handles finding new peers.
func (w *Worker) peerOperations() {
	w.evHandler("worker: peerOperations: G started")
	defer w.evHandler("worker: peerOperations: G completed")

	w.runPeersOperation()

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
		peerStatus, err := w.state.NetRequestPeerStatus(peer)
		if err != nil {
			w.evHandler("worker: runPeersOperation: requestPeerStatus: %s: ERROR: %s", peer.Host, err)

			// Since this known peer is unavailable, remove them from the list.
			w.state.RemoveKnownPeer(peer)
		}

		// Add new peers to this nodes list.
		w.addNewPeers(peerStatus.KnownPeers)
	}

	// Share with peers this node is available to participate in the network.
	w.state.NetSendNodeAvailableToPeers()
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

		if w.state.AddKnownPeer(peer) {
			w.evHandler("worker: runPeerUpdatesOperation: addNewPeers: add peer nodes: adding peer-node %s", peer)
		}
	}

	return nil
}
