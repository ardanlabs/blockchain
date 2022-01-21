package node

import "fmt"

// Account represents a user in the system.
type Account string

// BalanceSheet represents the data representation to maintain the balances.
type BalanceSheet map[Account]uint

// newBalanceSheet constructs a new balance sheet for use.
func newBalanceSheet() BalanceSheet {
	return make(BalanceSheet)
}

// update updates the balance sheet for a given account.
func (bs BalanceSheet) update(acct Account, value uint) {
	bs[acct] = value
}

// =============================================================================

// copyBalanceSheet makes a copy of the specified balance sheet.
func copyBalanceSheet(org BalanceSheet) BalanceSheet {
	balanceSheet := newBalanceSheet()
	for acct, value := range org {
		balanceSheet.update(acct, value)
	}
	return balanceSheet
}

// applyTransactionsToBalances applies the transactions to the specified
// balances, adding new accounts as they are found.
func applyTransactionsToBalances(balanceSheet BalanceSheet, txs []Tx) error {
	for _, tx := range txs {
		applyTransactionToBalance(balanceSheet, tx)
	}

	return nil
}

// applyTransactionToBalance performs the business logic for applying a
// transaction to the balance sheet.
func applyTransactionToBalance(balanceSheet BalanceSheet, tx Tx) error {
	if tx.Status == TxStatusError {
		return nil
	}

	if tx.Data == TxDataReward {
		balanceSheet[tx.To] += tx.Value
		return nil
	}

	if tx.From == tx.To {
		return fmt.Errorf("invalid transaction, do you mean to give a reward, from %s, to %s", tx.From, tx.To)
	}

	if tx.Value > balanceSheet[tx.From] {
		return fmt.Errorf("%s has an insufficient balance", tx.From)
	}

	balanceSheet[tx.From] -= tx.Value
	balanceSheet[tx.To] += tx.Value

	return nil
}
