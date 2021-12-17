package chain

import (
	"encoding/json"
	"os"
)

// Chain represents a block chain of data.
type Chain struct {
	Genesis   Genesis
	TxMempool []Tx
	Balances  map[string]uint
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

	ch := Chain{
		Genesis:   genesis,
		TxMempool: txs,
		Balances:  balances,
	}

	return &ch, nil
}

// Add appends a new transactions to the blockchain.
func (ch *Chain) Add(tx Tx) error {
	if err := ch.Apply(tx); err != nil {
		return err
	}

	ch.TxMempool = append(ch.TxMempool, tx)
	return nil
}

// Apply applies a transaction to the Chain's balance sheet.
func (ch *Chain) Apply(tx Tx) error {
	return applyTranToBalance(tx, ch.Balances)
}

// Persist appends the current set of transactions in memory to disk
// and clears the memory. Transactions that fail are returned.
func (ch *Chain) Persist() error {
	path := "zblock/tx.db"
	dbFile, err := os.OpenFile(path, os.O_APPEND|os.O_RDWR, 0600)
	if err != nil {
		return err
	}
	defer dbFile.Close()

	// Marshal and write each transaction to disk. If any
	// transactions fail, keep a record.
	errors := make(TxErrors)
	for _, tx := range ch.TxMempool {
		data, err := json.Marshal(tx)
		if err != nil {
			errors[tx.ID] = TxError{Tx: tx, Err: err}
			continue
		}

		if _, err = dbFile.Write(append(data, '\n')); err != nil {
			errors[tx.ID] = TxError{Tx: tx, Err: err}
			continue
		}
	}

	// Remove all the transactions from memory.
	ch.TxMempool = nil

	// Let the caller decide what to do with the errors if
	// there are any.
	return errors
}
