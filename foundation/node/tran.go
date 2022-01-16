package node

import (
	"fmt"

	"github.com/google/uuid"
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

// Tx represents a transaction in the block.
type Tx struct {
	ID         string `json:"id"`
	Nonce      uint64 `json:"nonce"`
	From       string `json:"from"`
	To         string `json:"to"`
	Value      uint   `json:"value"`
	Data       string `json:"data"`
	Status     string `json:"status"`
	StatusInfo string `json:"status_info"`
}

// NewTx constructs a new TxRecord.
func NewTx(from, to string, value uint, data string) Tx {
	return Tx{
		ID:     uuid.New().String(),
		Nonce:  generateNonce(),
		From:   from,
		To:     to,
		Value:  value,
		Data:   data,
		Status: TxStatusNew,
	}
}

// =============================================================================

// applyTransactionsToBalances applies the transactions to the specified
// balances, adding new accounts as they are found.
func applyTransactionsToBalances(balances map[string]uint, txs []Tx) error {
	for _, tx := range txs {
		applyTransactionToBalance(balances, tx)
	}

	return nil
}

// applyTransactionToBalance performs the business logic for applying a
// transaction to the balance sheet.
func applyTransactionToBalance(balances map[string]uint, tx Tx) error {
	if tx.Status == TxStatusError {
		return nil
	}

	if tx.Data == TxDataReward {
		balances[tx.To] += tx.Value
		return nil
	}

	if tx.From == tx.To {
		return fmt.Errorf("invalid transaction, do you mean to give a reward, from %s, to %s", tx.From, tx.To)
	}

	if tx.Value > balances[tx.From] {
		return fmt.Errorf("%s has an insufficient balance", tx.From)
	}

	balances[tx.From] -= tx.Value
	balances[tx.To] += tx.Value

	return nil
}
