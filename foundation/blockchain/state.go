package blockchain

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
	Choose the best transactions based on fees.
	Need a wallet to sign transactions properly.
	Maybe adjust difficulty based on time to mine. Currently hardcoded to 6 zeros.
	Add fees to transactions.
	Add the token supply and global blockchain settings in the genesis file.
*/

// =============================================================================

// ErrNotEnoughTransactions is returned when a block is requested to be created
// and there are not enough transactions.
var ErrNotEnoughTransactions = errors.New("not enough transactions in mempool")

// ErrChainForked is returned from validateNextBlock if another node's chain
// is two or more blocks ahead of ours.
var ErrChainForked = errors.New("blockchain forked, start resync")

// =============================================================================

// EventHandler defines a function that is called when events
// occur in the processing of persisting blocks.
type EventHandler func(v string, args ...interface{})

// Config represents the configuration required to start
// the blockchain node.
type Config struct {
	MinerAccount  string
	Host          string
	DBPath        string
	Reward        uint // What a miner gets for mining a block.
	Difficulty    int  // How many leading eros a block hash must have.
	TransPerBlock int  // How many transactions need to be in a block.
	KnownPeers    PeerSet
	EvHandler     EventHandler
}

// State manages the blockchain database.
type State struct {
	minerAccount  string
	host          string
	dbPath        string
	reward        uint
	difficulty    int
	transPerBlock int
	knownPeers    PeerSet
	evHandler     EventHandler

	genesis      Genesis
	txMempool    map[ID]Tx
	latestBlock  Block
	balanceSheet BalanceSheet
	dbFile       *os.File
	mu           sync.Mutex

	bcWorker *bcWorker
}

// New constructs a new blockchain for data management.
func New(cfg Config) (*State, error) {

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
		applyMiningRewardToBalance(balanceSheet, block.Header.Beneficiary, cfg.Reward)
	}

	// Open the blockchain database file for processing.
	dbFile, err := os.OpenFile(cfg.DBPath, os.O_APPEND|os.O_RDWR, 0600)
	if err != nil {
		return nil, err
	}

	// Create the State to provide support for managing the blockchain.
	state := State{
		minerAccount:  cfg.MinerAccount,
		host:          cfg.Host,
		dbPath:        cfg.DBPath,
		reward:        cfg.Reward,
		difficulty:    cfg.Difficulty,
		transPerBlock: cfg.TransPerBlock,
		knownPeers:    cfg.KnownPeers,
		evHandler:     ev,

		genesis:      genesis,
		txMempool:    make(map[ID]Tx),
		latestBlock:  latestBlock,
		balanceSheet: balanceSheet,
		dbFile:       dbFile,
	}

	ev("node: Started: blocks[%d]", latestBlock.Header.Number)

	// Run the blockchain workers.
	state.bcWorker = runBCWorker(&state, cfg.EvHandler)

	return &state, nil
}

// Shutdown cleanly brings the node down.
func (s *State) Shutdown() error {
	s.mu.Lock()
	defer func() {
		s.dbFile.Close()
		s.mu.Unlock()
	}()

	// Stop all blockchain writing activity.
	s.bcWorker.shutdown()

	return nil
}

// SignalMining sends a signal to the mining G to start.
func (s *State) SignalMining() {
	s.bcWorker.signalStartMining()
}

// SignalCancelMining sends a signal to the mining G to stop.
func (s *State) SignalCancelMining() {
	s.bcWorker.signalCancelMining()
}

// =============================================================================

// AddTransactions appends a new transactions to the mempool.
func (s *State) AddTransactions(txs []Tx, share bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.evHandler("node: AddTransactions: started : txrs[%d]", len(txs))
	defer s.evHandler("node: AddTransactions: completed")

	s.evHandler("node: AddTransactions: before: mempool[%d]", len(s.txMempool))
	for _, tx := range txs {
		if _, exists := s.txMempool[tx.ID]; !exists {
			s.txMempool[tx.ID] = tx
		}
	}
	s.evHandler("node: AddTransactions: after: mempool[%d]", len(s.txMempool))

	if share {
		s.evHandler("node: AddTransactions: signal tx sharing")
		s.bcWorker.signalShareTransactions(txs)
	}

	if len(s.txMempool) >= s.transPerBlock {
		s.evHandler("node: AddTransactions: signal mining")
		s.bcWorker.signalStartMining()
	}
}

// CopyGenesis returns a copy of the genesis information.
func (s *State) CopyGenesis() Genesis {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.genesis
}

// CopyLatestBlock returns the current hash of the latest block.
func (s *State) CopyLatestBlock() Block {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.latestBlock
}

// CopyMempool returns a copy of the mempool.
func (s *State) CopyMempool() []Tx {
	s.mu.Lock()
	defer s.mu.Unlock()

	cpy := make([]Tx, 0, len(s.txMempool))
	for _, tx := range s.txMempool {
		cpy = append(cpy, tx)
	}
	return cpy
}

// CopyBalanceSheet returns a copy of the balance sheet.
func (s *State) CopyBalanceSheet() BalanceSheet {
	s.mu.Lock()
	defer s.mu.Unlock()

	return copyBalanceSheet(s.balanceSheet)
}

// CopyKnownPeerSet retrieves information about the peer for updating
// the known peer list and their current block number.
func (s *State) CopyKnownPeers() []Peer {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Can't include ourselves in this list.
	peers := make([]Peer, 0, len(s.knownPeers)-1)
	for peer := range s.knownPeers {
		if !peer.match(s.host) {
			peers = append(peers, peer)
		}
	}

	return peers
}

// =============================================================================

// QueryLastest represents to query the latest block in the chain.
const QueryLastest = ^uint64(0) >> 1

// QueryBalances returns a copy of the set of balances by account.
func (s *State) QueryBalances(account string) BalanceSheet {
	s.mu.Lock()
	defer s.mu.Unlock()

	balanceSheet := newBalanceSheet()
	for acct, value := range s.balanceSheet {
		if account == acct {
			balanceSheet.replace(acct, value)
		}
	}

	return balanceSheet
}

// QueryMempoolLength returns the current length of the mempool.
func (s *State) QueryMempoolLength() int {
	s.mu.Lock()
	defer s.mu.Unlock()

	return len(s.txMempool)
}

// QueryBlocksByNumber returns the set of blocks based on block numbers. This
// function reads the blockchain from disk first.
func (s *State) QueryBlocksByNumber(from uint64, to uint64) []Block {
	blocks, err := loadBlocksFromDisk(s.dbPath)
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
func (s *State) QueryBlocksByAccount(account string) []Block {
	blocks, err := loadBlocksFromDisk(s.dbPath)
	if err != nil {
		return nil
	}

	var out []Block
blocks:
	for _, block := range blocks {
		for _, tran := range block.Transactions {
			if account == "" || tran.From == account || tran.To == account {
				out = append(out, block)
				continue blocks
			}
		}
	}

	return out
}

// =============================================================================

// WriteNextBlock takes a block received from a peer, validates it and
// if that passes, writes the block to disk.
func (s *State) WriteNextBlock(block Block) error {
	s.evHandler("node: WriteNextBlock: started : block[%s]", block.Hash())
	defer s.evHandler("node: WriteNextBlock: completed")

	hash, err := s.validateNextBlock(block)
	if err != nil {

		// We need to attempt to correct the fork in our chain. We will wipe
		// out our current chain on disk and reset from our peers.
		if errors.Is(err, ErrChainForked) {
			s.clearChainAndReset()
		}

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
		s.mu.Lock()
		defer s.mu.Unlock()

		// Write the new block to the chain on disk.
		if _, err := s.dbFile.Write(append(blockFSJson, '\n')); err != nil {
			return err
		}

		// Apply the transactions to the balance sheet and remove
		// from the mempool.
		for _, tx := range block.Transactions {
			applyTransactionToBalance(s.balanceSheet, tx)
			delete(s.txMempool, tx.ID)
		}

		// Add the miner reward for the beneficiary to the balance sheet.
		applyMiningRewardToBalance(s.balanceSheet, block.Header.Beneficiary, s.reward)

		// Save this as the latest block.
		s.latestBlock = block

		return nil
	}(); err != nil {
		return err
	}

	return nil
}

// validateNextBlock takes a block and validates it to be included into
// the blockchain.
func (s *State) validateNextBlock(block Block) (string, error) {
	hash := block.Hash()
	if !isHashSolved(s.difficulty, hash) {
		return zeroHash, fmt.Errorf("%s invalid hash", hash)
	}

	latestBlock := s.CopyLatestBlock()
	nextNumber := latestBlock.Header.Number + 1

	// The node who sent this block has a chain that is two or more blocks ahead
	// of ours. This means there has been a fork and we are on the wrong side.
	if block.Header.Number >= (nextNumber + 2) {
		return zeroHash, ErrChainForked
	}

	if block.Header.Number != nextNumber {
		return zeroHash, fmt.Errorf("this block is not the next number, got %d, exp %d", block.Header.Number, nextNumber)
	}

	if block.Header.ParentHash != latestBlock.Hash() {
		return zeroHash, fmt.Errorf("prev block doesn't match our latest, got %s, exp %s", block.Header.ParentHash, latestBlock.Hash())
	}

	return hash, nil
}

// clearChainAndReset clears the state of the blockchain to start over.
// This is a simplistic way to approach this for now.
func (s *State) clearChainAndReset() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Stop the peer ticker and then reset it.
	// TODO: It might be important to run if this is already running.
	s.bcWorker.ticker.Stop()
	defer s.bcWorker.ticker.Reset(peerUpdateInterval)

	// Close the remove the current blockchain database file.
	s.dbFile.Close()
	if err := os.Remove(s.dbPath); err != nil {
		return err
	}

	// Open a new blockchain database file for processing.
	dbFile, err := os.OpenFile(s.dbPath, os.O_APPEND|os.O_RDWR, 0600)
	if err != nil {
		return err
	}

	// Reload the genesis file to get starting balances for
	// founders of the block chain.
	genesis, err := loadGenesis()
	if err != nil {
		return err
	}

	// Apply the genesis balances to the balance sheet.
	balanceSheet := copyBalanceSheet(genesis.Balances)

	// Reset the state of the database.
	s.genesis = genesis
	s.txMempool = make(map[ID]Tx)
	s.latestBlock = Block{}
	s.balanceSheet = balanceSheet
	s.dbFile = dbFile

	// Attempt to update the blockchain on disk from the peer's.
	// TODO: It might be important to run if this is already running.
	s.bcWorker.runPeerOperation()

	return nil
}

// =============================================================================

// addPeerNode adds an address to the list of peers.
func (s *State) addPeerNode(peer Peer) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Don't add this node to the known peer list.
	if peer.match(s.host) {
		return errors.New("already exists")
	}

	s.knownPeers.Add(peer)
	return nil
}

// =============================================================================

// MineNewBlock writes the published transaction from the memory pool to disk.
func (s *State) MineNewBlock(ctx context.Context) (Block, time.Duration, error) {
	var nb Block
	var balanceSheet BalanceSheet

	// Execute this code inside a lock.
	if err := func() error {
		s.mu.Lock()
		defer s.mu.Unlock()

		// Are there enough transactions in the pool.
		if len(s.txMempool) < s.transPerBlock {
			s.mu.Unlock()
			return ErrNotEnoughTransactions
		}

		// Create a new block which owns it's own copy of the transactions.
		nb = NewBlock(s.minerAccount, s.difficulty, s.latestBlock, s.txMempool)

		// Get a copy of the balance sheet.
		balanceSheet = copyBalanceSheet(s.balanceSheet)

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
	applyMiningRewardToBalance(balanceSheet, s.minerAccount, s.reward)

	// Attempt to create a new BlockFS by solving the POW puzzle.
	// This can be cancelled.
	blockFS, duration, err := performPOW(ctx, s.difficulty, nb, s.evHandler)
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
		s.mu.Lock()
		defer s.mu.Unlock()

		// Write the new block to the chain on disk.
		if _, err := s.dbFile.Write(append(blockFSJson, '\n')); err != nil {
			s.mu.Unlock()
			return err
		}

		s.balanceSheet = balanceSheet
		s.latestBlock = blockFS.Block

		// Remove the transactions from this block.
		for _, tx := range nb.Transactions {
			delete(s.txMempool, tx.ID)
		}

		return nil
	}(); err != nil {
		return Block{}, duration, err
	}

	return blockFS.Block, duration, nil
}
