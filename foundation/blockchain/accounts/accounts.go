// Package accounts maintains account balances and other account information.
package accounts

import (
	"fmt"
	"sync"

	"github.com/ardanlabs/blockchain/foundation/blockchain/genesis"
	"github.com/ardanlabs/blockchain/foundation/blockchain/storage"
)

// Info represents information stored for an individual account.
type Info struct {
	Balance uint
	Nonce   uint
}

// Accounts manages data related to accounts who have transacted on
// the blockchain.
type Accounts struct {
	genesis genesis.Genesis
	info    map[storage.Account]Info
	mu      sync.RWMutex
}

func New(genesis genesis.Genesis) *Accounts {
	accounts := Accounts{
		genesis: genesis,
		info:    make(map[storage.Account]Info),
	}

	for account, balance := range genesis.Balances {
		accounts.info[account] = Info{Balance: balance}
	}

	return &accounts
}

// Reset re-initalizes the accounts back to the genesis information.
func (act *Accounts) Reset() {
	act.mu.Lock()
	defer act.mu.Unlock()

	act.info = make(map[storage.Account]Info)
	for account, balance := range act.genesis.Balances {
		act.info[account] = Info{Balance: balance}
	}
}

// Replace updates the accounts based on the specified accounts.
func (act *Accounts) Replace(accounts *Accounts) {
	act.mu.Lock()
	defer act.mu.Unlock()

	act.info = accounts.info
}

// Remove deletes an account from the accounts.
func (act *Accounts) Remove(account storage.Account) {
	act.mu.Lock()
	defer act.mu.Unlock()

	delete(act.info, account)
}

// Clone makes a copy of the current accounts.
func (act *Accounts) Clone() *Accounts {
	act.mu.RLock()
	defer act.mu.RUnlock()

	accounts := New(act.genesis)
	for account, value := range act.info {
		accounts.info[account] = value
	}
	return accounts
}

// Copy makes a copy of the current information for all accounts.
func (act *Accounts) Copy() map[storage.Account]Info {
	act.mu.RLock()
	defer act.mu.RUnlock()

	accounts := make(map[storage.Account]Info)
	for account, info := range act.info {
		accounts[account] = info
	}
	return accounts
}

// ValidateNonce validates the nonce for the specified transaction is larger
// than the last nonce used by the account who signed the transaction.
func (act *Accounts) ValidateNonce(tx storage.SignedTx) error {
	from, err := tx.FromAccount()
	if err != nil {
		return err
	}

	var info Info
	act.mu.RLock()
	{
		info = act.info[from]
	}
	act.mu.RUnlock()

	if tx.Nonce <= info.Nonce {
		return fmt.Errorf("invalid nonce, got %d, exp >= %d", tx.Nonce, info.Nonce)
	}

	return nil
}

// ApplyMiningReward gives the specififed account the mining reward.
func (act *Accounts) ApplyMiningReward(minerAccount storage.Account) {
	act.mu.Lock()
	defer act.mu.Unlock()

	info := act.info[minerAccount]
	info.Balance += act.genesis.MiningReward

	act.info[minerAccount] = info
}

// ApplyTransaction performs the business logic for applying a transaction
// to the accounts information.
func (act *Accounts) ApplyTransaction(minerAccount storage.Account, tx storage.BlockTx) error {
	from, err := tx.FromAccount()
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
		minerInfo := act.info[minerAccount]

		fromInfo.Balance -= tx.Value
		toInfo.Balance += tx.Value

		fee := tx.Gas + tx.Tip
		minerInfo.Balance += fee
		fromInfo.Balance -= fee

		fromInfo.Nonce = tx.Nonce

		act.info[from] = fromInfo
		act.info[tx.To] = toInfo
		act.info[minerAccount] = minerInfo
	}

	return nil
}
