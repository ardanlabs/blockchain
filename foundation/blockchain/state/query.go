package state

import (
	"errors"

	"github.com/ardanlabs/blockchain/foundation/blockchain/database"
)

// QueryLastest represents to query the latest block in the chain.
const QueryLastest = ^uint64(0) >> 1

// =============================================================================

// QueryAccounts returns a copy of the account from the database.
func (s *State) QueryAccounts(account database.AccountID) (database.Account, error) {
	accounts := s.db.CopyAccounts()

	if info, exists := accounts[account]; exists {
		return info, nil
	}

	return database.Account{}, errors.New("not found")
}

// QueryMempoolLength returns the current length of the mempool.
func (s *State) QueryMempoolLength() int {
	return s.mempool.Count()
}

// QueryBlocksByNumber returns the set of blocks based on block numbers. This
// function reads the blockchain from disk first.
func (s *State) QueryBlocksByNumber(from uint64, to uint64) []database.Block {
	blocks, err := s.db.ReadAllBlocks(s.evHandler, false)
	if err != nil {
		return []database.Block{}
	}

	if from == QueryLastest {
		from = blocks[len(blocks)-1].Header.Number
		to = from
	}

	var out []database.Block
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
func (s *State) QueryBlocksByAccount(accountID database.AccountID) []database.Block {
	blocks, err := s.db.ReadAllBlocks(s.evHandler, false)
	if err != nil {
		return []database.Block{}
	}

	var out []database.Block
blocks:
	for _, block := range blocks {
		for _, tx := range block.Trans.Values() {
			fromID, err := tx.FromAccount()
			if err != nil {
				continue
			}
			if accountID == "" || fromID == accountID || tx.ToID == accountID {
				out = append(out, block)
				continue blocks
			}
		}
	}

	return out
}
