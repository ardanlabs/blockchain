package state

import (
	"github.com/ardanlabs/blockchain/foundation/blockchain/database"
	"github.com/ardanlabs/blockchain/foundation/blockchain/genesis"
	"github.com/ardanlabs/blockchain/foundation/blockchain/peer"
)

// RetrieveHost returns a copy of host information.
func (s *State) RetrieveHost() string {
	return s.host
}

// RetrieveConsensus returns a copy of consensus algorithm being used.
func (s *State) RetrieveConsensus() string {
	return s.consensus
}

// RetrieveGenesis returns a copy of the genesis information.
func (s *State) RetrieveGenesis() genesis.Genesis {
	return s.genesis
}

// RetrieveLatestBlock returns a copy the current latest block.
func (s *State) RetrieveLatestBlock() database.Block {
	return s.db.LatestBlock()
}

// RetrieveMempool returns a copy of the mempool.
func (s *State) RetrieveMempool() []database.BlockTx {
	return s.mempool.PickBest()
}

// RetrieveAccounts returns a copy of the database accounts.
func (s *State) RetrieveAccounts() map[database.AccountID]database.Account {
	return s.db.CopyAccounts()
}

// RetrieveKnownExternalPeers retrieves a copy of the known peer list without
// including this node.
func (s *State) RetrieveKnownExternalPeers() []peer.Peer {
	return s.knownPeers.Copy(s.host)
}

// RetrieveKnownPeers retrieves a copy of the full known peer list which includes
// this node as well. Used by the PoA selection algorithm.
func (s *State) RetrieveKnownPeers() []peer.Peer {
	return s.knownPeers.Copy("")
}
