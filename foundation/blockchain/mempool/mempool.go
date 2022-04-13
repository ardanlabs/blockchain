// Package mempool maintains the mempool for the blockchain.
package mempool

import (
	"fmt"
	"strings"
	"sync"

	"github.com/ardanlabs/blockchain/foundation/blockchain/mempool/selector"
	"github.com/ardanlabs/blockchain/foundation/blockchain/storage"
)

// Mempool represents a cache of transactions organized by account:nonce.
type Mempool struct {
	mu       sync.RWMutex
	pool     map[string]storage.BlockTx
	selectFn selector.Func
}

// New constructs a new mempool using the default sort strategy.
func New() (*Mempool, error) {
	return NewWithStrategy(selector.StrategyTip)
}

// NewWithStrategy constructs a new mempool with specified sort strategy.
func NewWithStrategy(strategy string) (*Mempool, error) {
	selectFn, err := selector.Retrieve(strategy)
	if err != nil {
		return nil, err
	}

	mp := Mempool{
		pool:     make(map[string]storage.BlockTx),
		selectFn: selectFn,
	}

	return &mp, nil
}

// Count returns the current number of transaction in the pool.
func (mp *Mempool) Count() int {
	mp.mu.RLock()
	defer mp.mu.RUnlock()

	return len(mp.pool)
}

// Upsert adds or replaces a transaction from the mempool.
func (mp *Mempool) Upsert(tx storage.BlockTx) error {
	mp.mu.Lock()
	defer mp.mu.Unlock()

	key, err := mapKey(tx)
	if err != nil {
		return err
	}

	mp.pool[key] = tx

	return nil
}

// Delete removed a transaction from the mempool.
func (mp *Mempool) Delete(tx storage.BlockTx) error {
	mp.mu.Lock()
	defer mp.mu.Unlock()

	key, err := mapKey(tx)
	if err != nil {
		return err
	}

	delete(mp.pool, key)

	return nil
}

// Truncate clears all the transactions from the pool.
func (mp *Mempool) Truncate() {
	mp.mu.Lock()
	defer mp.mu.Unlock()

	mp.pool = make(map[string]storage.BlockTx)
}

// PickBest uses the configured sort strategy to return the next set
// of transactions for the next block.
func (mp *Mempool) PickBest(howMany ...int) []storage.BlockTx {
	number := -1
	if len(howMany) > 0 {
		number = howMany[0]
	}

	// Copy all the transactions for each account into separate
	// slices for each account.

	m := make(map[storage.Account][]storage.BlockTx)
	mp.mu.RLock()
	{
		if number == -1 {
			number = len(mp.pool)
		}

		for key, tx := range mp.pool {
			account := accountFromMapKey(key)
			m[account] = append(m[account], tx)
		}
	}
	mp.mu.RUnlock()

	return mp.selectFn(m, number)
}

// =============================================================================

// mapKey is used to generate the map key.
func mapKey(tx storage.BlockTx) (string, error) {
	account, err := tx.FromAccount()
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s:%d", account, tx.Nonce), nil
}

// accountFromMapKey extracts the account information from the mapkey.
func accountFromMapKey(key string) storage.Account {
	return storage.Account(strings.Split(key, ":")[0])
}
