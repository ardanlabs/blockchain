package chain

import (
	"bufio"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"io"
	"os"

	"github.com/google/uuid"
)

const (
	TxTypeReward = "reward"
)

// Tx represents a transaction in the database.
type Tx struct {
	ID    string `json:"id"`
	From  string `json:"from"`
	To    string `json:"to"`
	Value uint   `json:"value"`
	Data  string `json:"data"`
}

// NewTx constructs a new Tx for use.
func NewTx(from, to string, value uint, data string) Tx {
	return Tx{
		ID:    uuid.New().String(),
		From:  from,
		To:    to,
		Value: value,
		Data:  data,
	}
}

// IsReward tests if the transaction is associated with an award.
func (t Tx) IsReward() bool {
	return t.Data == TxTypeReward
}

// =============================================================================

// loadTransactions the current set of transactions.
func loadTransactions() ([]Tx, error) {
	path := "zblock/tx.db"
	dbFile, err := os.Open(path)
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

// applyTransToBalances applies the transactions to the specified
// balances, adding new accounts as they are found.
func applyTransToBalances(txs []Tx, balances map[string]uint) error {
	for _, tx := range txs {
		applyTranToBalance(tx, balances)
	}

	return nil
}

// applyTranToBalance performs the business logic for applying a transaction to
// the balance sheet.
func applyTranToBalance(tx Tx, balances map[string]uint) error {
	if tx.IsReward() {
		balances[tx.To] += tx.Value
		return nil
	}

	if tx.Value > balances[tx.From] {
		return errors.New("insufficient balance, data integrity issue")
	}

	balances[tx.From] -= tx.Value
	balances[tx.To] += tx.Value

	return nil
}

// persistTran writes the transaction to disk and returns a new
// snapshot of the transaction database.
func persistTran(dbFile *os.File, tx Tx) error {
	data, err := json.Marshal(tx)
	if err != nil {
		return err
	}

	if _, err = dbFile.Write(append(data, '\n')); err != nil {
		return err
	}

	return nil
}

// TakeSnapshot produces a hash of the current contents of
// the transaction database.
func takeSnapshot(dbFile *os.File) ([32]byte, error) {

	// Re-read the whole file from the first byte.
	_, err := dbFile.Seek(0, 0)
	if err != nil {
		return [32]byte{}, err
	}

	txsData, err := io.ReadAll(dbFile)
	if err != nil {
		return [32]byte{}, err
	}

	snapshot := sha256.Sum256(txsData)
	return snapshot, nil
}
