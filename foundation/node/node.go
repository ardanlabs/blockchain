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
	Set a maximum number of transactions per block.
	Choose the best transactions based on fees.
	Need a way to identify my chain is no longer the valid chain, re-sync.
	Need a way to validate a new block against the entire known blockchain.
	Need a wallet to sign transactions properly.
	Maybe adjust difficulty based on time to mine. Currently hardcoded to 6 zeros.
	Add fees to transactions.
	Add the token supply and global blockchain settings in the genesis file.
*/

// =============================================================================

// ErrNotEnoughTransactions is returned when a block is requested to be created
// and there are not enough transactions.
var ErrNotEnoughTransactions = errors.New("not enough transactions in mempool")

// =============================================================================

// EventHandler defines a function that is called when events
// occur in the processing of persisting blocks.
type EventHandler func(v string, args ...interface{})

// Config represents the configuration required to start
// the blockchain node.
type Config struct {
	Account    string
	Host       string
	DBPath     string
	KnownPeers PeerSet
	Reward     uint
	Difficulty int
	EvHandler  EventHandler
}

// Node manages the blockchain database.
type Node struct {
	account    string
	host       string
	dbPath     string
	knownPeers PeerSet
	reward     uint
	difficulty int
	evHandler  EventHandler

	genesis      Genesis
	txMempool    map[ID]Tx
	latestBlock  Block
	balanceSheet BalanceSheet
	file         *os.File
	mu           sync.Mutex

	bcWorker *bcWorker
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

	// Apply the genesis balances to the balance sheet.
	balanceSheet := copyBalanceSheet(genesis.Balances)

	// Update the balance sheet by processing all the transactions in
	// the set of blocks.
	for _, block := range blocks {
		if err := applyTransactionsToBalances(balanceSheet, block.Transactions); err != nil {
			return nil, err
		}

		// Add the miner reward to the balance sheet.
		applyMiningRewardToBalance(balanceSheet, block.Header.MinerAccount, cfg.Reward)
	}

	// Open the blockchain database file for processing.
	file, err := os.OpenFile(cfg.DBPath, os.O_APPEND|os.O_RDWR, 0600)
	if err != nil {
		return nil, err
	}

	// Create the node to provide support for managing the blockchain.
	node := Node{
		account:    cfg.Account,
		host:       cfg.Host,
		dbPath:     cfg.DBPath,
		knownPeers: cfg.KnownPeers,
		reward:     cfg.Reward,
		difficulty: cfg.Difficulty,
		evHandler:  ev,

		genesis:      genesis,
		txMempool:    make(map[ID]Tx),
		latestBlock:  latestBlock,
		balanceSheet: balanceSheet,
		file:         file,
	}

	ev("node: Started: blocks[%d]", latestBlock.Header.Number)

	// Run the blockchain workers.
	node.bcWorker = runBCWorker(&node, cfg.EvHandler)

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

// CopyBalanceSheet returns a copy of the balance sheet.
func (n *Node) CopyBalanceSheet() BalanceSheet {
	n.mu.Lock()
	defer n.mu.Unlock()

	return copyBalanceSheet(n.balanceSheet)
}

// CopyKnownPeerSet retrieves information about the peer for updating
// the known peer list and their current block number.
func (n *Node) CopyKnownPeers() []Peer {
	n.mu.Lock()
	defer n.mu.Unlock()

	// Can't include ourselves in this list.
	peers := make([]Peer, 0, len(n.knownPeers)-1)
	for peer := range n.knownPeers {
		if !peer.match(n.host) {
			peers = append(peers, peer)
		}
	}

	return peers
}

// =============================================================================

// QueryLastest represents to query the latest block in the chain.
const QueryLastest = ^uint64(0) >> 1

// QueryBalances returns a copy of the set of balances by account.
func (n *Node) QueryBalances(account string) BalanceSheet {
	n.mu.Lock()
	defer n.mu.Unlock()

	balanceSheet := newBalanceSheet()
	for acct, value := range n.balanceSheet {
		if account == acct {
			balanceSheet.replace(acct, value)
		}
	}

	return balanceSheet
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

	var out []Block
	for _, block := range blocks {
		for _, tran := range block.Transactions {
			if tran.FromAccount == account || tran.ToAccount == account {
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

	// Execute this code inside a lock.
	if err := func() error {
		n.mu.Lock()
		defer n.mu.Unlock()

		// Write the new block to the chain on disk.
		if _, err := n.file.Write(append(blockFSJson, '\n')); err != nil {
			return err
		}

		// Apply the transactions to the balance sheet and remove
		// from the mempool.
		for _, tx := range block.Transactions {
			applyTransactionToBalance(n.balanceSheet, tx)
			delete(n.txMempool, tx.ID)
		}

		// Add the miner reward to the balance sheet.
		applyMiningRewardToBalance(n.balanceSheet, block.Header.MinerAccount, n.reward)

		// Save this as the latest block.
		n.latestBlock = block

		return nil
	}(); err != nil {
		return err
	}

	return nil
}

// validateNextBlock takes a block and validates it to be included into
// the blockchain.
func (n *Node) validateNextBlock(block Block) (Hash, error) {
	hash := block.Hash()
	if !isHashSolved(n.difficulty, hash) {
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
func (n *Node) addPeerNode(peer Peer) error {
	n.mu.Lock()
	defer n.mu.Unlock()

	// Don't add this node to the known peer list.
	if peer.match(n.host) {
		return errors.New("already exists")
	}

	n.knownPeers.Add(peer)
	return nil
}

// =============================================================================

// MineNewBlock writes the published transaction from the memory pool to disk.
func (n *Node) MineNewBlock(ctx context.Context) (Block, time.Duration, error) {
	var nb Block
	var balanceSheet BalanceSheet

	// Execute this code inside a lock.
	if err := func() error {
		n.mu.Lock()
		defer n.mu.Unlock()

		// Are there enough transactions in the pool.
		if len(n.txMempool) < 2 {
			n.mu.Unlock()
			return ErrNotEnoughTransactions
		}

		// Create a new block which owns it's own copy of the transactions.
		nb = NewBlock(n.account, n.latestBlock, n.txMempool)

		// Get a copy of the balance sheet.
		balanceSheet = copyBalanceSheet(n.balanceSheet)

		return nil
	}(); err != nil {
		return Block{}, 0, ErrNotEnoughTransactions
	}

	// Apply the transactions to the copy of the balance sheet and
	// set the status information.
	for i, tx := range nb.Transactions {
		if err := applyTransactionToBalance(balanceSheet, tx); err != nil {
			nb.Transactions[i].Status = TxStatusError
			nb.Transactions[i].StatusInfo = err.Error()
			continue
		}
		nb.Transactions[i].Status = TxStatusAccepted
	}

	// Add the miner reward to the balance sheet.
	applyMiningRewardToBalance(balanceSheet, n.account, n.reward)

	// Attempt to create a new BlockFS by solving the POW puzzle.
	// This can be cancelled.
	blockFS, duration, err := performPOW(ctx, n.difficulty, nb, n.evHandler)
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

	// Execute this code inside a lock.
	if err := func() error {
		n.mu.Lock()
		defer n.mu.Unlock()

		// Write the new block to the chain on disk.
		if _, err := n.file.Write(append(blockFSJson, '\n')); err != nil {
			n.mu.Unlock()
			return err
		}

		n.balanceSheet = balanceSheet
		n.latestBlock = blockFS.Block

		// Remove the transactions from this block.
		for _, tx := range nb.Transactions {
			delete(n.txMempool, tx.ID)
		}

		return nil
	}(); err != nil {
		return Block{}, duration, err
	}

	return blockFS.Block, duration, nil
}
