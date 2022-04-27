// Package database handles all the lower level support for maintaining the
// blockchain on disk and maintaining an in memory databse of account information.
package database

import (
	"fmt"
	"sync"

	"github.com/ardanlabs/blockchain/foundation/blockchain/genesis"
)

// Serializer interface represents the behavior required to be implemented by any
// package providing support for storing and reading the blockchain.
type Serializer interface {
	Write(blockData BlockData) error
	GetBlock(num uint64) (BlockData, error)
	ForEach() Iterator
	Close() error
	Reset() error
}

// Iterator interface represents the behavior required to be implemented by any
// package providing support to iterate over the blocks.
type Iterator interface {
	Next() (BlockData, error)
	Done() bool
}

// =============================================================================

type DatabaseIterator struct {
	iterator Iterator
}

// Next retrieves the next block from disk.
func (di *DatabaseIterator) Next() (Block, error) {
	blockData, err := di.iterator.Next()
	if err != nil {
		return Block{}, err
	}

	return ToBlock(blockData)
}

// Done returns the end of chain value.
func (di *DatabaseIterator) Done() bool {
	return di.iterator.Done()
}

// Database manages data related to accounts who have transacted on the blockchain.
type Database struct {
	mu sync.RWMutex

	genesis     genesis.Genesis
	latestBlock Block
	accounts    map[AccountID]Account

	serializer Serializer
}

// New constructs a new database and applies account genesis information and
// reads/writes the blockchain database on disk if a dbPath is provided.
func New(genesis genesis.Genesis, serializer Serializer, evHandler func(v string, args ...any)) (*Database, error) {

	db := Database{
		genesis:    genesis,
		accounts:   make(map[AccountID]Account),
		serializer: serializer,
	}

	// Update the database with account balance information from genesis.
	for accountStr, balance := range genesis.Balances {
		accountID, err := ToAccountID(accountStr)
		if err != nil {
			return nil, err
		}
		db.accounts[accountID] = Account{Balance: balance}
	}

	// Read all the blocks from disk if a path is provided.
	var latestBlock Block

	iter := db.serializer.ForEach()
	for blockData, err := iter.Next(); !iter.Done(); blockData, err = iter.Next() {
		if err != nil {
			return nil, err
		}

		block, err := ToBlock(blockData)
		if err != nil {
			return nil, err
		}

		if err := block.ValidateBlock(latestBlock, evHandler); err != nil {
			return nil, err
		}

		// Update the database with account balance information from genesis.
		for accountStr, balance := range genesis.Balances {
			accountID, err := ToAccountID(accountStr)
			if err != nil {
				return nil, err
			}
			db.accounts[accountID] = Account{Balance: balance}
		}
		for _, tx := range block.Trans.Values() {
			db.ApplyTransaction(block, tx)
		}
		db.ApplyMiningReward(block)

		latestBlock = block
	}
	return &db, nil
}

// Close closes the open blocks database.
func (db *Database) Close() {
	db.serializer.Close()
}

// Reset re-initalizes the database back to the genesis state.
func (db *Database) Reset() error {
	db.serializer.Reset()

	// Initalizes the database back to the genesis information.
	db.latestBlock = Block{}
	db.accounts = make(map[AccountID]Account)
	for accountStr, balance := range db.genesis.Balances {
		accountID, err := ToAccountID(accountStr)
		if err != nil {
			return err
		}

		db.accounts[accountID] = Account{Balance: balance}
	}

	return nil
}

// Remove deletes an account from the database.
func (db *Database) Remove(accountID AccountID) {
	db.mu.Lock()
	defer db.mu.Unlock()

	delete(db.accounts, accountID)
}

// CopyAccounts makes a copy of the current accounts in the database.
func (db *Database) CopyAccounts() map[AccountID]Account {
	db.mu.RLock()
	defer db.mu.RUnlock()

	accounts := make(map[AccountID]Account)
	for accountID, account := range db.accounts {
		accounts[accountID] = account
	}
	return accounts
}

// ApplyMiningReward gives the specififed account the mining reward.
func (db *Database) ApplyMiningReward(block Block) {
	db.mu.Lock()
	defer db.mu.Unlock()

	account := db.accounts[block.Header.BeneficiaryID]
	account.Balance += block.Header.MiningReward

	db.accounts[block.Header.BeneficiaryID] = account
}

// ApplyTransaction performs the business logic for applying a transaction
// to the database.
func (db *Database) ApplyTransaction(block Block, tx BlockTx) error {

	// Capture the from address from the signature of the transaction.
	fromID, err := tx.FromAccount()
	if err != nil {
		return fmt.Errorf("invalid signature, %s", err)
	}

	db.mu.Lock()
	defer db.mu.Unlock()
	{
		// Capture these accounts from the database.
		from := db.accounts[fromID]
		to := db.accounts[tx.ToID]
		bnfc := db.accounts[block.Header.BeneficiaryID]

		// The account needs to pay the gas fee regardless. Take the
		// remaining balance if the account doesn't hold enough for the
		// full amount of gas. This is the only way to stop bad actors.
		gasFee := tx.GasPrice * tx.GasUnits
		if gasFee > from.Balance {
			gasFee = from.Balance
		}
		from.Balance -= gasFee
		bnfc.Balance += gasFee

		// Make sure these changes get applied.
		db.accounts[fromID] = from
		db.accounts[block.Header.BeneficiaryID] = bnfc

		// Perform basic accounting checks.
		{
			if tx.ChainID != db.genesis.ChainID {
				return fmt.Errorf("transaction invalid, wrong chain id, got %d, exp %d", tx.ChainID, db.genesis.ChainID)
			}

			if fromID == tx.ToID {
				return fmt.Errorf("transaction invalid, sending money to yourself, from %s, to %s", fromID, tx.ToID)
			}

			if tx.Nonce <= from.Nonce {
				return fmt.Errorf("transaction invalid, nonce too small, current %d, provided %d", from.Nonce, tx.Nonce)
			}

			if from.Balance == 0 || from.Balance < (tx.Value+tx.Tip) {
				return fmt.Errorf("transaction invalid, insufficient funds, bal %d, needed %d", from.Balance, (tx.Value + tx.Tip))
			}
		}

		// Update the balances between the two parties.
		from.Balance -= tx.Value
		to.Balance += tx.Value

		// Give the beneficiary the tip.
		from.Balance -= tx.Tip
		bnfc.Balance += tx.Tip

		// Update the nonce for the next transaction check.
		from.Nonce = tx.Nonce

		// Update the final changes to these accounts.
		db.accounts[fromID] = from
		db.accounts[tx.ToID] = to
		db.accounts[block.Header.BeneficiaryID] = bnfc
	}

	return nil
}

// UpdateLatestBlock provides safe access to update the latest block.
func (db *Database) UpdateLatestBlock(block Block) {
	db.mu.Lock()
	defer db.mu.Unlock()

	db.latestBlock = block
}

// LatestBlock returns the latest block.
func (db *Database) LatestBlock() Block {
	db.mu.RLock()
	defer db.mu.RUnlock()

	return db.latestBlock
}

// Write adds a new block to the chain.
func (db *Database) Write(block Block) error {
	return db.serializer.Write(NewBlockData(block))
}

// ForEach returns an iterator to walk through all the blocks
// starting with block number 1.
func (db *Database) ForEach() DatabaseIterator {
	return DatabaseIterator{iterator: db.serializer.ForEach()}
}

// GetBlock searches the blockchain on disk to locate and return the
// contents of the specified block by number.
func (db *Database) GetBlock(num uint64) (Block, error) {
	blockData, err := db.serializer.GetBlock(num)
	if err != nil {
		return Block{}, err
	}
	return ToBlock(blockData)
}
