// Package mempool maintains the mempool for the blockchain.
package mempool

import (
	"errors"
	"fmt"
	"math"
	"strings"
	"sync"

	"github.com/ardanlabs/blockchain/foundation/blockchain/database"
	"github.com/ardanlabs/blockchain/foundation/blockchain/mempool/selector"
)

// Mempool represents a cache of transactions organized by account:nonce.
type Mempool struct {
	mu       sync.RWMutex
	pool     map[string]database.BlockTx
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
		pool:     make(map[string]database.BlockTx),
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
func (mp *Mempool) Upsert(tx database.BlockTx) error {
	mp.mu.Lock()
	defer mp.mu.Unlock()

	// CORE NOTE: Different blockchains have different algorithms to limit the
	// size of the mempool. Some limit based on the amount of memory being
	// consumed and some may limit based on the number of transaction. If a limit
	// is met, then either the transaction that has the least return on investment
	// or the oldest will be dropped from the pool to make room for new the transaction.

	// For now, the Ardan blockchain in not imposing any limits.
	key, err := mapKey(tx)
	if err != nil {
		return err
	}

	// Ethereum requires a 10% bump in the tip to replace an existing
	// transaction in the mempool and so do we. We want to limit users
	// from this sort of behavior.
	if etx, exists := mp.pool[key]; exists {
		if tx.Tip < uint(math.Round(float64(etx.Tip)*1.10)) {
			return errors.New("replacing a transaction requires a 10% bump in the tip")
		}
	}

	mp.pool[key] = tx

	return nil
}

// Delete removed a transaction from the mempool.
func (mp *Mempool) Delete(tx database.BlockTx) error {
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

	mp.pool = make(map[string]database.BlockTx)
}

// PickBest uses the configured sort strategy to return the next set
// of transactions for the next block. If 0 is passed, all transactions
// in the mempool will be returned.
func (mp *Mempool) PickBest(howMany ...uint) []database.BlockTx {
	number := 0
	if len(howMany) > 0 {
		number = int(howMany[0])
	}

	// Copy all the transactions for each account into separate
	// slices for each account.

	m := make(map[database.AccountID][]database.BlockTx)
	mp.mu.RLock()
	{
		if number == 0 {
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
func mapKey(tx database.BlockTx) (string, error) {
	account, err := tx.FromAccount()
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s:%d", account, tx.Nonce), nil
}

// accountFromMapKey extracts the account information from the mapkey.
func accountFromMapKey(key string) database.AccountID {
	return database.AccountID(strings.Split(key, ":")[0])
}
