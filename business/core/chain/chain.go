package chain

import "os"

// Chain represents a block chain of data.
type Chain struct {
	Genesis   Genesis
	TxMempool []Tx
	Balances  map[string]uint
	dbFile    *os.File
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
		Genesis:   genesis,
		TxMempool: txs,
		Balances:  balances,
		dbFile:    dbFile,
	}

	return &ch, nil
}

// Close cleanly closes the database file underneath.
func (ch *Chain) Close() error {
	return ch.dbFile.Close()
}

// Add appends a new transactions to the blockchain.
func (ch *Chain) Add(tx Tx) error {

	// First apply the transaction to the balance.
	if err := applyTranToBalance(tx, ch.Balances); err != nil {
		return err
	}

	// Next, write the transaction to disk.
	if err := persistTran(ch.dbFile, tx); err != nil {
		return err
	}

	// Append the transaction to the in-memory store.
	ch.TxMempool = append(ch.TxMempool, tx)

	return nil
}
