package blockchain

// Peer represents information about a Node in the network.
type Peer struct {
	Host string
}

// NewPeer contructs a new info value.
func NewPeer(host string) Peer {
	return Peer{
		Host: host,
	}
}

// match validates if the specified host matches this node.
func (p Peer) match(host string) bool {
	return p.Host == host
}

// =============================================================================

// PeerSet represents the data representation to maintain a set of known peers.
type PeerSet map[Peer]struct{}

// NewPeerSet constructs a new info set to manage node peer information.
func NewPeerSet() PeerSet {
	return make(PeerSet)
}

// Add adds a new node to the set.
func (ps PeerSet) Add(peer Peer) {
	if _, exists := ps[peer]; !exists {
		ps[peer] = struct{}{}
	}
}

// =============================================================================

// PeerStatus represents information about the status
// of any given peer.
type PeerStatus struct {
	LatestBlockHash   string `json:"latest_block_hash"`
	LatestBlockNumber uint64 `json:"latest_block_number"`
	KnownPeers        []Peer `json:"known_peers"`
}
