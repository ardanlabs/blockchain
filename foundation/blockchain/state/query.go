package state

import (
	"github.com/ardanlabs/blockchain/foundation/blockchain/database"
	"github.com/ardanlabs/blockchain/foundation/blockchain/storage"
)

// QueryLastest represents to query the latest block in the chain.
const QueryLastest = ^uint64(0) >> 1

// =============================================================================

// QueryDatabase returns a copy of the database information by account.
func (s *State) QueryDatabase(account storage.Account) map[storage.Account]database.Info {
	cpy := s.db.Copy()

	final := make(map[storage.Account]database.Info)
	if info, exists := cpy[account]; exists {
		final[account] = info
	}

	return final
}

// QueryMempoolLength returns the current length of the mempool.
func (s *State) QueryMempoolLength() int {
	return s.mempool.Count()
}

// QueryBlocksByNumber returns the set of blocks based on block numbers. This
// function reads the blockchain from disk first.
func (s *State) QueryBlocksByNumber(from uint64, to uint64) []storage.Block {
	blocks, err := s.storage.ReadAllBlocks(s.evHandler, false)
	if err != nil {
		return nil
	}

	if from == QueryLastest {
		from = blocks[len(blocks)-1].Header.Number
		to = from
	}

	var out []storage.Block
	for _, block := range blocks {
		if block.Header.Number >= from && block.Header.Number <= to {
			out = append(out, block)
		}
	}

	return out
}

// QueryBlocksByAccount returns the set of blocks by account. If the account
// is empty, all blocks are returned. This function reads the blockchain
// from disk first.
func (s *State) QueryBlocksByAccount(account storage.Account) []storage.Block {
	blocks, err := s.storage.ReadAllBlocks(s.evHandler, false)
	if err != nil {
		return nil
	}

	var out []storage.Block
blocks:
	for _, block := range blocks {
		for _, tx := range block.Trans.Values() {
			from, err := tx.FromAccount()
			if err != nil {
				continue
			}
			if account == "" || from == account || tx.To == account {
				out = append(out, block)
				continue blocks
			}
		}
	}

	return out
}
