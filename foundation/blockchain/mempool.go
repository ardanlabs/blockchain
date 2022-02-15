package blockchain

import (
	"sort"
	"sync"
)

// txMempool represents a cache of transactions each with a unique id.
type txMempool struct {
	pool map[string]BlockTx
	mu   sync.RWMutex
}

// newTxMempool constructs a new info set to manage node peer information.
func newTxMempool() *txMempool {
	return &txMempool{
		pool: make(map[string]BlockTx),
	}
}

// count returns the current number of transaction in the pool.
func (tm *txMempool) count() int {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	return len(tm.pool)
}

// add adds a new transaction to the mempool.
func (tm *txMempool) add(tx BlockTx) int {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	hash := tx.Hash()
	if _, exists := tm.pool[hash]; !exists {
		tm.pool[hash] = tx
	}
	return len(tm.pool)
}

// delete removed a transaction from the mempool.
func (tm *txMempool) delete(tx BlockTx) {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	hash := tx.Hash()
	delete(tm.pool, hash)
}

// truncate clears all the transactions from the pool.
func (tm *txMempool) truncate() {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	tm.pool = make(map[string]BlockTx)
}

// copy returns a list of the current transaction in the pool.
func (tm *txMempool) copy() []BlockTx {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	cpy := make([]BlockTx, 0, len(tm.pool))
	for _, tx := range tm.pool {
		cpy = append(cpy, tx)
	}
	return cpy
}

// copyBestByTip returns a list of the best transactions for the next
// mining operation based on the offered tip. The caller specifies
// how many transaction they want.
func (tm *txMempool) copyBestByTip(amount int) []BlockTx {
	txs := tm.copy()

	sort.Sort(byTip(txs))

	cpy := make([]BlockTx, amount)
	for i := 0; i < amount; i++ {
		cpy[i] = txs[i]
	}
	return cpy
}

// =============================================================================

// byTip provides sorting support by the transaction tip value.
type byTip []BlockTx

// Len returns the number of transactions in the list.
func (bt byTip) Len() int {
	return len(bt)
}

// Less returns true or false based on the tip value between two transactions.
func (bt byTip) Less(i, j int) bool {
	return bt[j].Tip < bt[i].Tip
}

// Swap moves transactions in the order of the tip value.
func (bt byTip) Swap(i, j int) {
	bt[i], bt[j] = bt[j], bt[i]
}
