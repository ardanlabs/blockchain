package node

// Info represents information about a Node in the network.
type Info string

// NewInfo contructs a new info value.
func NewInfo(name string) Info {
	return Info(name)
}

// match validates if the specified know matches this node.
func (n Info) match(node Info) bool {
	return n == node
}

// =============================================================================

// PeerSet represents the data representation to maintain a set of known peers.
type PeerSet map[Info]struct{}

// NewPeerSet constructs a new info set to manage node peer information.
func NewPeerSet() PeerSet {
	return make(PeerSet)
}

// Add adds a new node to the set.
func (ps PeerSet) Add(peer Info) {
	if _, exists := ps[peer]; !exists {
		ps[peer] = struct{}{}
	}
}
