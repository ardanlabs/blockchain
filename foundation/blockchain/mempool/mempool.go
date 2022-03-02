// Package mempool maintains the mempool for the blockchain.
package mempool

import (
	"fmt"
	"sort"
	"strings"
	"sync"

	"github.com/ardanlabs/blockchain/foundation/blockchain/storage"
)

// Mempool represents a cache of transactions organized by address
// with a second key on the transaction nonce.
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
func (mp *Mempool) Upsert(tx storage.BlockTx) (int, error) {
	mp.mu.Lock()
	defer mp.mu.Unlock()

	key, err := key(tx)
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

	key, err := key(tx)
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

// PickBest returns a list of the best transactions for the next
// mining operation. The caller specifies how many transactions they want.
// Pass -1 for all the transactions.
// The algorithm focuses on the transactions with the best tip while
// respecting the nonce for each address/transaction.
func (mp *Mempool) PickBest(howMany int) []storage.BlockTx {

	// Group the transactions by address.
	m := make(map[string][]storage.BlockTx)
	mp.mu.RLock()
	{
		if howMany == -1 {
			howMany = len(mp.pool)
		}

		for key, tx := range mp.pool {
			fromAddr := strings.Split(key, ":")[0]
			m[fromAddr] = append(m[fromAddr], tx)
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
		if len(m[key]) > 1 {
			sort.Sort(byNonce(m[key]))
		}
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
		0: Bill: {Nonce: 1, To: "0xbEE6ACE826eC3DE1B6349888B9151B92522F7F76", Tip: 150},
		0: Pavl: {Nonce: 1, To: "0xbEE6ACE826eC3DE1B6349888B9151B92522F7F76", Tip: 75},
		0: Edua: {Nonce: 1, To: "0x6Fe6CF3c8fF57c58d24BfC869668F48BCbDb3BD9", Tip: 100},
		1: Bill: {Nonce: 2, To: "0x6Fe6CF3c8fF57c58d24BfC869668F48BCbDb3BD9", Tip: 250},
		1: Pavl: {Nonce: 2, To: "0xa988b1866EaBF72B4c53b592c97aAD8e4b9bDCC0", Tip: 200},
		1: Edua: {Nonce: 2, To: "0xa988b1866EaBF72B4c53b592c97aAD8e4b9bDCC0", Tip: 75},
	*/

	// Sort each row by tip unless we will take all transactions from that row
	// anyway. Then try to select the number of requested transactions. Keep
	// pulling transactions from each row until the amount of fulfilled or
	// there are no more transactions.
	var final []storage.BlockTx
done:
	for _, row := range rows {
		need := howMany - len(final)
		if len(row) > need {
			sort.Sort(byTip(row))
			final = append(final, row[:need]...)
			break done
		}
		final = append(final, row...)
	}

	/*
		0: Bill: {Nonce: 1, To: "0xbEE6ACE826eC3DE1B6349888B9151B92522F7F76", Tip: 150},
		1: Pavl: {Nonce: 1, To: "0xbEE6ACE826eC3DE1B6349888B9151B92522F7F76", Tip: 75},
		2: Edua: {Nonce: 1, To: "0x6Fe6CF3c8fF57c58d24BfC869668F48BCbDb3BD9", Tip: 100},
		3: Bill: {Nonce: 2, To: "0x6Fe6CF3c8fF57c58d24BfC869668F48BCbDb3BD9", Tip: 250},
	*/

	return final
}

// =============================================================================

// key is used to generate the map key.
func key(tx storage.BlockTx) (string, error) {
	from, err := tx.FromAddress()
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s:%d", from, tx.Nonce), nil
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
