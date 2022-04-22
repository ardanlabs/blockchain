// Package database handles all the lower level support for maintaining the
// blockchain on disk and maintaining an in memory databse of account information.
package database

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"sync"

	"github.com/ardanlabs/blockchain/foundation/blockchain/genesis"
)

// Database manages data related to accounts who have transacted on the blockchain.
type Database struct {
	mu sync.RWMutex

	genesis     genesis.Genesis
	latestBlock Block
	accounts    map[AccountID]Account

	dbPath string
	dbFile *os.File
}

// New constructs a new database and applies account genesis information and
// reads/writes the blockchain database on disk if a dbPath is provided.
func New(dbPath string, genesis genesis.Genesis, evHandler func(v string, args ...any)) (*Database, error) {
	var dbFile *os.File

	// If not path is provided, the database will sync only to the
	// genesis information.
	if dbPath != "" {
		var err error
		dbFile, err = os.OpenFile(dbPath, os.O_APPEND|os.O_RDWR|os.O_CREATE, 0600)
		if err != nil {
			return nil, err
		}
	}

	db := Database{
		genesis:  genesis,
		accounts: make(map[AccountID]Account),
		dbPath:   dbPath,
		dbFile:   dbFile,
	}

	// Read all the blocks from disk if a path is provided.
	var blocks []Block
	if dbFile != nil {
		var err error
		blocks, err = db.ReadAllBlocks(evHandler, true)
		if err != nil {
			return nil, err
		}
	}

	// Update the database with account balance information from genesis.
	for accountStr, balance := range genesis.Balances {
		accountID, err := ToAccountID(accountStr)
		if err != nil {
			return nil, err
		}
		db.accounts[accountID] = Account{Balance: balance}
	}

	// Set the current latest block in the chain.
	if len(blocks) > 0 {
		db.latestBlock = blocks[len(blocks)-1]
	}

	// Update the databse with account balance information from blocks.
	for _, block := range blocks {
		for _, tx := range block.Trans.Values() {
			db.ApplyTransaction(block.Header.BeneficiaryID, tx)
		}
		db.ApplyMiningReward(block.Header.BeneficiaryID)
	}

	return &db, nil
}

// Close closes the open blocks database.
func (db *Database) Close() {
	db.mu.Lock()
	defer db.mu.Unlock()

	db.dbFile.Close()
}

// Reset re-initalizes the database back to the genesis state.
func (db *Database) Reset() error {
	db.mu.Lock()
	defer db.mu.Unlock()

	// Close and remove the current file.
	db.dbFile.Close()
	if err := os.Remove(db.dbPath); err != nil {
		return err
	}

	// Open a new blockchain database file with create.
	dbFile, err := os.OpenFile(db.dbPath, os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		return err
	}
	db.dbFile = dbFile

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
func (db *Database) ApplyMiningReward(beneficiaryID AccountID) {
	db.mu.Lock()
	defer db.mu.Unlock()

	account := db.accounts[beneficiaryID]
	account.Balance += db.genesis.MiningReward

	db.accounts[beneficiaryID] = account
}

// ApplyTransaction performs the business logic for applying a transaction
// to the database.
func (db *Database) ApplyTransaction(beneficiaryID AccountID, tx BlockTx) error {

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
		bnfc := db.accounts[beneficiaryID]

		// The account needs to pay the gas fee regardless. Take the
		// remaining balance if the account doesn't hold enough for the
		// full amount of gas.
		gasFee := tx.GasPrice * tx.GasUnits
		if gasFee > from.Balance {
			gasFee = from.Balance
		}
		from.Balance -= gasFee
		bnfc.Balance += gasFee

		// Make sure these changes get applied.
		db.accounts[fromID] = from
		db.accounts[beneficiaryID] = bnfc

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

			if from.Balance == 0 || (tx.Value+tx.Tip) > from.Balance {
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
		db.accounts[beneficiaryID] = bnfc
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
func (db *Database) Write(block BlockFS) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	// Marshal the block for writing to disk.
	blockFSJson, err := json.Marshal(block)
	if err != nil {
		return err
	}

	// Write the new block to the chain on disk.
	if _, err := db.dbFile.Write(append(blockFSJson, '\n')); err != nil {
		return err
	}

	return nil
}

// ReadAllBlocks loads all existing blocks from storage into memory. In a real
// world situation this would require a lot of memory.
func (db *Database) ReadAllBlocks(evHandler func(v string, args ...any), validate bool) ([]Block, error) {
	dbFile, err := os.Open(db.dbPath)
	if err != nil {
		return nil, err
	}
	defer dbFile.Close()

	var blocks []Block
	var latestBlock Block
	scanner := bufio.NewScanner(dbFile)
	for scanner.Scan() {
		if err := scanner.Err(); err != nil {
			return nil, err
		}

		var blockFS BlockFS
		if err := json.Unmarshal(scanner.Bytes(), &blockFS); err != nil {
			return nil, err
		}

		block, err := ToBlock(blockFS)
		if err != nil {
			return nil, err
		}

		// We want to skip the block validation for query and retrieve operations.
		if validate {
			if err := block.ValidateBlock(latestBlock, evHandler); err != nil {
				return nil, err
			}
		}

		blocks = append(blocks, block)
		latestBlock = block
	}

	return blocks, nil
}
