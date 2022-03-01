// Package mempool maintains the mempool for the blockchain.
package mempool

import (
	"sort"
	"sync"

	"github.com/ardanlabs/blockchain/foundation/blockchain/signature"
	"github.com/ardanlabs/blockchain/foundation/blockchain/storage"
)

// Mempool represents a cache of transactions each with a unique id.
type Mempool struct {
	pool map[string]map[uint]storage.BlockTx
	mu   sync.RWMutex
}

// New constructs a new mempool to manage pending transactions.
func New() *Mempool {
	return &Mempool{
		pool: make(map[string]map[uint]storage.BlockTx),
	}
}

// Count returns the current number of transaction in the pool.
func (mp *Mempool) Count() int {
	mp.mu.RLock()
	defer mp.mu.RUnlock()

	var total int
	for k := range mp.pool {
		total += len(mp.pool[k])
	}

	return total
}

// Upsert adds or replaces a transaction from the mempool.
func (mp *Mempool) Upsert(tx storage.BlockTx) int {
	mp.mu.Lock()
	defer mp.mu.Unlock()

	from, err := signature.FromAddress(tx.UserTx, tx.V, tx.R, tx.S)
	if err != nil {
		from = "unknown"
	}

	if mp.pool[from] == nil {
		mp.pool[from] = make(map[uint]storage.BlockTx)
	}
	mp.pool[from][tx.Nonce] = tx

	return len(mp.pool)
}

// Delete removed a transaction from the mempool.
func (mp *Mempool) Delete(tx storage.BlockTx) {
	mp.mu.Lock()
	defer mp.mu.Unlock()

	from, err := signature.FromAddress(tx.UserTx, tx.V, tx.R, tx.S)
	if err != nil {
		from = "unknown"
	}

	if _, exists := mp.pool[from]; exists {
		delete(mp.pool[from], tx.Nonce)
		if len(mp.pool[from]) == 0 {
			delete(mp.pool, from)
		}
	}
}

// Truncate clears all the transactions from the pool.
func (mp *Mempool) Truncate() {
	mp.mu.Lock()
	defer mp.mu.Unlock()

	mp.pool = make(map[string]map[uint]storage.BlockTx)
}

// PickBest returns a list of the best transactions for the next
// mining operation. The caller specifies how many transactions they want.
// Pass -1 for all the transactions.
// The algorithm focuses on the transactions with the best tip while
// respecting the nonce for each address/transaction.
func (mp *Mempool) PickBest(howMany int) []storage.BlockTx {

	// Convert the transactions by address to a slice.
	m := make(map[string][]storage.BlockTx)
	mp.mu.RLock()
	{
		for from, trans := range mp.pool {
			for _, tx := range trans {
				m[from] = append(m[from], tx)
			}
		}
	}
	mp.mu.RUnlock()

	/*
		Bill: {Nonce: 2, To: "0x6Fe6CF3c8fF57c58d24BfC869668F48BCbDb3BD9", Tip: 250},
			  {Nonce: 1, To: "0xbEE6ACE826eC3DE1B6349888B9151B92522F7F76", Tip: 150},
		Pavl: {Nonce: 2, To: "0xa988b1866EaBF72B4c53b592c97aAD8e4b9bDCC0", Tip: 200},
			  {Nonce: 1, To: "0xbEE6ACE826eC3DE1B6349888B9151B92522F7F76", Tip: 75},
		Edua: {Nonce: 2, To: "0xa988b1866EaBF72B4c53b592c97aAD8e4b9bDCC0", Tip: 75},
			  {Nonce: 1, To: "0x6Fe6CF3c8fF57c58d24BfC869668F48BCbDb3BD9", Tip: 100},
	*/

	// Sort the transactions per address by nonce.
	for key := range m {
		sort.Sort(byNonce(m[key]))
	}

	/*
		Bill: {Nonce: 1, To: "0xbEE6ACE826eC3DE1B6349888B9151B92522F7F76", Tip: 150},
		      {Nonce: 2, To: "0x6Fe6CF3c8fF57c58d24BfC869668F48BCbDb3BD9", Tip: 250},
		Pavl: {Nonce: 1, To: "0xbEE6ACE826eC3DE1B6349888B9151B92522F7F76", Tip: 75},
		      {Nonce: 2, To: "0xa988b1866EaBF72B4c53b592c97aAD8e4b9bDCC0", Tip: 200},
		Edua: {Nonce: 1, To: "0x6Fe6CF3c8fF57c58d24BfC869668F48BCbDb3BD9", Tip: 100},
		      {Nonce: 2, To: "0xa988b1866EaBF72B4c53b592c97aAD8e4b9bDCC0", Tip: 75},
	*/

	// Pick the first transaction in the slice for each address. Each iteration
	// represents a new row of selections. Keep doing that until all the
	// transactions have been selected.
	var rows [][]storage.BlockTx
	for {
		var row []storage.BlockTx
		for key := range m {
			if len(m[key]) > 0 {
				row = append(row, m[key][0])
				m[key] = m[key][1:]
			}
		}
		if row == nil {
			break
		}
		rows = append(rows, row)
	}

	/*
		0: {Nonce: 1, To: "0xbEE6ACE826eC3DE1B6349888B9151B92522F7F76", Tip: 150},
		0: {Nonce: 1, To: "0xbEE6ACE826eC3DE1B6349888B9151B92522F7F76", Tip: 75},
		0: {Nonce: 1, To: "0x6Fe6CF3c8fF57c58d24BfC869668F48BCbDb3BD9", Tip: 100},
		1: {Nonce: 2, To: "0x6Fe6CF3c8fF57c58d24BfC869668F48BCbDb3BD9", Tip: 250},
		1: {Nonce: 2, To: "0xa988b1866EaBF72B4c53b592c97aAD8e4b9bDCC0", Tip: 200},
		1: {Nonce: 2, To: "0xa988b1866EaBF72B4c53b592c97aAD8e4b9bDCC0", Tip: 75},
	*/

	if howMany == -1 {
		howMany = 0
		for k := range m {
			howMany += len(m[k])
		}
	}

	// Sort each row by tip and then try to select the number of requested
	// transactions. Keep pulling transactions from each row until the
	// amount of fulfilled or there are no more transactions.
	var final []storage.BlockTx
done:
	for _, row := range rows {
		sort.Sort(byTip(row))
		for _, tx := range row {
			final = append(final, tx)
			if len(final) == howMany {
				break done
			}
		}
	}

	/*
		{Nonce: 1, To: "0xbEE6ACE826eC3DE1B6349888B9151B92522F7F76", Tip: 150},
		{Nonce: 1, To: "0xbEE6ACE826eC3DE1B6349888B9151B92522F7F76", Tip: 75},
		{Nonce: 1, To: "0x6Fe6CF3c8fF57c58d24BfC869668F48BCbDb3BD9", Tip: 100},
		{Nonce: 2, To: "0x6Fe6CF3c8fF57c58d24BfC869668F48BCbDb3BD9", Tip: 250},
	*/

	return final
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
