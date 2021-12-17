package db

import (
	"bufio"
	"encoding/json"
	"errors"
	"os"
)

// Chain represents a block chain of data.
type Chain struct {
	Genesis   Genesis
	TxMempool []Tx
	Balances  map[string]uint
}

// NewChain constructs a new blockchain for data management.
func NewChain() (*Chain, error) {

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
	if err := applyTransactionsToBalances(txs, balances); err != nil {
		return nil, err
	}

	ch := Chain{
		Genesis:   genesis,
		TxMempool: txs,
		Balances:  balances,
	}

	return &ch, nil
}

// =============================================================================

// loadGenesis opens and consumes the genesis file.
func loadGenesis() (Genesis, error) {
	path := "zblock/genesis.json"
	content, err := os.ReadFile(path)
	if err != nil {
		return Genesis{}, err
	}

	var genesis Genesis
	err = json.Unmarshal(content, &genesis)
	if err != nil {
		return Genesis{}, err
	}

	return genesis, nil
}

// loadTransactions the current set of transactions.
func loadTransactions() ([]Tx, error) {
	path := "zblock/tx.db"
	dbFile, err := os.OpenFile(path, os.O_APPEND|os.O_RDWR, 0600)
	if err != nil {
		return nil, err
	}
	defer dbFile.Close()

	var txs []Tx

	scanner := bufio.NewScanner(dbFile)
	for scanner.Scan() {
		if err := scanner.Err(); err != nil {
			return nil, err
		}

		var tx Tx
		json.Unmarshal(scanner.Bytes(), &tx)

		txs = append(txs, tx)
	}

	return txs, nil
}

// applyTransactionsToBalances applies the transactions to the specified
// balances, adding new accounts as they are found.
func applyTransactionsToBalances(txs []Tx, balances map[string]uint) error {
	for _, tx := range txs {
		if tx.IsReward() {
			balances[tx.To] += tx.Value
			continue
		}

		if tx.Value > balances[tx.From] {
			return errors.New("insufficient balance, data integrity issue")
		}

		balances[tx.From] -= tx.Value
		balances[tx.To] += tx.Value
	}

	return nil
}
