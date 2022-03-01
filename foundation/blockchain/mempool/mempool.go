// Package mempool maintains the mempool for the blockchain.
package mempool

import (
	"sort"
	"sync"

	"github.com/ardanlabs/blockchain/foundation/blockchain/storage"
)

// Mempool represents a cache of transactions each with a unique id.
type Mempool struct {
	pool map[string]storage.BlockTx
	mu   sync.RWMutex
}

// New constructs a new mempool to manage pending transactions.
func New() *Mempool {
	return &Mempool{
		pool: make(map[string]storage.BlockTx),
	}
}

// Count returns the current number of transaction in the pool.
func (mp *Mempool) Count() int {
	mp.mu.RLock()
	defer mp.mu.RUnlock()

	return len(mp.pool)
}

// Upsert adds or replaces a transaction from the mempool.
func (mp *Mempool) Upsert(tx storage.BlockTx) int {
	mp.mu.Lock()
	defer mp.mu.Unlock()

	hash := tx.UniqueKey()
	mp.pool[hash] = tx

	return len(mp.pool)
}

// Delete removed a transaction from the mempool.
func (mp *Mempool) Delete(tx storage.BlockTx) {
	mp.mu.Lock()
	defer mp.mu.Unlock()

	hash := tx.UniqueKey()
	delete(mp.pool, hash)
}

// Truncate clears all the transactions from the pool.
func (mp *Mempool) Truncate() {
	mp.mu.Lock()
	defer mp.mu.Unlock()

	mp.pool = make(map[string]storage.BlockTx)
}

// Copy returns a list of the current transaction in the pool.
func (mp *Mempool) Copy() []storage.BlockTx {
	mp.mu.RLock()
	defer mp.mu.RUnlock()

	cpy := make([]storage.BlockTx, 0, len(mp.pool))
	for _, tx := range mp.pool {
		cpy = append(cpy, tx)
	}
	return cpy
}

// CopyBestByTip returns a list of the best transactions for the next
// mining operation based on the offered tip. The caller specifies
// how many transactions they want.
func (mp *Mempool) CopyBestByTip(howMany int) []storage.BlockTx {
	txs := mp.Copy()
	sort.Sort(byTip(txs))

	if howMany == -1 {
		howMany = len(txs)
	}

	cpy := make([]storage.BlockTx, howMany)
	for i := 0; i < howMany; i++ {
		cpy[i] = txs[i]
	}

	sort.Sort(byID(cpy))
	return cpy
}

// =============================================================================

// byTip provides sorting support by the transaction tip value.
type byTip []storage.BlockTx

// Len returns the number of transactions in the list.
func (bt byTip) Len() int {
	return len(bt)
}

// Less helps to sort the list by tip in descending order to choose the
// transactions with the best tip first.
func (bt byTip) Less(i, j int) bool {
	return bt[j].Tip < bt[i].Tip
}

// Swap moves transactions in the order of the tip value.
func (bt byTip) Swap(i, j int) {
	bt[i], bt[j] = bt[j], bt[i]
}

// =============================================================================

// byID provides sorting support by the transaction id value.
type byID []storage.BlockTx

// Len returns the number of transactions in the list.
func (bi byID) Len() int {
	return len(bi)
}

// Less helps to sort the list by ID in ascending order to keep the transactions
// in the right order of processing based on accounts choosing.
func (bi byID) Less(i, j int) bool {
	return bi[i].ID < bi[j].ID
}

// Swap moves transactions in the order of the id value.
func (bi byID) Swap(i, j int) {
	bi[i], bi[j] = bi[j], bi[i]
}
