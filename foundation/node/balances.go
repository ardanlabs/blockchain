package node

import "fmt"

// BalanceSheet represents the data representation to maintain account balances.
type BalanceSheet map[string]uint

// newBalanceSheet constructs a new balance sheet for use.
func newBalanceSheet() BalanceSheet {
	return make(BalanceSheet)
}

// replace updates the balance sheet for a given account.
func (bs BalanceSheet) replace(account string, value uint) {
	bs[account] = value
}

// =============================================================================

// copyBalanceSheet makes a copy of the specified balance sheet.
func copyBalanceSheet(org BalanceSheet) BalanceSheet {
	balanceSheet := newBalanceSheet()
	for acct, value := range org {
		balanceSheet.replace(acct, value)
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

// applyMiningRewardToBalance gives the miner account a reward for mining a block.
func applyMiningRewardToBalance(balanceSheet BalanceSheet, minerAccount string, reward uint) {
	balanceSheet[minerAccount] += reward
}

// applyTransactionToBalance performs the business logic for applying a
// transaction to the balance sheet.
func applyTransactionToBalance(balanceSheet BalanceSheet, tx Tx) error {
	if tx.Status == TxStatusError {
		return nil
	}

	if tx.Data == TxDataReward {
		balanceSheet[tx.ToAccount] += tx.Value
		return nil
	}

	if tx.FromAccount == tx.ToAccount {
		return fmt.Errorf("invalid transaction, do you mean to give a reward, from %s, to %s", tx.FromAccount, tx.ToAccount)
	}

	if tx.Value > balanceSheet[tx.FromAccount] {
		return fmt.Errorf("%s has an insufficient balance", tx.FromAccount)
	}

	balanceSheet[tx.FromAccount] -= tx.Value
	balanceSheet[tx.ToAccount] += tx.Value

	return nil
}
