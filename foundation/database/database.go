package database

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sync"
	"time"
)

// DB represents a block chain of data.
type DB struct {
	genesis      Genesis
	txMempool    []Tx
	lastestBlock [32]byte
	dbPath       string
	balances     map[string]uint
	file         *os.File
	mu           sync.Mutex

	blockWriter *blockWriter
}

// New constructs a new blockchain for data management.
func New(dbPath string, persistInterval time.Duration, evHandler EventHandler) (*DB, error) {

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
		file:         file,
	}

	// Apply the transactions to the initial genesis balances, adding new
	// accounts as it is processed.
	for _, block := range blocks {
		if err := db.applyTransToBalances(block.Transactions); err != nil {
			return nil, err
		}
	}

	// Start the block writer.
	db.blockWriter = newBlockWriter(&db, persistInterval, evHandler)

	return &db, nil
}

// Close cleanly closes the database file underneath.
func (db *DB) Close() error {
	db.mu.Lock()
	defer func() {
		db.file.Close()
		db.mu.Unlock()
	}()

	db.blockWriter.shutdown()

	// Persist the remaining transactions to disk.
	if _, err := db.createBlock(); err != nil {
		if !errors.Is(err, ErrNoTransactions) {
			return err
		}
	}

	return nil
}

// AddTransaction appends a new transactions to the mempool.
func (db *DB) AddTransaction(tx Tx) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	// Append the transaction to the in-memory store.
	db.txMempool = append(db.txMempool, tx)

	return nil
}

// CreateBlock writes the current transactions from the
// memory pool to disk.
func (db *DB) CreateBlock() (Block, error) {
	db.mu.Lock()
	defer db.mu.Unlock()

	return db.createBlock()
}

// =============================================================================

// QueryGenesis returns a copy of the genesis information.
func (db *DB) QueryGenesis() Genesis {
	db.mu.Lock()
	defer db.mu.Unlock()

	return db.genesis
}

// QueryLastestBlock returns the current hash of the latest block.
func (db *DB) QueryLastestBlock() [32]byte {
	db.mu.Lock()
	defer db.mu.Unlock()

	return db.lastestBlock
}

// QueryMempool returns a copy of the mempool.
func (db *DB) QueryMempool() []Tx {
	db.mu.Lock()
	defer db.mu.Unlock()

	cpy := make([]Tx, len(db.txMempool))
	copy(cpy, db.txMempool)
	return cpy
}

// QueryBalances returns the set of balances by account. If the account
// is empty, all balances are returned.
func (db *DB) QueryBalances(account string) map[string]uint {
	db.mu.Lock()
	defer db.mu.Unlock()

	balances := make(map[string]uint)
	for act, bal := range db.balances {
		if account == "" || account == act {
			balances[act] = bal
		}
	}

	return balances
}

// QueryBlocks returns the set of blocks by account. If the account
// is empty, all blocks are returned.
func (db *DB) QueryBlocks(account string) []Block {
	blocks, err := loadBlocks(db.dbPath)
	if err != nil {
		return nil
	}

	return blocks
}

// =============================================================================

// ErrNoTransactions is returned when a block is requested to be created
// and there are no transactions.
var ErrNoTransactions = errors.New("no transactions in mempool")

// createBlock writes the current transaction memory pool to disk.
// It assumes it's always inside a mutex lock.
func (db *DB) createBlock() (Block, error) {
	if len(db.txMempool) == 0 {
		return Block{}, ErrNoTransactions
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
		return Block{}, err
	}

	blockFSJson, err := json.Marshal(blockFS)
	if err != nil {
		return Block{}, err
	}

	if _, err := db.file.Write(append(blockFSJson, '\n')); err != nil {
		return Block{}, err
	}

	db.lastestBlock = blockFS.Hash
	db.txMempool = []Tx{}

	return blockFS.Block, nil
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
