// Package mempool maintains the mempool for the blockchain.
package mempool

import (
	"fmt"
	"strings"
	"sync"

	"github.com/ardanlabs/blockchain/foundation/blockchain/storage"
)

// Mempool represents a cache of transactions organized by address
// with a second key on the transaction nonce.
type Mempool struct {
	pool map[string]storage.BlockTx
	mu   sync.RWMutex
	sort SortStrategy
}

// New constructs a new mempool to manage pending transactions.
func New() *Mempool {
	return NewWithSort(SimpleSort)
}

// NewWithSort constructs a new mempool with specified sort strategy.
func NewWithSort(sort SortStrategy) *Mempool {
	return &Mempool{
		pool: make(map[string]storage.BlockTx),
		sort: sort,
	}
}

// Count returns the current number of transaction in the pool.
func (mp *Mempool) Count() int {
	mp.mu.RLock()
	defer mp.mu.RUnlock()

	return len(mp.pool)
}

// Upsert adds or replaces a transaction from the mempool.
func (mp *Mempool) Upsert(tx storage.BlockTx) (int, error) {
	mp.mu.Lock()
	defer mp.mu.Unlock()

	key, err := mapKey(tx)
	if err != nil {
		return 0, err
	}

	mp.pool[key] = tx

	return len(mp.pool), nil
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

func (mp *Mempool) PickBest(howMany int) []storage.BlockTx {

	// Group the transactions by address.
	m := make(map[string][]storage.BlockTx)
	mp.mu.RLock()
	{
		if howMany == -1 {
			howMany = len(mp.pool)
		}

		for key, tx := range mp.pool {
			addr := strings.Split(key, ":")[0]
			m[addr] = append(m[addr], tx)
		}
	}
	mp.mu.RUnlock()
	return mp.sort(m, howMany)
}

// =============================================================================

// mapKey is used to generate the map key.
func mapKey(tx storage.BlockTx) (string, error) {
	addr, err := tx.FromAddress()
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s:%d", addr, tx.Nonce), nil
}

// =============================================================================

// byNonce provides sorting support by the transaction id value.
type byNonce []storage.BlockTx

// Len returns the number of transactions in the list.
func (bn byNonce) Len() int {
	return len(bn)
}

// Less helps to sort the list by nonce in ascending order to keep the
// transactions in the right order of processing.
func (bn byNonce) Less(i, j int) bool {
	return bn[i].Nonce < bn[j].Nonce
}

// Swap moves transactions in the order of the nonce value.
func (bn byNonce) Swap(i, j int) {
	bn[i], bn[j] = bn[j], bn[i]
}

// =============================================================================

// byTip provides sorting support by the transaction tip value.
type byTip []storage.BlockTx

// Len returns the number of transactions in the list.
func (bt byTip) Len() int {
	return len(bt)
}

// Less helps to sort the list by tip in decending order to pick the
// transactions that provide the best reward.
func (bt byTip) Less(i, j int) bool {
	return bt[i].Tip > bt[j].Tip
}

// Swap moves transactions in the order of the tip value.
func (bt byTip) Swap(i, j int) {
	bt[i], bt[j] = bt[j], bt[i]
}
