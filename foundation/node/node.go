// Package node is the implementation of our blockchain DB.
package node

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sync"
	"time"
)

// ErrNoTransactions is returned when a block is requested to be created
// and there are no transactions.
var ErrNoTransactions = errors.New("no transactions in mempool")

// EventHandler defines a function that is called when events
// occur in the processing of persisting blocks.
type EventHandler func(v string)

// Config represents the configuration required to start
// the blockchain node.
type Config struct {
	IPPort          string
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
	dbPath      string
	file        *os.File
	mu          sync.Mutex
	blockWriter *blockWriter
	ipPort      string
	knownPeers  map[string]struct{}
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
	blocks, err := loadBlocksFromDisk(cfg.DBPath)
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

	// Convert the list of peers to the internal map.
	peers := make(map[string]struct{})
	for _, peer := range cfg.KnownPeers {
		peers[peer] = struct{}{}
	}

	// Create the chain with no transactions currently in memory.
	n := Node{
		genesis:     genesis,
		latestBlock: latestBlock,
		balances:    balances,
		dbPath:      cfg.DBPath,
		file:        file,
		ipPort:      cfg.IPPort,
		knownPeers:  peers,
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
	if _, err := n.writeNewBlockFromMempool(); err != nil {
		if !errors.Is(err, ErrNoTransactions) {
			return err
		}
	}

	return nil
}

// =============================================================================

// AddTransaction appends a new transactions to the mempool.
func (n *Node) AddTransaction(tx Tx) error {
	n.mu.Lock()
	defer n.mu.Unlock()

	// Append the transaction to the in-memory store.
	n.txMempool = append(n.txMempool, tx)

	return nil
}

// WriteNewBlockFromMempool writes the current transactions from the
// memory pool to disk.
func (n *Node) WriteNewBlockFromMempool() (Block, error) {
	n.mu.Lock()
	defer n.mu.Unlock()

	return n.writeNewBlockFromMempool()
}

// =============================================================================

// AddPeerNode adds an address to the list of peers.
func (n *Node) AddPeerNode(ipPort string) error {
	n.mu.Lock()
	defer n.mu.Unlock()

	// Don't add this node to the known peer list.
	if ipPort == n.ipPort {
		return errors.New("already exists")
	}

	if _, exists := n.knownPeers[ipPort]; !exists {
		n.knownPeers[ipPort] = struct{}{}
	}

	return nil
}

// KnownPeersList retrieves information about the peer for updating
// the known peer list and their current block number.
func (n *Node) KnownPeersList() map[string]struct{} {
	n.mu.Lock()
	defer n.mu.Unlock()

	// Return a copy of the list and remove this node
	// from the list.
	peers := make(map[string]struct{})
	for k := range n.knownPeers {
		if k != n.ipPort {
			peers[k] = struct{}{}
		}
	}

	return peers
}

// =============================================================================

// Genesis returns a copy of the genesis information.
func (n *Node) Genesis() Genesis {
	n.mu.Lock()
	defer n.mu.Unlock()

	return n.genesis
}

// LatestBlock returns the current hash of the latest block.
func (n *Node) LatestBlock() Block {
	n.mu.Lock()
	defer n.mu.Unlock()

	return n.latestBlock
}

// Mempool returns a copy of the mempool.
func (n *Node) Mempool() []Tx {
	n.mu.Lock()
	defer n.mu.Unlock()

	cpy := make([]Tx, len(n.txMempool))
	copy(cpy, n.txMempool)
	return cpy
}

// Balances returns the set of balances by account. If the account
// is empty, all balances are returned.
func (n *Node) Balances(account string) map[string]uint {
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

// LastestBlock represents the latest block in the DB.
const LastestBlock = ^uint64(0) >> 1

// BlocksByNumber returns the set of blocks based on block numbers.
func (n *Node) BlocksByNumber(from uint64, to uint64) []Block {
	blocks, err := loadBlocksFromDisk(n.dbPath)
	if err != nil {
		return nil
	}

	if from == LastestBlock {
		from = blocks[len(blocks)-1].Header.Number
		to = from
	}

	var out []Block
	for _, block := range blocks {
		if block.Header.Number >= from && block.Header.Number <= to {
			out = append(out, block)
		}
	}

	return out
}

// BlocksByAccount returns the set of blocks by account. If the account
// is empty, all blocks are returned.
func (n *Node) BlocksByAccount(account string) []Block {
	blocks, err := loadBlocksFromDisk(n.dbPath)
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

// writeNewBlockFromPeer writes the specified peer block to disk.
// It assumes it's always inside a mutex lock.
func (n *Node) writeNewBlockFromPeer(peerBlock PeerBlock) (Block, error) {

	// Convert the peer block to a block.
	block := peerBlock.ToBlock()

	// Validate the hash is correct.
	hash := peerBlock.Header.ThisBlock
	if block.Hash() != hash {
		return Block{}, errors.New("hash does not match")
	}

	// Validate the block number is the next in sequence.
	nextNumber := n.latestBlock.Header.Number + 1
	if block.Header.Number != nextNumber {
		return Block{}, fmt.Errorf("wrong block number, got %d, exp %d", peerBlock.Header.Number, nextNumber)
	}

	// Validate the prev block hash matches our latest node.
	if peerBlock.Header.PrevBlock != n.LatestBlock().Hash() {
		return Block{}, fmt.Errorf("prev block doesn't match our latest, got %s, exp %s", peerBlock.Header.PrevBlock, n.LatestBlock().Hash())
	}

	blockFS := BlockFS{
		Hash:  hash,
		Block: block,
	}

	blockFSJson, err := json.Marshal(blockFS)
	if err != nil {
		return Block{}, err
	}

	if _, err := n.file.Write(append(blockFSJson, '\n')); err != nil {
		return Block{}, err
	}

	n.latestBlock = blockFS.Block

	if err := n.applyTransToBalances(block.Transactions); err != nil {
		return blockFS.Block, err
	}

	return blockFS.Block, nil
}

// writeNewBlock writes the current transaction memory pool to disk.
// It assumes it's always inside a mutex lock.
func (n *Node) writeNewBlockFromMempool() (Block, error) {
	if len(n.txMempool) == 0 {
		return Block{}, ErrNoTransactions
	}

	// If the transaction can't be applied to the balance,
	// mark the transaction as failed.
	for i := range n.txMempool {
		if err := n.applyTranToBalance(n.txMempool[i]); err != nil {
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
