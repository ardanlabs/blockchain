package node

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sync"
	"time"
)

// Config represents the configuration required to start
// the blockchain node.
type Config struct {
	DBPath          string
	PersistInterval time.Duration
	KnownPeers      []string
	EvHandler       EventHandler
}

// Node manages the blockchain database.
type Node struct {
	genesis     Genesis
	txMempool   []Tx
	latestBlock Block
	balances    map[string]uint
	knownPeers  []string
	dbPath      string
	file        *os.File
	mu          sync.Mutex
	blockWriter *blockWriter
}

// New constructs a new blockchain for data management.
func New(cfg Config) (*Node, error) {

	// Load the genesis file to get starting balances for
	// founders of the block chain.
	genesis, err := loadGenesis()
	if err != nil {
		return nil, err
	}

	// Load the current set of recorded transactions.
	blocks, err := loadBlocks(cfg.DBPath)
	if err != nil {
		return nil, err
	}

	// Make a copy of the genesis balances for the next step.
	balances := make(map[string]uint)
	for key, value := range genesis.Balances {
		balances[key] = value
	}

	// Open the transaction database file.
	file, err := os.OpenFile(cfg.DBPath, os.O_APPEND|os.O_RDWR, 0600)
	if err != nil {
		return nil, err
	}

	// Capture the hash of the latest block.
	var latestBlock Block
	if len(blocks) > 0 {
		latestBlock = blocks[len(blocks)-1]
	}

	// Create the chain with no transactions currently in memory.
	n := Node{
		genesis:     genesis,
		latestBlock: latestBlock,
		balances:    balances,
		knownPeers:  cfg.KnownPeers,
		dbPath:      cfg.DBPath,
		file:        file,
	}

	// Apply the transactions to the initial genesis balances, adding new
	// accounts as it is processed.
	for _, block := range blocks {
		if err := n.applyTransToBalances(block.Transactions); err != nil {
			return nil, err
		}
	}

	// Start the block writer.
	n.blockWriter = newBlockWriter(&n, cfg.PersistInterval, cfg.EvHandler)

	return &n, nil
}

// Shutdown cleanly brings the node down.
func (n *Node) Shutdown() error {
	n.mu.Lock()
	defer func() {
		n.file.Close()
		n.mu.Unlock()
	}()

	n.blockWriter.shutdown()

	// Persist the remaining transactions to disk.
	if _, err := n.writeBlock(); err != nil {
		if !errors.Is(err, ErrNoTransactions) {
			return err
		}
	}

	return nil
}

// AddTransaction appends a new transactions to the mempool.
func (n *Node) AddTransaction(tx Tx) error {
	n.mu.Lock()
	defer n.mu.Unlock()

	// Append the transaction to the in-memory store.
	n.txMempool = append(n.txMempool, tx)

	return nil
}

// WriteBlock writes the current transactions from the
// memory pool to disk.
func (n *Node) WriteBlock() (Block, error) {
	n.mu.Lock()
	defer n.mu.Unlock()

	return n.writeBlock()
}

// =============================================================================

// QueryGenesis returns a copy of the genesis information.
func (n *Node) QueryGenesis() Genesis {
	n.mu.Lock()
	defer n.mu.Unlock()

	return n.genesis
}

// QueryLatestBlock returns the current hash of the latest block.
func (n *Node) QueryLatestBlock() Block {
	n.mu.Lock()
	defer n.mu.Unlock()

	return n.latestBlock
}

// QueryMempool returns a copy of the mempool.
func (n *Node) QueryMempool() []Tx {
	n.mu.Lock()
	defer n.mu.Unlock()

	cpy := make([]Tx, len(n.txMempool))
	copy(cpy, n.txMempool)
	return cpy
}

// QueryBalances returns the set of balances by account. If the account
// is empty, all balances are returned.
func (n *Node) QueryBalances(account string) map[string]uint {
	n.mu.Lock()
	defer n.mu.Unlock()

	balances := make(map[string]uint)
	for act, bal := range n.balances {
		if account == "" || account == act {
			balances[act] = bal
		}
	}

	return balances
}

// QueryBlocks returns the set of blocks by account. If the account
// is empty, all blocks are returned.
func (n *Node) QueryBlocks(account string) []Block {
	blocks, err := loadBlocks(n.dbPath)
	if err != nil {
		return nil
	}

	if account == "" {
		return blocks
	}

	var out []Block
	for _, block := range blocks {
		for _, tran := range block.Transactions {
			if tran.From == account || tran.To == account {
				out = append(out, block)
			}
		}
	}

	return out
}

// =============================================================================

// ErrNoTransactions is returned when a block is requested to be created
// and there are no transactions.
var ErrNoTransactions = errors.New("no transactions in mempool")

// writeBlock writes the current transaction memory pool to disk.
// It assumes it's always inside a mutex lock.
func (n *Node) writeBlock() (Block, error) {
	if len(n.txMempool) == 0 {
		return Block{}, ErrNoTransactions
	}

	// If the transaction can't be applied to the balance,
	// mark the transaction as failed.
	for i := range n.txMempool {
		if err := n.validateTransaction(n.txMempool[i]); err != nil {
			n.txMempool[i].Status = TxStatusError
			n.txMempool[i].StatusInfo = err.Error()
			continue
		}
		n.txMempool[i].Status = TxStatusAccepted
	}

	blockFS, err := NewBlockFS(n.latestBlock, n.txMempool)
	if err != nil {
		return Block{}, err
	}

	blockFSJson, err := json.Marshal(blockFS)
	if err != nil {
		return Block{}, err
	}

	if _, err := n.file.Write(append(blockFSJson, '\n')); err != nil {
		return Block{}, err
	}

	n.latestBlock = blockFS.Block
	n.txMempool = []Tx{}

	return blockFS.Block, nil
}

// validateTransaction performs integrity checks on a transaction.
func (n *Node) validateTransaction(tx Tx) error {

	// Validate the transaction can be applied to the balance,
	// checking for things like insufficient funds.
	if err := n.applyTranToBalance(tx); err != nil {
		return err
	}

	return nil
}

// applyTransToBalances applies the transactions to the specified
// balances, adding new accounts as they are found.
func (n *Node) applyTransToBalances(txs []Tx) error {
	for _, tx := range txs {
		n.applyTranToBalance(tx)
	}

	return nil
}

// applyTranToBalance performs the business logic for applying a transaction to
// the balance sheet.
func (n *Node) applyTranToBalance(tx Tx) error {
	if tx.Status == TxStatusError {
		return nil
	}

	if tx.Data == TxDataReward {
		n.balances[tx.To] += tx.Value
		return nil
	}

	if tx.From == tx.To {
		return fmt.Errorf("invalid transaction, do you mean to give a reward, from %s, to %s", tx.From, tx.To)
	}

	if tx.Value > n.balances[tx.From] {
		return fmt.Errorf("%s has an insufficient balance", tx.From)
	}

	n.balances[tx.From] -= tx.Value
	n.balances[tx.To] += tx.Value

	return nil
}
