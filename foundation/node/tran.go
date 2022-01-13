package node

import (
	"fmt"
)

const (
	TxDataReward = "reward"
)

const (
	TxStatusAccepted  = "accepted"
	TxStatusError     = "error"
	TxStatusNew       = "new"
	TxStatusPublished = "published"
)

// TxRecord represents a transaction in the database.
type TxRecord struct {
	Nonce uint64 `json:"nonce"`
	From  string `json:"from"`
	To    string `json:"to"`
	Value uint   `json:"value"`
	Data  string `json:"data"`
}

// Tx represents what is written to the DB file.
type Tx struct {
	Hash       string   `json:"hash"`
	Status     string   `json:"status"`
	StatusInfo string   `json:"status_info"`
	Record     TxRecord `json:"tx"`
}

// NewTx constructs a new Tx for use.
func NewTx(from, to string, value uint, data string) Tx {
	txRecord := TxRecord{
		Nonce: generateNonce(),
		From:  from,
		To:    to,
		Value: value,
		Data:  data,
	}

	return Tx{
		Hash:   generateHash(txRecord),
		Status: TxStatusNew,
		Record: txRecord,
	}
}

// =============================================================================

// applyTransToBalances applies the transactions to the specified
// balances, adding new accounts as they are found.
func applyTransToBalances(balances map[string]uint, txs []Tx) error {
	for _, tx := range txs {
		applyTranToBalance(balances, tx)
	}

	return nil
}

// applyTranToBalance performs the business logic for applying a transaction to
// the balance sheet.
func applyTranToBalance(balances map[string]uint, tx Tx) error {
	if tx.Status == TxStatusError {
		return nil
	}

	if tx.Record.Data == TxDataReward {
		balances[tx.Record.To] += tx.Record.Value
		return nil
	}

	if tx.Record.From == tx.Record.To {
		return fmt.Errorf("invalid transaction, do you mean to give a reward, from %s, to %s", tx.Record.From, tx.Record.To)
	}

	if tx.Record.Value > balances[tx.Record.From] {
		return fmt.Errorf("%s has an insufficient balance", tx.Record.From)
	}

	balances[tx.Record.From] -= tx.Record.Value
	balances[tx.Record.To] += tx.Record.Value

	return nil
}
