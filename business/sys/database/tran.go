package database

import "fmt"

const (
	TxTypeReward = "reward"
)

// Tx represents a transaction in the database.
type Tx struct {
	From  string `json:"from"`
	To    string `json:"to"`
	Value uint   `json:"value"`
	Data  string `json:"data"`
}

// NewTx constructs a new Tx for use.
func NewTx(from, to string, value uint, data string) Tx {
	return Tx{
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
