package database

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
)

// DB represents a block chain of data.
type DB struct {
	genesis      Genesis
	txMempool    []Tx
	lastestBlock [32]byte
	dbPath       string
	balances     map[string]uint
	persistRatio int
	file         *os.File
	mu           sync.Mutex
}

// New constructs a new blockchain for data management.
func New(dbPath string, persistRatio int) (*DB, error) {

	// Load the genesis file to get starting balances for
	// founders of the block chain.
	genesis, err := loadGenesis()
	if err != nil {
		return nil, err
	}

	// Load the current set of recorded transactions.
	blocks, err := loadBlocks(dbPath)
	if err != nil {
		return nil, err
	}

	// Make a copy of the genesis balances for the next step.
	balances := make(map[string]uint)
	for key, value := range genesis.Balances {
		balances[key] = value
	}

	// Open the transaction database file.
	file, err := os.OpenFile(dbPath, os.O_APPEND|os.O_RDWR, 0600)
	if err != nil {
		return nil, err
	}

	// Capture the hash of the latest block.
	var lastestBlock [32]byte
	if len(blocks) > 0 {
		lastestBlock, err = blocks[len(blocks)-1].Hash()
		if err != nil {
			return nil, err
		}
	}

	// Create the chain with no transactions currently in memory.
	db := DB{
		genesis:      genesis,
		lastestBlock: lastestBlock,
		dbPath:       dbPath,
		balances:     balances,
		persistRatio: persistRatio,
		file:         file,
	}

	// Apply the transactions to the initial genesis balances, adding new
	// accounts as it is processed.
	for _, block := range blocks {
		if err := db.applyTransToBalances(block.Transactions); err != nil {
			return nil, err
		}
	}

	return &db, nil
}

// Close cleanly closes the database file underneath.
func (db *DB) Close() error {
	db.mu.Lock()
	defer func() {
		db.file.Close()
		db.mu.Unlock()
	}()

	// Persist the remaining transactions to disk.
	if err := db.createBlock(); err != nil {
		return err
	}

	return nil
}

// Add appends a new transactions to the mempool.
func (db *DB) AddMempool(tx Tx) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	// Append the transaction to the in-memory store.
	db.txMempool = append(db.txMempool, tx)

	// If the number of transactions in the mempool match
	// the number of transactions we want in each block, persist.
	if db.persistRatio == len(db.txMempool) {
		return db.createBlock()
	}

	return nil
}

// Persist writes the current transaction memory pool to disk.
func (db *DB) Persist() error {
	db.mu.Lock()
	defer db.mu.Unlock()

	return db.createBlock()
}

// Genesis returns a copy of the genesis information.
func (db *DB) Genesis() Genesis {
	return db.genesis
}

// LastestBlock returns the current hash of the latest block.
func (db *DB) LastestBlock() [32]byte {
	db.mu.Lock()
	defer db.mu.Unlock()

	return db.lastestBlock
}

// UncommittedTransactions returns a copy of the mempool.
func (db *DB) UncommittedTransactions() []Tx {
	cpy := make([]Tx, len(db.txMempool))
	copy(cpy, db.txMempool)
	return cpy
}

// Balances returns the set of balances by account. If the account
// is empty, all balances are returned.
func (db *DB) Balances(account string) map[string]uint {
	balances := make(map[string]uint)

	db.mu.Lock()
	{
		for act, bal := range db.balances {
			if account == "" || account == act {
				balances[act] = bal
			}
		}
	}
	db.mu.Unlock()

	return balances
}

// Blocks returns the set of blocks by account. If the account
// is empty, all blocks are returned.
func (db *DB) Blocks(account string) []Block {
	blocks, err := loadBlocks(db.dbPath)
	if err != nil {
		return nil
	}

	return blocks
}

// =============================================================================

// createBlock writes the current transaction memory pool to disk.
// It assumes it's always inside a mutex lock.
func (db *DB) createBlock() error {
	if len(db.txMempool) == 0 {
		return nil
	}

	// If the transaction can't be applied to the balance,
	// mark the transaction as failed.
	for i := range db.txMempool {
		if err := db.validateTransaction(db.txMempool[i]); err != nil {
			db.txMempool[i].Status = TxStatusError
			db.txMempool[i].StatusInfo = err.Error()
			continue
		}
		db.txMempool[i].Status = TxStatusAccepted
	}

	blockFS, err := NewBlockFS(db.lastestBlock, db.txMempool)
	if err != nil {
		return err
	}

	blockFSJson, err := json.Marshal(blockFS)
	if err != nil {
		return err
	}

	if _, err := db.file.Write(append(blockFSJson, '\n')); err != nil {
		return err
	}

	db.lastestBlock = blockFS.Hash
	db.txMempool = []Tx{}

	return nil
}

// validateTransaction performs integrity checks on a transaction.
func (db *DB) validateTransaction(tx Tx) error {

	// Validate the transaction can be applied to the balance,
	// checking for things like insufficient funds.
	if err := db.applyTranToBalance(tx); err != nil {
		return err
	}

	return nil
}

// applyTransToBalances applies the transactions to the specified
// balances, adding new accounts as they are found.
func (db *DB) applyTransToBalances(txs []Tx) error {
	for _, tx := range txs {
		db.applyTranToBalance(tx)
	}

	return nil
}

// applyTranToBalance performs the business logic for applying a transaction to
// the balance sheet.
func (db *DB) applyTranToBalance(tx Tx) error {
	if tx.Status == TxStatusError {
		return nil
	}

	if tx.Data == TxDataReward {
		db.balances[tx.To] += tx.Value
		return nil
	}

	if tx.From == tx.To {
		return fmt.Errorf("invalid transaction, do you mean to give a reward, from %s, to %s", tx.From, tx.To)
	}

	if tx.Value > db.balances[tx.From] {
		return fmt.Errorf("%s has an insufficient balance", tx.From)
	}

	db.balances[tx.From] -= tx.Value
	db.balances[tx.To] += tx.Value

	return nil
}
