package node

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sync"
	"time"
)

/*
	Need a way to identify my chain is no longer the valid chain, re-sync.
	Need a way to validate a new block against the entire known blockchain.
	Need a wallet to sign transactions properly.
	Maybe adjust difficulty based on time to mine. Currently hardcoded to 6 zeros.
	Add transaction for the reward in the block being mined.
	Add fees to transactions.
*/

// zeroHash represents a has code of zeros.
const zeroHash = "00000000000000000000000000000000"

// ErrNotEnoughTransactions is returned when a block is requested to be created
// and there are not enough transactions.
var ErrNotEnoughTransactions = errors.New("not enough transactions in mempool")

// EventHandler defines a function that is called when events
// occur in the processing of persisting blocks.
type EventHandler func(v string, args ...interface{})

// Config represents the configuration required to start
// the blockchain node.
type Config struct {
	IPPort     string
	DBPath     string
	KnownPeers []string
	EvHandler  EventHandler
}

// Node manages the blockchain database.
type Node struct {
	genesis     Genesis
	txMempool   map[string]Tx
	latestBlock Block
	balances    map[string]uint
	dbPath      string
	file        *os.File
	mu          sync.Mutex
	ipPort      string
	knownPeers  map[string]struct{}
	bcWorker    *bcWorker
	evHandler   EventHandler
}

// New constructs a new blockchain for data management.
func New(cfg Config) (*Node, error) {

	// Build a safe event handler function for use.
	ev := func(v string, args ...interface{}) {
		if cfg.EvHandler != nil {
			cfg.EvHandler(v, args...)
		}
	}

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
		if err := applyTransactionsToBalances(balances, block.Transactions); err != nil {
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
		txMempool:   make(map[string]Tx),
		latestBlock: latestBlock,
		balances:    balances,
		dbPath:      cfg.DBPath,
		file:        file,
		ipPort:      cfg.IPPort,
		knownPeers:  peers,
		evHandler:   ev,
	}

	ev("node: Started: blocks[%d]", latestBlock.Header.Number)

	// Start the blockchain worker.
	node.bcWorker = newBCWorker(&node, cfg.EvHandler)

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

// SignalMining sends a signal to the mining G to start.
func (n *Node) SignalMining() {
	n.bcWorker.signalStartMining()
}

// SignalCancelMining sends a signal to the mining G to stop.
func (n *Node) SignalCancelMining() {
	n.bcWorker.signalCancelMining()
}

// =============================================================================

// AddTransactions appends a new transactions to the mempool.
func (n *Node) AddTransactions(txs []Tx, share bool) {
	n.mu.Lock()
	defer n.mu.Unlock()

	n.evHandler("node: AddTransactions: started : txrs[%d]", len(txs))
	defer n.evHandler("node: AddTransactions: completed")

	n.evHandler("node: AddTransactions: before: mempool[%d]", len(n.txMempool))
	for _, tx := range txs {
		if _, exists := n.txMempool[tx.ID]; !exists {
			n.txMempool[tx.ID] = tx
		}
	}
	n.evHandler("node: AddTransactions: after: mempool[%d]", len(n.txMempool))

	if share {
		n.evHandler("node: AddTransactions: signal tx sharing")
		n.bcWorker.signalShareTransactions(txs)
	}

	if len(n.txMempool) >= 2 {
		n.evHandler("node: AddTransactions: signal mining")
		n.bcWorker.signalStartMining()
	}
}

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

	cpy := make([]Tx, 0, len(n.txMempool))
	for _, tx := range n.txMempool {
		cpy = append(cpy, tx)
	}
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

// QueryMempoolLength returns the current length of the mempool.
func (n *Node) QueryMempoolLength() int {
	n.mu.Lock()
	defer n.mu.Unlock()

	return len(n.txMempool)
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
			if tran.From == account || tran.To == account {
				out = append(out, block)
			}
		}
	}

	return out
}

// =============================================================================

// WriteNextBlock takes a block received from a peer, validates it and
// if that passes, writes the block to disk.
func (n *Node) WriteNextBlock(block Block) error {
	n.evHandler("node: WriteNextBlock: started : block[%s]", block.Hash())
	defer n.evHandler("node: WriteNextBlock: completed")

	hash, err := n.validateNextBlock(block)
	if err != nil {
		return err
	}

	blockFS := BlockFS{
		Hash:  hash,
		Block: block,
	}

	// Marshal the block for writing to disk.
	blockFSJson, err := json.Marshal(blockFS)
	if err != nil {
		return err
	}

	n.mu.Lock()
	{
		// Write the new block to the chain on disk.
		if _, err := n.file.Write(append(blockFSJson, '\n')); err != nil {
			return err
		}

		// Apply the transactions to the balance sheet and remove
		// from the mempool.
		for _, tx := range block.Transactions {
			applyTransactionToBalance(n.balances, tx)
			delete(n.txMempool, tx.ID)
		}

		// Save this as the latest block.
		n.latestBlock = block
	}
	n.mu.Unlock()

	return nil
}

// validateNextBlock takes a block and validates it to be included into
// the blockchain.
func (n *Node) validateNextBlock(block Block) (string, error) {
	hash := block.Hash()
	if !isHashSolved(hash) {
		return zeroHash, fmt.Errorf("%s invalid hash", hash)
	}

	latestBlock := n.CopyLatestBlock()
	nextNumber := latestBlock.Header.Number + 1

	if block.Header.Number != nextNumber {
		return zeroHash, fmt.Errorf("this block is not the next number, got %d, exp %d", block.Header.Number, nextNumber)
	}

	if block.Header.PrevBlock != latestBlock.Hash() {
		return zeroHash, fmt.Errorf("prev block doesn't match our latest, got %s, exp %s", block.Header.PrevBlock, latestBlock.Hash())
	}

	return hash, nil
}

// =============================================================================

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

// =============================================================================

// MineNewBlock writes the published transaction from the memory pool to disk.
func (n *Node) MineNewBlock(ctx context.Context) (Block, time.Duration, error) {
	var nb Block
	balances := make(map[string]uint)

	n.mu.Lock()
	{
		// Are there enough transactions in the pool.
		if len(n.txMempool) < 2 {
			return Block{}, 0, ErrNotEnoughTransactions
		}

		// Create a new block which owns it's own copy of the transactions.
		nb = NewBlock(n.latestBlock, n.txMempool)

		// Get a copy of the balance sheet.
		for act, bal := range n.balances {
			balances[act] = bal
		}
	}
	n.mu.Unlock()

	// Apply the transactions to that copy, setting status information.
	for i, tx := range nb.Transactions {
		if err := applyTransactionToBalance(balances, tx); err != nil {
			nb.Transactions[i].Status = TxStatusError
			nb.Transactions[i].StatusInfo = err.Error()
			continue
		}
		nb.Transactions[i].Status = TxStatusAccepted
	}

	// Attempt to create a new BlockFS by solving the POW puzzle.
	// This can be cancelled.
	blockFS, duration, err := performPOW(ctx, nb, n.evHandler)
	if err != nil {
		return Block{}, duration, err
	}

	// Just check one more time we were not cancelled.
	if ctx.Err() != nil {
		return Block{}, duration, ctx.Err()
	}

	// Marshal the block for writing to disk.
	blockFSJson, err := json.Marshal(blockFS)
	if err != nil {
		return Block{}, duration, err
	}

	// Update the state of the node.
	n.mu.Lock()
	{
		// Write the new block to the chain on disk.
		if _, err := n.file.Write(append(blockFSJson, '\n')); err != nil {
			return Block{}, duration, err
		}

		n.balances = balances
		n.latestBlock = blockFS.Block

		// Remove the transactions from this block.
		for _, tx := range nb.Transactions {
			delete(n.txMempool, tx.ID)
		}
	}
	n.mu.Unlock()

	return blockFS.Block, duration, nil
}
