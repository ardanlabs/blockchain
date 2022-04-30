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
	if from == QueryLastest {
		from = s.db.LatestBlock().Header.Number
		to = from
	}
	if to == QueryLastest {
		to = s.db.LatestBlock().Header.Number
	}

	var out []database.Block
	for i := from; i <= to; i++ {
		block, err := s.db.GetBlock(i)
		if err != nil {
			s.evHandler("state: getblock: ERROR: %s", err)
			return nil
		}
		out = append(out, block)
	}

	return out
}

// QueryBlocksByAccount returns the set of blocks by account. If the account
// is empty, all blocks are returned. This function reads the blockchain
// from disk first.
func (s *State) QueryBlocksByAccount(accountID database.AccountID) ([]database.Block, error) {
	var out []database.Block

	iter := s.db.ForEach()
	for block, err := iter.Next(); !iter.Done(); block, err = iter.Next() {
		if err != nil {
			return nil, err
		}

		for _, tx := range block.Trans.Values() {
			fromID, err := tx.FromAccount()
			if err != nil {
				continue
			}

			if accountID == "" || fromID == accountID || tx.ToID == accountID {
				out = append(out, block)
				break
			}
		}
	}

	return out, nil
}
