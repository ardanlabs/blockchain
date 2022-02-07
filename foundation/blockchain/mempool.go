package blockchain

import (
	"sort"
)

// txMempool represents a cache of transactions each with a unique id.
type txMempool map[string]BlockTx

// newTxMempool constructs a new info set to manage node peer information.
func newTxMempool() txMempool {
	return make(txMempool)
}

// count returns the current number of transaction in the pool.
func (tm txMempool) count() int {
	return len(tm)
}

// add adds a new transaction to the mempool.
func (tm txMempool) add(tx BlockTx) {
	hash := tx.Hash()
	if _, exists := tm[hash]; !exists {
		tm[hash] = tx
	}
}

// delete removed a transaction from the mempool.
func (tm txMempool) delete(tx BlockTx) {
	hash := tx.Hash()
	delete(tm, hash)
}

// copy returns a list of the current transaction in the pool.
func (tm txMempool) copy() []BlockTx {
	cpy := make([]BlockTx, 0, len(tm))
	for _, tx := range tm {
		cpy = append(cpy, tx)
	}
	return cpy
}

// copyBestByTip returns a list of the best transactions for the next
// mining operation based on the offered tip. The caller specifies
// how many transaction they want.
func (tm txMempool) copyBestByTip(amount int) []BlockTx {
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
