// Package peer maintains the peer related information such as the set
// of know peers and their status.
package peer

import "sync"

// Peer represents information about a Node in the network.
type Peer struct {
	Host string
}

// New contructs a new info value.
func New(host string) Peer {
	return Peer{
		Host: host,
	}
}

// Match validates if the specified host matches this node.
func (p Peer) Match(host string) bool {
	return p.Host == host
}

// =============================================================================

// PeerStatus represents information about the status
// of any given peer.
type PeerStatus struct {
	LatestBlockHash   string `json:"latest_block_hash"`
	LatestBlockNumber uint64 `json:"latest_block_number"`
	KnownPeers        []Peer `json:"known_peers"`
}

// =============================================================================

// PeerSet represents the data representation to maintain a set of known peers.
type PeerSet struct {
	set map[Peer]struct{}
	mu  sync.RWMutex
}

// NewPeerSet constructs a new info set to manage node peer information.
func NewPeerSet() *PeerSet {
	return &PeerSet{
		set: make(map[Peer]struct{}),
	}
}

// Add adds a new node to the set.
func (ps *PeerSet) Add(peer Peer) {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	if _, exists := ps.set[peer]; !exists {
		ps.set[peer] = struct{}{}
	}
}

// Copy returns a list of the known peers.
func (ps *PeerSet) Copy(host string) []Peer {
	ps.mu.RLock()
	defer ps.mu.RUnlock()

	var peers []Peer
	for peer := range ps.set {
		if !peer.Match(host) {
			peers = append(peers, peer)
		}
	}

	return peers
}
