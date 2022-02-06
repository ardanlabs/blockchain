package blockchain

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"sort"
)

// TxMempool represents a cache of transactions each with a unique id.
type TxMempool map[string]BlockTx

// NewTxMempool constructs a new info set to manage node peer information.
func NewTxMempool() TxMempool {
	return make(TxMempool)
}

// Count returns the current number of transaction in the pool.
func (tm TxMempool) Count() int {
	return len(tm)
}

// Add adds a new transaction to the mempool.
func (tm TxMempool) Add(tx BlockTx) {
	hash := hashBlock(tx)
	if _, exists := tm[hash]; !exists {
		tm[hash] = tx
	}
}

// Delete removed a transaction from the mempool.
func (tm TxMempool) Delete(tx BlockTx) {
	hash := hashBlock(tx)
	delete(tm, hash)
}

// Copy returns a list of the current transaction in the pool.
func (tm TxMempool) Copy() []BlockTx {
	cpy := make([]BlockTx, 0, len(tm))
	for _, tx := range tm {
		cpy = append(cpy, tx)
	}
	return cpy
}

// CopyBestByTip returns a list of the best transactions for the next
// mining operation based on the offered tip. The caller specifies
// how many transaction they want.
func (tm TxMempool) CopyBestByTip(amount int) []BlockTx {
	txs := tm.Copy()
	sort.Sort(byTip(txs))

	cpy := make([]BlockTx, amount)
	for i := 0; i < amount; i++ {
		cpy[i] = txs[i]
	}
	return cpy
}

// Hash returns the unique hash for the block by marshaling
// the block into JSON and performing a hashing operation.
func hashBlock(tx BlockTx) string {
	data, err := json.Marshal(tx.UserTx)
	if err != nil {
		return zeroHash
	}

	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
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
