package blockchain

import (
	"fmt"
	"sync"
)

// BalanceSheet represents the data representation to maintain address balances.
type BalanceSheet struct {
	sheet map[string]uint
	mu    sync.RWMutex
}

// newBalanceSheet constructs a new balance sheet for use.
func newBalanceSheet() *BalanceSheet {
	return &BalanceSheet{
		sheet: make(map[string]uint),
	}
}

// newBalanceSheetFromSheet constructs a new balance sheet with existing values.
func newBalanceSheetFromSheet(sheet map[string]uint) *BalanceSheet {
	balanceSheet := newBalanceSheet()
	for address, value := range sheet {
		balanceSheet.sheet[address] = value
	}
	return balanceSheet
}

// resetFromSheet takes the specified sheet and resets the balances.
func (bs *BalanceSheet) resetFromSheet(sheet map[string]uint) {
	bs.mu.Lock()
	defer bs.mu.Unlock()

	bs.sheet = make(map[string]uint)
	for address, value := range sheet {
		bs.sheet[address] = value
	}
}

// replace updates the balance sheet for a new version.
func (bs *BalanceSheet) replace(newBS *BalanceSheet) {
	bs.mu.Lock()
	defer bs.mu.Unlock()

	bs.sheet = newBS.sheet
}

// remove deletes the address from the balance sheet.
func (bs *BalanceSheet) remove(address string) {
	bs.mu.Lock()
	defer bs.mu.Unlock()

	delete(bs.sheet, address)
}

// copy makes a copy of the current balance sheet.
func (bs *BalanceSheet) copy() *BalanceSheet {
	bs.mu.Lock()
	defer bs.mu.Unlock()

	balanceSheet := newBalanceSheet()
	for address, value := range bs.sheet {
		balanceSheet.sheet[address] = value
	}
	return balanceSheet
}

func (bs *BalanceSheet) applyMiningRewardToBalance(beneficiary string, reward uint) {
	bs.mu.Lock()
	defer bs.mu.Unlock()

	bs.sheet[beneficiary] += reward
}

func (bs *BalanceSheet) applyMiningFeeToBalance(beneficiary string, tx BlockTx) {
	bs.mu.Lock()
	defer bs.mu.Unlock()

	from, err := tx.FromAddress()
	if err != nil {
		return
	}

	fee := tx.Gas + tx.Tip
	bs.sheet[beneficiary] += fee
	bs.sheet[from] -= fee
}

// applyTransactionToBalance performs the business logic for applying a
// transaction to the balance sheet.
func (bs *BalanceSheet) applyTransactionToBalance(tx BlockTx) error {

	// Capture the address of the account that signed this transaction.
	from, err := tx.FromAddress()
	if err != nil {
		return fmt.Errorf("invalid signature, %s", err)
	}

	bs.mu.Lock()
	defer bs.mu.Unlock()

	if string(tx.Data) == TxDataReward {
		bs.sheet[tx.To] += tx.Value
		return nil
	}

	if from == tx.To {
		return fmt.Errorf("invalid transaction, do you mean to give a reward, from %s, to %s", from, tx.To)
	}

	if tx.Value > bs.sheet[from] {
		return fmt.Errorf("%s has an insufficient balance", from)
	}

	bs.sheet[from] -= tx.Value
	bs.sheet[tx.To] += tx.Value

	if tx.Tip > 0 {
		bs.sheet[from] -= tx.Tip
	}

	return nil
}
