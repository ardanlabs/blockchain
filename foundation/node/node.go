// Package node is the implementation of our blockchain DB.
package node

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"os"
	"sync"
	"time"
)

// zeroHash represents a has code of zeros.
const zeroHash = "00000000000000000000000000000000"

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
	ipPort      string
	knownPeers  map[string]struct{}
	bcWorker    *bcWorker
}

// New constructs a new blockchain for data management.
func New(cfg Config) (*Node, error) {

	// Load the genesis file to get starting balances for
	// founders of the block chain.
	genesis, err := loadGenesis()
	if err != nil {
		return nil, err
	}

	// Load the blockchain from disk. This would not make sense
	// with the current Ethereum blockchain. Ours is small.
	blocks, err := loadBlocksFromDisk(cfg.DBPath)
	if err != nil {
		return nil, err
	}

	// Keep the latest block from the blockchain.
	var latestBlock Block
	if len(blocks) > 0 {
		latestBlock = blocks[len(blocks)-1]
	}

	// Save the list of known peers.
	peers := make(map[string]struct{})
	for _, peer := range cfg.KnownPeers {
		peers[peer] = struct{}{}
	}

	// Apply the genesis balances to the balance sheet.
	balances := make(map[string]uint)
	for key, value := range genesis.Balances {
		balances[key] = value
	}

	// Update the balance sheet by processing all the transactions in
	// the set of blocks.
	for _, block := range blocks {
		if err := applyTransToBalances(balances, block.Transactions); err != nil {
			return nil, err
		}
	}

	// Open the blockchain database file for processing.
	file, err := os.OpenFile(cfg.DBPath, os.O_APPEND|os.O_RDWR, 0600)
	if err != nil {
		return nil, err
	}

	// Create the node to provide support for managing the blockchain.
	node := Node{
		genesis:     genesis,
		latestBlock: latestBlock,
		balances:    balances,
		dbPath:      cfg.DBPath,
		file:        file,
		ipPort:      cfg.IPPort,
		knownPeers:  peers,
	}

	// Start the blockchain worker.
	node.bcWorker = newBCWorker(&node, cfg.PersistInterval, cfg.EvHandler)

	return &node, nil
}

// Shutdown cleanly brings the node down.
func (n *Node) Shutdown() error {
	n.mu.Lock()
	defer func() {
		n.file.Close()
		n.mu.Unlock()
	}()

	// Stop all blockchain writing activity.
	n.bcWorker.shutdown()

	return nil
}

// =============================================================================

// SignalBlockWork signals the blockchain worker to perform work.
func (n *Node) SignalBlockWork(ctx context.Context) error {
	return n.bcWorker.SignalBlockWork(ctx)
}

// SignalAddTransactions signals a new transaction to be added to the mempool.
func (n *Node) SignalAddTransactions(ctx context.Context, txs []Tx) error {
	return n.bcWorker.SignalAddTransactions(ctx, txs)
}

// =============================================================================

// CopyGenesis returns a copy of the genesis information.
func (n *Node) CopyGenesis() Genesis {
	n.mu.Lock()
	defer n.mu.Unlock()

	return n.genesis
}

// CopyLatestBlock returns the current hash of the latest block.
func (n *Node) CopyLatestBlock() Block {
	n.mu.Lock()
	defer n.mu.Unlock()

	return n.latestBlock
}

// CopyMempool returns a copy of the mempool.
func (n *Node) CopyMempool() []Tx {
	n.mu.Lock()
	defer n.mu.Unlock()

	cpy := make([]Tx, len(n.txMempool))
	copy(cpy, n.txMempool)
	return cpy
}

// CopyBalances returns a copy of the set of balances by account.
func (n *Node) CopyBalances() map[string]uint {
	n.mu.Lock()
	defer n.mu.Unlock()

	balances := make(map[string]uint)
	for act, bal := range n.balances {
		balances[act] = bal
	}

	return balances
}

// CopyKnownPeersList retrieves information about the peer for updating
// the known peer list and their current block number.
func (n *Node) CopyKnownPeersList() map[string]struct{} {
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

// QueryLastest represents to query the latest block in the chain.
const QueryLastest = ^uint64(0) >> 1

// QueryBalances returns a copy of the set of balances by account.
// If the account parameter is empty, all balances are returned.
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

// QueryBlocksByNumber returns the set of blocks based on block numbers. This
// function reads the blockchain from disk first.
func (n *Node) QueryBlocksByNumber(from uint64, to uint64) []Block {
	blocks, err := loadBlocksFromDisk(n.dbPath)
	if err != nil {
		return nil
	}

	if from == QueryLastest {
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

// QueryBlocksByAccount returns the set of blocks by account. If the account
// is empty, all blocks are returned. This function reads the blockchain
// from disk first.
func (n *Node) QueryBlocksByAccount(account string) []Block {
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
			if tran.Record.From == account || tran.Record.To == account {
				out = append(out, block)
			}
		}
	}

	return out
}

// QueryMempool returns the set of transaction for the specified status.
func (n *Node) QueryMempool(status string) []Tx {
	n.mu.Lock()
	defer n.mu.Unlock()

	var txs []Tx
	for _, tx := range txs {
		if tx.Status == status {
			txs = append(txs, tx)
		}
	}

	return txs
}

// =============================================================================

// addTransactionsUnderLock appends a new transactions to the mempool under a lock.
func (n *Node) addTransactions(txs []Tx) {
	n.mu.Lock()
	defer n.mu.Unlock()

	n.txMempool = append(n.txMempool, txs...)
}

// addTransactionsUnderLock appends a new transactions to the mempool under a lock.
func (n *Node) updateTransactions(tx []Tx, status string) {
	n.mu.Lock()
	defer n.mu.Unlock()

	// WE NEED TO FIND A TRANSACTION BASED ON THE HASH.
	// WE NEED A MAP NOW.
}

// addPeerNode adds an address to the list of peers.
func (n *Node) addPeerNode(ipPort string) error {
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

// copyTransactions retrieves a copy of transactions from the mempool that match
// the specified statuses. This is unexported and is called inside of a lock.
func (n *Node) copyTransactions(statuses ...string) []Tx {
	var txs []Tx
	for _, tx := range n.txMempool {
		for _, status := range statuses {
			if tx.Status == status {
				txs = append(txs, tx)
				break
			}
		}
	}

	return txs
}

// =============================================================================

// writeNewBlockFromPeer writes the specified peer block to disk.
func (n *Node) writeNewBlockFromPeer(peerBlock PeerBlock) (Block, error) {
	n.mu.Lock()
	defer n.mu.Unlock()

	// Convert the peer block to a block.
	block := peerBlock.ToBlock()

	// Validate the hash is correct.
	hash := peerBlock.Header.ThisBlock
	if block.Hash() != hash {
		return Block{}, errors.New("generated hash does not match peer")
	}

	// Validate the hash matches the POW puzzle.
	if !isHashSolved(hash) {
		return Block{}, errors.New("hash does not match POW")
	}

	// Validate the block number is the next in sequence.
	nextNumber := n.latestBlock.Header.Number + 1
	if block.Header.Number != nextNumber {
		return Block{}, fmt.Errorf("wrong block number, got %d, exp %d", peerBlock.Header.Number, nextNumber)
	}

	// Validate the prev block hash matches our latest node.
	if peerBlock.Header.PrevBlock != n.latestBlock.Hash() {
		return Block{}, fmt.Errorf("prev block doesn't match our latest, got %s, exp %s", peerBlock.Header.PrevBlock, n.latestBlock.Hash())
	}

	// Write the new block to the chain on disk.
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

	// Update the state of the node.
	n.latestBlock = blockFS.Block
	if err := applyTransToBalances(n.balances, block.Transactions); err != nil {
		return blockFS.Block, err
	}

	return blockFS.Block, nil
}

// writeNewBlockFromTransactions writes the published transaction from the
// memory pool to disk.
func (n *Node) writeNewBlockFromTransactions() (Block, error) {
	n.mu.Lock()
	defer n.mu.Unlock()

	// Get the set of new and published transactions.
	txs := n.copyTransactions(TxStatusNew, TxStatusPublished)
	if len(txs) == 0 {
		return Block{}, ErrNoTransactions
	}

	// Create a new block.
	nb := NewBlock(n.latestBlock, txs)

	// Get a copy of the balance sheet.
	balances := make(map[string]uint)
	for act, bal := range n.balances {
		balances[act] = bal
	}

	// Apply the transactions to that copy, setting status information.
	for i, tx := range nb.Transactions {
		if err := applyTranToBalance(balances, tx); err != nil {
			nb.Transactions[i].Status = TxStatusError
			nb.Transactions[i].StatusInfo = err.Error()
			continue
		}
		nb.Transactions[i].Status = TxStatusAccepted
	}

	// Give ourselves 10 seconds to perform the POW.
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Attempt to create a new BlockFS by solving the POW puzzle.
	blockFS, err := performPOW(ctx, nb)
	if err != nil {
		return Block{}, err
	}

	// Write the new block to the chain on disk.
	blockFSJson, err := json.Marshal(blockFS)
	if err != nil {
		return Block{}, err
	}
	if _, err := n.file.Write(append(blockFSJson, '\n')); err != nil {
		return Block{}, err
	}

	// Update the state of the node.
	n.balances = balances
	n.latestBlock = blockFS.Block
	n.txMempool = []Tx{}

	return blockFS.Block, nil
}

// =============================================================================

// generateHash takes a value and produces a 32 byte hash.
func generateHash(v interface{}) string {
	blockJson, err := json.Marshal(v)
	if err != nil {
		return zeroHash
	}

	hash := sha256.Sum256(blockJson)
	return hex.EncodeToString(hash[:])
}

// generateNonce generates a new nonce (number once).
func generateNonce() uint64 {
	const max = 1_000_000

	nBig, err := rand.Int(rand.Reader, big.NewInt(max))
	if err != nil {
		return 0
	}

	return nBig.Uint64()
}
