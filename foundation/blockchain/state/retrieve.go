package state

import (
	"github.com/ardanlabs/blockchain/foundation/blockchain/database"
	"github.com/ardanlabs/blockchain/foundation/blockchain/genesis"
	"github.com/ardanlabs/blockchain/foundation/blockchain/peer"
	"github.com/ardanlabs/blockchain/foundation/blockchain/storage"
)

// RetrieveHost returns a copy of host information.
func (s *State) RetrieveHost() string {
	return s.host
}

// RetrieveGenesis returns a copy of the genesis information.
func (s *State) RetrieveGenesis() genesis.Genesis {
	return s.genesis
}

// RetrieveLatestBlock returns a copy the current latest block.
func (s *State) RetrieveLatestBlock() storage.Block {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.db.LatestBlock()
}

// RetrieveMempool returns a copy of the mempool.
func (s *State) RetrieveMempool() []storage.BlockTx {
	return s.mempool.PickBest()
}

// RetrieveDatabaseRecords returns a copy of the database records.
func (s *State) RetrieveDatabaseRecords() map[storage.Account]database.Info {
	return s.db.CopyRecords()
}

// RetrieveKnownPeers retrieves a copy of the known peer list.
func (s *State) RetrieveKnownPeers() []peer.Peer {
	return s.knownPeers.Copy(s.host)
}
