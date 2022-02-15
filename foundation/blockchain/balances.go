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

// newBalanceSheet constructs a new balance sheet for use, expects a starting
// balance sheet usually from the genesis file.
func newBalanceSheet(sheet map[string]uint) *BalanceSheet {
	bs := BalanceSheet{
		sheet: make(map[string]uint),
	}

	if sheet != nil {
		bs.reset(sheet)
	}

	return &bs
}

// reset takes the specified sheet and resets the balances.
func (bs *BalanceSheet) reset(sheet map[string]uint) {
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

	balanceSheet := newBalanceSheet(nil)
	for address, value := range bs.sheet {
		balanceSheet.sheet[address] = value
	}
	return balanceSheet
}

// applyValue gives the specififed address the specified value.
func (bs *BalanceSheet) applyValue(address string, value uint) {
	bs.mu.Lock()
	defer bs.mu.Unlock()

	bs.sheet[address] += value
}

// applyMiningFee gives the specified miner the fee for the specified block.
func (bs *BalanceSheet) applyMiningFee(minerAddress string, tx BlockTx) {
	bs.mu.Lock()
	defer bs.mu.Unlock()

	from, err := tx.FromAddress()
	if err != nil {
		return
	}

	fee := tx.Gas + tx.Tip
	bs.sheet[minerAddress] += fee
	bs.sheet[from] -= fee
}

// applyTransaction performs the business logic for applying a transaction
// to the balance sheet.
func (bs *BalanceSheet) applyTransaction(tx BlockTx) error {

	// Capture the address of the account that signed this transaction.
	from, err := tx.FromAddress()
	if err != nil {
		return fmt.Errorf("invalid signature, %s", err)
	}

	bs.mu.Lock()
	defer bs.mu.Unlock()
	{
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
	}

	return nil
}
