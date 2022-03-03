// Package accounts maintains account balances in memory.
package accounts

import (
	"fmt"
	"sync"

	"github.com/ardanlabs/blockchain/foundation/blockchain/genesis"
	"github.com/ardanlabs/blockchain/foundation/blockchain/storage"
)

// Info represents information stored in the balance sheet.
type Info struct {
	Balance uint
	Nonce   uint
}

// Accounts manages data related to accounts who have transacted on
// the blockchain.
type Accounts struct {
	genesis genesis.Genesis
	info    map[string]Info
	mu      sync.RWMutex
}

func New(genesis genesis.Genesis) *Accounts {
	accts := Accounts{
		genesis: genesis,
		info:    make(map[string]Info),
	}

	for addr, balance := range genesis.Balances {
		accts.info[addr] = Info{Balance: balance}
	}

	return &accts
}

// Reset re-initalizes the balance sheet from the genesis information.
func (act *Accounts) Reset() {
	act.mu.Lock()
	defer act.mu.Unlock()

	act.info = make(map[string]Info)
	for addr, balance := range act.genesis.Balances {
		act.info[addr] = Info{Balance: balance}
	}
}

// Replace updates the balance sheet for a new version.
func (act *Accounts) Replace(accounts *Accounts) {
	act.mu.Lock()
	defer act.mu.Unlock()

	act.info = accounts.info
}

// Remove deletes the addr from the balance sheet.
func (act *Accounts) Remove(addr string) {
	act.mu.Lock()
	defer act.mu.Unlock()

	delete(act.info, addr)
}

// Clone makes a copy of the current balance sheet.
func (act *Accounts) Clone() *Accounts {
	act.mu.RLock()
	defer act.mu.RUnlock()

	accounts := New(act.genesis)
	for addr, value := range act.info {
		accounts.info[addr] = value
	}
	return accounts
}

// Copy makes a copy of the current information for all accounts.
func (act *Accounts) Copy() map[string]Info {
	act.mu.RLock()
	defer act.mu.RUnlock()

	accounts := make(map[string]Info)
	for addr, info := range act.info {
		accounts[addr] = info
	}
	return accounts
}

// ValidateNonce validates the nonce for the specified transaction is larger
// than the last nonce used by the account who signed the transaction.
func (act *Accounts) ValidateNonce(tx storage.SignedTx) error {
	from, err := tx.FromAddress()
	if err != nil {
		return err
	}

	var info Info
	act.mu.RLock()
	{
		info = act.info[from]
	}
	act.mu.RUnlock()

	if tx.Nonce < info.Nonce {
		return fmt.Errorf("invalid nonce, got %d, exp >= %d", tx.Nonce, info.Nonce)
	}

	return nil
}

// ApplyMiningReward gives the specififed addr the mining reward.
func (act *Accounts) ApplyMiningReward(minerAddr string) {
	act.mu.Lock()
	defer act.mu.Unlock()

	info := act.info[minerAddr]
	info.Balance += act.genesis.MiningReward

	act.info[minerAddr] = info
}

// ApplyTransaction performs the business logic for applying a transaction
// to the balance sheet.
func (act *Accounts) ApplyTransaction(minerAddr string, tx storage.BlockTx) error {
	from, err := tx.FromAddress()
	if err != nil {
		return fmt.Errorf("invalid signature, %s", err)
	}

	act.mu.Lock()
	defer act.mu.Unlock()
	{
		if from == tx.To {
			return fmt.Errorf("invalid transaction, sending money to yourself, from %s, to %s", from, tx.To)
		}

		fromInfo := act.info[from]
		if tx.Nonce < fromInfo.Nonce {
			return fmt.Errorf("invalid transaction, nonce too small, last %d, tx %d", fromInfo.Nonce, tx.Nonce)
		}

		if tx.Value > act.info[from].Balance {
			return fmt.Errorf("%s has an insufficient balance", from)
		}

		toInfo := act.info[tx.To]
		minerInfo := act.info[minerAddr]

		fromInfo.Balance -= tx.Value
		toInfo.Balance += tx.Value

		fee := tx.Gas + tx.Tip
		minerInfo.Balance += fee
		fromInfo.Balance -= fee

		fromInfo.Nonce = tx.Nonce

		act.info[from] = fromInfo
		act.info[tx.To] = toInfo
		act.info[minerAddr] = minerInfo
	}

	return nil
}
