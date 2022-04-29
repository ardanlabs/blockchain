package state

import (
	"github.com/ardanlabs/blockchain/foundation/blockchain/database"
	"github.com/ardanlabs/blockchain/foundation/blockchain/peer"
)

// AddKnownPeer provides the ability to add a new peer to
// the known peer list.
func (s *State) AddKnownPeer(peer peer.Peer) bool {
	return s.knownPeers.Add(peer)
}

// RemoveKnownPeer provides the ability to remove a peer from
// the known peer list.
func (s *State) RemoveKnownPeer(peer peer.Peer) {
	s.knownPeers.Remove(peer)
}

// UpsertMempool adds a new transaction to the mempool.
func (s *State) UpsertMempool(tx database.BlockTx) error {
	return s.mempool.Upsert(tx)
}
