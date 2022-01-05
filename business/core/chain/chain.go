package chain

import (
	"crypto/sha256"
	"encoding/json"
	"io"
	"os"
	"sync"
)

// Chain represents a block chain of data.
type Chain struct {
	genesis   Genesis
	txMempool []Tx
	hash      [32]byte // Represents a hash of the database.
	balances  map[string]uint
	dbFile    *os.File
	mu        sync.Mutex
}

// New constructs a new blockchain for data management.
func New() (*Chain, error) {

	// Load the genesis file to get starting balances for
	// founders of the block chain.
	genesis, err := loadGenesis()
	if err != nil {
		return nil, err
	}

	// Load the current set of recorded transactions.
	txs, err := loadTransactions()
	if err != nil {
		return nil, err
	}

	// Make a copy of the genesis balances for the next step.
	balances := make(map[string]uint)
	for key, value := range genesis.Balances {
		balances[key] = value
	}

	// Apply the transactions to the initial genesis balances, adding new
	// accounts as it is processed.
	if err := applyTransToBalances(txs, balances); err != nil {
		return nil, err
	}

	// Open the transaction database file.
	path := "zblock/tx.db"
	dbFile, err := os.OpenFile(path, os.O_APPEND|os.O_RDWR, 0600)
	if err != nil {
		return nil, err
	}

	// Create the chain with no transactions currently in memory.
	ch := Chain{
		genesis:  genesis,
		balances: balances,
		dbFile:   dbFile,
	}

	// Capture the current hash.
	if err := ch.captureHash(); err != nil {
		return nil, err
	}

	return &ch, nil
}

// Close cleanly closes the database file underneath.
func (ch *Chain) Close() error {
	ch.mu.Lock()
	defer ch.mu.Unlock()

	return ch.dbFile.Close()
}

// Add appends a new transactions to the blockchain.
func (ch *Chain) Add(tx Tx) error {
	ch.mu.Lock()
	defer ch.mu.Unlock()

	// First apply the transaction to the balance.
	if err := applyTranToBalance(tx, ch.balances); err != nil {
		return err
	}

	// Append the transaction to the in-memory store.
	ch.txMempool = append(ch.txMempool, tx)

	return nil
}

// Persist writes the current transaction memory pool
// to disk.
func (ch *Chain) Persist() error {
	ch.mu.Lock()
	defer ch.mu.Unlock()

	// Make a copy of the data for processing.
	mempool := make([]Tx, len(ch.txMempool))
	copy(mempool, ch.txMempool)

	// Iterate over the set of transactions and after
	// persisting each tran, remove from mempool.
	for _, tx := range mempool {
		if err := ch.persistTran(tx); err != nil {
			return err
		}

		ch.txMempool = ch.txMempool[1:]
	}

	// Capture the new hash for the database.
	if err := ch.captureHash(); err != nil {
		return err
	}

	return nil
}

// Genesis returns a copy of the genesis information.
func (ch *Chain) Genesis() Genesis {
	return ch.genesis
}

// Hash returns the current hash of the blockchain.
func (ch *Chain) Hash() [32]byte {
	return ch.hash
}

// Balances returns the set of balances by account. If the account
// is empty, all balances are returned.
func (ch *Chain) Balances(account string) map[string]uint {
	balances := make(map[string]uint)

	ch.mu.Lock()
	{
		for act, bal := range ch.balances {
			if account == "" || account == act {
				balances[act] = bal
			}
		}
	}
	ch.mu.Unlock()

	return balances
}

// Transactions returns the set of transactions by account. If the account
// is empty, all balances are returned.
func (ch *Chain) Transactions(account string) []Tx {
	var trans []Tx

	ch.mu.Lock()
	{
		for _, tx := range ch.txMempool {
			if account == "" || account == tx.From || account == tx.To {
				trans = append(trans, tx)
			}
		}
	}
	ch.mu.Unlock()

	return trans
}

// =============================================================================

// captureHash produces a hash of the current contents of
// the transaction database.
func (ch *Chain) captureHash() error {

	// Re-read the whole file from the first byte.
	_, err := ch.dbFile.Seek(0, 0)
	if err != nil {
		return err
	}

	txsData, err := io.ReadAll(ch.dbFile)
	if err != nil {
		return err
	}

	ch.hash = sha256.Sum256(txsData)
	return nil
}

// persistTran writes the transaction to disk and returns a new
// snapshot of the transaction database.
func (ch *Chain) persistTran(tx Tx) error {
	data, err := json.Marshal(tx)
	if err != nil {
		return err
	}

	if _, err = ch.dbFile.Write(append(data, '\n')); err != nil {
		return err
	}

	return nil
}
