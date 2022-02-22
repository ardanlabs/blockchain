// Package state is the core API for the blockchain and implements all the
// business rules and processing.
package state

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/ardanlabs/blockchain/foundation/blockchain/balance"
	"github.com/ardanlabs/blockchain/foundation/blockchain/genesis"
	"github.com/ardanlabs/blockchain/foundation/blockchain/mempool"
	"github.com/ardanlabs/blockchain/foundation/blockchain/peer"
	"github.com/ardanlabs/blockchain/foundation/blockchain/signature"
	"github.com/ardanlabs/blockchain/foundation/blockchain/storage"
)

/*
	-- Chrome Extension Wallet
	Get list of defined accounts (minus miner).
	Show the balance of each account.
	Concept of connecting and receiving events.

	-- Web Application
	Provide support to read a file of transactions to send.
	Need to verify enough money at the address before sending a transaction.
	See the different nodes, view activity.

	-- Blockchain
	Sync up mempool after blocks before starting.
	Create a block index file for query and clean up forks.
	Publishing events. (New Blocks)
	Unit need tests for all packages.
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
	MinerAddress string
	Host         string
	DBPath       string
	KnownPeers   *peer.PeerSet
	EvHandler    EventHandler
}

// State manages the blockchain database.
type State struct {
	minerAddress string
	host         string
	dbPath       string
	knownPeers   *peer.PeerSet
	evHandler    EventHandler

	genesis      genesis.Genesis
	storage      *storage.Storage
	mempool      *mempool.Mempool
	latestBlock  storage.Block
	balanceSheet *balance.Sheet
	mu           sync.Mutex

	worker *worker
}

// New constructs a new blockchain for data management.
func New(cfg Config) (*State, error) {

	// Load the genesis file to get starting balances for
	// founders of the block chain.
	genesis, err := genesis.Load()
	if err != nil {
		return nil, err
	}

	// Access the storage for the blockchain.
	strg, err := storage.New(cfg.DBPath)
	if err != nil {
		return nil, err
	}

	// Load all existing blocks from storage into memory for processing. This
	// won't work in a system like Ethereum.
	blocks, err := strg.ReadAllBlocks()
	if err != nil {
		return nil, err
	}

	// Keep the latest block from the blockchain.
	var latestBlock storage.Block
	if len(blocks) > 0 {
		latestBlock = blocks[len(blocks)-1]
	}

	// Create a new balance sheet from the genesis balances.
	sheet := balance.NewSheet(genesis.MiningReward, genesis.Balances)

	// Process the blocks and transactions against the balance sheet.
	for _, block := range blocks {
		for _, tx := range block.Transactions {

			// Apply the balance changes based for this transaction.
			sheet.ApplyTransaction(block.Header.MinerAddress, tx)
		}

		// Apply the mining reward for this block.
		sheet.ApplyMiningReward(block.Header.MinerAddress)
	}

	// Build a safe event handler function for use.
	ev := func(v string, args ...interface{}) {
		if cfg.EvHandler != nil {
			cfg.EvHandler(v, args...)
		}
	}

	// Create the State to provide support for managing the blockchain.
	state := State{
		minerAddress: cfg.MinerAddress,
		host:         cfg.Host,
		dbPath:       cfg.DBPath,
		knownPeers:   cfg.KnownPeers,
		evHandler:    ev,

		genesis:      genesis,
		storage:      strg,
		mempool:      mempool.New(),
		latestBlock:  latestBlock,
		balanceSheet: sheet,
	}

	// Run the worker which will assign itself to this state.
	runWorker(&state, cfg.EvHandler)

	return &state, nil
}

// Shutdown cleanly brings the node down.
func (s *State) Shutdown() error {

	// Make sure the database file is properly closed.
	defer func() {
		s.storage.Close()
	}()

	// Stop all blockchain writing activity.
	s.worker.shutdown()

	return nil
}

// =============================================================================

// SubmitWalletTransaction accepts a transaction from a wallet for inclusion.
func (s *State) SubmitWalletTransaction(signedTx storage.SignedTx) error {
	tx := storage.BlockTx{
		SignedTx: signedTx,
		Gas:      s.genesis.GasPrice,
	}

	if err := tx.VerifySignature(); err != nil {
		return err
	}

	n := s.mempool.Add(tx)
	s.worker.signalShareTransactions(tx)

	if n >= s.genesis.TransPerBlock {
		s.worker.signalStartMining()
	}

	return nil
}

// SubmitNodeTransaction accepts a transaction from a node for inclusion.
func (s *State) SubmitNodeTransaction(tx storage.BlockTx) error {
	if err := tx.VerifySignature(); err != nil {
		return err
	}

	n := s.mempool.Add(tx)
	if n >= s.genesis.TransPerBlock {
		s.worker.signalStartMining()
	}

	return nil
}

// =============================================================================

// WriteNextBlock takes a block received from a peer, validates it and
// if that passes, writes the block to disk.
func (s *State) WriteNextBlock(block storage.Block) error {
	s.evHandler("state: WriteNextBlock: started : block[%s]", block.Hash())
	defer s.evHandler("state: WriteNextBlock: completed")

	// If the runMiningOperation function is being executed it needs to stop
	// immediately. The G executing runMiningOperation will not return from the
	// function until done is called. That allows this function to complete
	// its state changes before a new mining operation takes place.
	done := s.worker.signalCancelMining()
	defer func() {
		s.evHandler("state: WriteNextBlock: signal runMiningOperation to terminate")
		done()
	}()

	hash, err := s.validateNextBlock(block)
	if err != nil {
		return err
	}

	blockFS := storage.BlockFS{
		Hash:  hash,
		Block: block,
	}

	// I want to make sure all these state changes are done atomically.
	s.mu.Lock()
	defer s.mu.Unlock()
	{
		s.evHandler("state: WriteNextBlock: write to disk")

		// Write the new block to the chain on disk.
		if err := s.storage.Write(blockFS); err != nil {
			return err
		}

		s.evHandler("state: WriteNextBlock: apply transactions to balance")

		// Process the transactions against the balance sheet.
		for _, tx := range block.Transactions {

			// Apply the balance changes based for this transaction.
			s.balanceSheet.ApplyTransaction(block.Header.MinerAddress, tx)

			s.evHandler("state: WriteNextBlock: remove from mempool: tx[%s]", tx.Hash())

			// Remove the transaction from the mempool if it exists.
			s.mempool.Delete(tx)
		}

		s.evHandler("state: WriteNextBlock: apply mining reward")

		// Apply the mining reward for this block.
		s.balanceSheet.ApplyMiningReward(block.Header.MinerAddress)

		// Save this as the latest block.
		s.latestBlock = block
	}

	return nil
}

// validateNextBlock takes a block and validates it to be included into
// the blockchain.
func (s *State) validateNextBlock(block storage.Block) (string, error) {
	s.evHandler("state: WriteNextBlock: validate: hash solved")

	hash := block.Hash()
	if !isHashSolved(s.genesis.Difficulty, hash) {
		return signature.ZeroHash, fmt.Errorf("%s invalid hash", hash)
	}

	latestBlock := s.RetrieveLatestBlock()
	nextNumber := latestBlock.Header.Number + 1

	s.evHandler("state: WriteNextBlock: validate: chain not forked")

	// The node who sent this block has a chain that is two or more blocks ahead
	// of ours. This means there has been a fork and we are on the wrong side.
	if block.Header.Number >= (nextNumber + 2) {
		return signature.ZeroHash, ErrChainForked
	}

	s.evHandler("state: WriteNextBlock: validate: block number")

	if block.Header.Number != nextNumber {
		return signature.ZeroHash, fmt.Errorf("this block is not the next number, got %d, exp %d", block.Header.Number, nextNumber)
	}

	s.evHandler("state: WriteNextBlock: validate: parent hash")

	if block.Header.ParentHash != latestBlock.Hash() {
		return signature.ZeroHash, fmt.Errorf("prev block doesn't match our latest, got %s, exp %s", block.Header.ParentHash, latestBlock.Hash())
	}

	s.evHandler("state: WriteNextBlock: validate: transaction signatures")

	for _, tx := range block.Transactions {
		if err := tx.VerifySignature(); err != nil {
			return signature.ZeroHash, fmt.Errorf("transaction has invalid signature, %w, tx[%v]", err, tx)
		}
	}

	return hash, nil
}

// =============================================================================

// Truncate resets the chain both on disk and in memory. This is used to
// correct an identified fork.
func (s *State) Truncate() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Reset the state of the database.
	s.mempool.Truncate()
	s.balanceSheet.Reset(s.genesis.Balances)
	s.latestBlock = storage.Block{}
	s.storage.Reset()

	return nil
}

// =============================================================================

// RetrieveGenesis returns a copy of the genesis information.
func (s *State) RetrieveGenesis() genesis.Genesis {
	return s.genesis
}

// RetrieveLatestBlock returns a copy the current latest block.
func (s *State) RetrieveLatestBlock() storage.Block {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.latestBlock
}

// RetrieveMempool returns a copy of the mempool.
func (s *State) RetrieveMempool() []storage.BlockTx {
	return s.mempool.Copy()
}

// RetrieveBalanceSheetValues returns a copy of the balance sheet values.
func (s *State) RetrieveBalanceSheetValues() map[string]uint {
	return s.balanceSheet.Values()
}

// RetrieveKnownPeers retrieves a copy of the known peer list.
func (s *State) RetrieveKnownPeers() []peer.Peer {
	return s.knownPeers.Copy(s.host)
}

// =============================================================================

// QueryLastest represents to query the latest block in the chain.
const QueryLastest = ^uint64(0) >> 1

// QueryBalances returns a copy of the set of balances by address.
func (s *State) QueryBalances(address string) map[string]uint {
	balanceSheet := s.balanceSheet.Clone()

	cpy := balanceSheet.Values()
	for addr := range cpy {
		if address != addr {
			balanceSheet.Remove(addr)
		}
	}

	return cpy
}

// QueryMempoolLength returns the current length of the mempool.
func (s *State) QueryMempoolLength() int {
	return s.mempool.Count()
}

// QueryBlocksByNumber returns the set of blocks based on block numbers. This
// function reads the blockchain from disk first.
func (s *State) QueryBlocksByNumber(from uint64, to uint64) []storage.Block {
	blocks, err := s.storage.ReadAllBlocks()
	if err != nil {
		return nil
	}

	if from == QueryLastest {
		from = blocks[len(blocks)-1].Header.Number
		to = from
	}

	var out []storage.Block
	for _, block := range blocks {
		if block.Header.Number >= from && block.Header.Number <= to {
			out = append(out, block)
		}
	}

	return out
}

// QueryBlocksByAddress returns the set of blocks by address. If the address
// is empty, all blocks are returned. This function reads the blockchain
// from disk first.
func (s *State) QueryBlocksByAddress(address string) []storage.Block {
	blocks, err := s.storage.ReadAllBlocks()
	if err != nil {
		return nil
	}

	var out []storage.Block
blocks:
	for _, block := range blocks {
		for _, tx := range block.Transactions {
			from, err := tx.FromAddress()
			if err != nil {
				continue
			}
			if address == "" || from == address || tx.To == address {
				out = append(out, block)
				continue blocks
			}
		}
	}

	return out
}

// =============================================================================

// addPeerNode adds an address to the list of peers.
func (s *State) addPeerNode(peer peer.Peer) error {

	// Don't add this node to the known peer list.
	if peer.Match(s.host) {
		return errors.New("already exists")
	}

	s.knownPeers.Add(peer)
	return nil
}

// =============================================================================

// MineNewBlock writes the published transaction from the memory pool to disk.
func (s *State) MineNewBlock(ctx context.Context) (storage.Block, time.Duration, error) {
	s.evHandler("worker: runMiningOperation: MINING: check mempool count")

	// Are there enough transactions in the pool.
	if s.mempool.Count() < s.genesis.TransPerBlock {
		return storage.Block{}, 0, ErrNotEnoughTransactions
	}

	s.evHandler("worker: runMiningOperation: MINING: create new block")

	// Create a new block which owns it's own copy of the transactions.
	trans := s.mempool.CopyBestByTip(s.genesis.TransPerBlock)
	nb := storage.NewBlock(s.minerAddress, s.genesis.Difficulty, s.genesis.TransPerBlock, s.RetrieveLatestBlock(), trans)

	s.evHandler("worker: runMiningOperation: MINING: copy balance sheet and update")

	// Process the transactions against the balance sheet.
	balanceSheet := s.balanceSheet.Clone()
	for _, tx := range nb.Transactions {

		// Apply the balance changes based on this transaction.
		if err := balanceSheet.ApplyTransaction(s.minerAddress, tx); err != nil {
			s.evHandler("worker: runMiningOperation: MINING: WARNING : %s", err)
			continue
		}

		// Update the total gas and tip fees.
		nb.Header.TotalGas += tx.Gas
		nb.Header.TotalTip += tx.Tip
	}

	// Apply the mining reward for this block.
	balanceSheet.ApplyMiningReward(s.minerAddress)

	s.evHandler("worker: runMiningOperation: MINING: perform POW")

	// Attempt to create a new BlockFS by solving the POW puzzle.
	// This can be cancelled.
	blockFS, duration, err := performPOW(ctx, s.genesis.Difficulty, nb, s.evHandler)
	if err != nil {
		return storage.Block{}, duration, err
	}

	// Just check one more time we were not cancelled.
	if ctx.Err() != nil {
		return storage.Block{}, duration, ctx.Err()
	}

	// I want to make sure all these state changes are done atomically.
	s.mu.Lock()
	defer s.mu.Unlock()
	{
		s.evHandler("worker: runMiningOperation: MINING: write to disk")

		// Write the new block to the chain on disk.
		if err := s.storage.Write(blockFS); err != nil {
			return storage.Block{}, duration, err
		}

		s.evHandler("worker: runMiningOperation: MINING: apply new balance sheet")

		s.balanceSheet.Replace(balanceSheet)
		s.latestBlock = blockFS.Block

		// Remove the transactions from this block.
		for _, tx := range nb.Transactions {
			s.evHandler("worker: runMiningOperation: MINING: remove from mempool: tx[%s]", tx.Hash())
			s.mempool.Delete(tx)
		}
	}

	return blockFS.Block, duration, nil
}
