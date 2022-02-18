// Package balance maintains account balances in memory.
package balance

import (
	"fmt"
	"sync"

	"github.com/ardanlabs/blockchain/foundation/blockchain/storage"
)

// Sheet represents the data representation to maintain address balances.
type Sheet struct {
	sheet map[string]uint
	mu    sync.RWMutex
}

// NewSheet constructs a new balance sheet for use, expects a starting
// balance sheet usually from a genesis file.
func NewSheet(sheet map[string]uint) *Sheet {
	bs := Sheet{
		sheet: make(map[string]uint),
	}

	if sheet != nil {
		bs.Reset(sheet)
	}

	return &bs
}

// Reset takes the specified sheet and resets the balances.
func (bs *Sheet) Reset(sheet map[string]uint) {
	bs.mu.Lock()
	defer bs.mu.Unlock()

	bs.sheet = make(map[string]uint)
	for address, value := range sheet {
		bs.sheet[address] = value
	}
}

// Replace updates the balance sheet for a new version.
func (bs *Sheet) Replace(newBS *Sheet) {
	bs.mu.Lock()
	defer bs.mu.Unlock()

	bs.sheet = newBS.sheet
}

// Remove deletes the address from the balance sheet.
func (bs *Sheet) Remove(address string) {
	bs.mu.Lock()
	defer bs.mu.Unlock()

	delete(bs.sheet, address)
}

// Clone makes a copy of the current balance sheet.
func (bs *Sheet) Clone() *Sheet {
	bs.mu.Lock()
	defer bs.mu.Unlock()

	balanceSheet := NewSheet(nil)
	for address, value := range bs.sheet {
		balanceSheet.sheet[address] = value
	}
	return balanceSheet
}

// Copy makes a copy of the current balance sheet but returns the raw data.
func (bs *Sheet) Copy() map[string]uint {
	bs.mu.Lock()
	defer bs.mu.Unlock()

	sheet := make(map[string]uint)
	for address, value := range bs.sheet {
		sheet[address] = value
	}
	return sheet
}

// ApplyValue gives the specififed address the specified value.
func (bs *Sheet) ApplyValue(address string, value uint) {
	bs.mu.Lock()
	defer bs.mu.Unlock()

	bs.sheet[address] += value
}

// ApplyMiningFee gives the specified miner the fee for the specified block.
func (bs *Sheet) ApplyMiningFee(minerAddress string, tx storage.BlockTx) {

	// Capture the address of the account that signed this transaction.
	from, err := tx.FromAddress()
	if err != nil {
		return
	}

	bs.mu.Lock()
	defer bs.mu.Unlock()
	{
		fee := tx.Gas + tx.Tip
		bs.sheet[minerAddress] += fee
		bs.sheet[from] -= fee
	}
}

// ApplyTransaction performs the business logic for applying a transaction
// to the balance sheet.
func (bs *Sheet) ApplyTransaction(tx storage.BlockTx) error {

	// Capture the address of the account that signed this transaction.
	from, err := tx.FromAddress()
	if err != nil {
		return fmt.Errorf("invalid signature, %s", err)
	}

	bs.mu.Lock()
	defer bs.mu.Unlock()
	{
		if string(tx.Data) == storage.TxDataReward {
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
