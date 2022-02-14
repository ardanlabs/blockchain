package blockchain

import (
	"context"
	"crypto/ecdsa"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sync"
	"time"
)

/*
	-- Wallet
	Provide name resolution for name => address
	Provide support to read a file of transactions to send.
	Concept of connecting and receiving events.
	Need to verify enough money at the address before sending a transaction.

	-- Blockchain
	Sign blockTx for delivery to another node. Use chain id as salt.
	Add the miner1 and miner2 accounts to the zblock/wallet folder.
	Add a name server for known account. Used for displaying information.
	Create a block index file for query and clean up forks.
	Publishing events. (New Blocks)

	-- Testing
	Need tests.
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
	PrivateKey *ecdsa.PrivateKey
	MinerName  string
	Host       string
	DBPath     string
	KnownPeers PeerSet
	EvHandler  EventHandler
}

// State manages the blockchain database.
type State struct {
	privateKey *ecdsa.PrivateKey
	minerName  string
	host       string
	dbPath     string
	knownPeers PeerSet
	evHandler  EventHandler

	genesis      Genesis
	txMempool    txMempool
	latestBlock  SignedBlock
	balanceSheet BalanceSheet
	dbFile       *os.File
	mu           sync.Mutex

	powWorker *powWorker
}

// New constructs a new blockchain for data management.
func New(cfg Config) (*State, error) {

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
	var latestBlock SignedBlock
	if len(blocks) > 0 {
		latestBlock = blocks[len(blocks)-1]
	}

	// Apply the genesis balances to the balance sheet.
	balanceSheet := copyBalanceSheet(genesis.Balances)

	// Process the blocks and transactions against the balance sheet.
	for _, block := range blocks {
		for _, tx := range block.Transactions {

			// Apply the balance changes based on this transaction.
			applyTransactionToBalance(balanceSheet, tx)

			// Apply the miner tip and gas fee for this transaction.
			applyMiningFeeToBalance(balanceSheet, block.Header.Beneficiary, tx)
		}

		// Apply the miner reward for this block.
		applyMiningRewardToBalance(balanceSheet, block.Header.Beneficiary, genesis.MiningReward)
	}

	// Open the blockchain database file for processing.
	dbFile, err := os.OpenFile(cfg.DBPath, os.O_APPEND|os.O_RDWR, 0600)
	if err != nil {
		return nil, err
	}

	// Build a safe event handler function for use.
	ev := func(v string, args ...interface{}) {
		if cfg.EvHandler != nil {
			cfg.EvHandler(v, args...)
		}
	}

	// Create the State to provide support for managing the blockchain.
	state := State{
		privateKey: cfg.PrivateKey,
		minerName:  cfg.MinerName,
		host:       cfg.Host,
		dbPath:     cfg.DBPath,
		knownPeers: cfg.KnownPeers,
		evHandler:  ev,

		genesis:      genesis,
		txMempool:    newTxMempool(),
		latestBlock:  latestBlock,
		balanceSheet: balanceSheet,
		dbFile:       dbFile,
	}

	// Run the POW worker which will assign itself to
	// this state.
	runPOWWorker(&state, cfg.EvHandler)

	return &state, nil
}

// Shutdown cleanly brings the node down.
func (s *State) Shutdown() error {

	// Make sure the database file is properly closed.
	defer s.dbFile.Close()

	// Stop all blockchain writing activity.
	s.powWorker.shutdown()

	return nil
}

func (s *State) MinerName() string {
	return s.minerName
}

// =============================================================================

// SubmitWalletTransaction accepts a transaction from a wallet for inclusion.
func (s *State) SubmitWalletTransaction(signedTx SignedTx) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	tx := BlockTx{
		SignedTx: signedTx,
		Gas:      s.genesis.GasPrice,
	}

	if err := tx.VerifySignature(); err != nil {
		return err
	}

	n := s.txMempool.add(tx)
	s.powWorker.signalShareTransactions(tx)

	if n >= s.genesis.TransPerBlock {
		s.powWorker.signalStartMining()
	}

	return nil
}

// SubmitNodeTransaction accepts a transaction from a node for inclusion.
func (s *State) SubmitNodeTransaction(tx BlockTx) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := tx.VerifySignature(); err != nil {
		return err
	}

	n := s.txMempool.add(tx)
	if n >= s.genesis.TransPerBlock {
		s.powWorker.signalStartMining()
	}

	return nil
}

// =============================================================================

// WriteNextBlock takes a block received from a peer, validates it and
// if that passes, writes the block to disk.
func (s *State) WriteNextBlock(block SignedBlock) error {
	s.evHandler("state: WriteNextBlock: started : block[%s]", block.Hash())
	defer s.evHandler("state: WriteNextBlock: completed")

	// If the runMiningOperation function is being executed it needs to stop
	// immediately. The G executing runMiningOperation will not return from the
	// function until done is called. That allows this function to complete
	// its state changes before a new mining operation takes place.
	done := s.powWorker.signalCancelMining()
	defer func() {
		s.evHandler("state: WriteNextBlock: signal runMiningOperation to terminate")
		done()
	}()

	hash, err := s.validateNextBlock(block)
	if err != nil {
		return err
	}

	blockFS := blockFS{
		Hash:        hash,
		SignedBlock: block,
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

		s.evHandler("state: WriteNextBlock: write to disk")

		// Write the new block to the chain on disk.
		if _, err := s.dbFile.Write(append(blockFSJson, '\n')); err != nil {
			return err
		}

		s.evHandler("state: WriteNextBlock: apply transactions to balance")

		// Process the transactions against the balance sheet.
		for _, tx := range block.Transactions {

			// Apply the balance changes based on this transaction.
			applyTransactionToBalance(s.balanceSheet, tx)

			// Apply the miner tip and gas fee for this transaction.
			applyMiningFeeToBalance(s.balanceSheet, block.Header.Beneficiary, tx)

			s.evHandler("state: WriteNextBlock: remove from mempool: tx[%s]", tx.Hash())

			// Remove the transaction from the mempool if it exists.
			s.txMempool.delete(tx)
		}

		s.evHandler("state: WriteNextBlock: apply mining reward")

		// Apply the miner reward for this block.
		applyMiningRewardToBalance(s.balanceSheet, block.Header.Beneficiary, s.genesis.MiningReward)

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
func (s *State) validateNextBlock(block SignedBlock) (string, error) {
	s.evHandler("state: WriteNextBlock: validate: block signature")

	if err := block.VerifySignature(); err != nil {
		return zeroHash, fmt.Errorf("invalid block signature")
	}

	s.evHandler("state: WriteNextBlock: validate: hash solved")

	hash := block.Hash()
	if !isHashSolved(s.genesis.Difficulty, hash) {
		return zeroHash, fmt.Errorf("%s invalid hash", hash)
	}

	latestBlock := s.CopyLatestBlock()
	nextNumber := latestBlock.Header.Number + 1

	s.evHandler("state: WriteNextBlock: validate: chain not forked")

	// The node who sent this block has a chain that is two or more blocks ahead
	// of ours. This means there has been a fork and we are on the wrong side.
	if block.Header.Number >= (nextNumber + 2) {
		return zeroHash, ErrChainForked
	}

	s.evHandler("state: WriteNextBlock: validate: block number")

	if block.Header.Number != nextNumber {
		return zeroHash, fmt.Errorf("this block is not the next number, got %d, exp %d", block.Header.Number, nextNumber)
	}

	s.evHandler("state: WriteNextBlock: validate: parent hash")

	if block.Header.ParentHash != latestBlock.Hash() {
		return zeroHash, fmt.Errorf("prev block doesn't match our latest, got %s, exp %s", block.Header.ParentHash, latestBlock.Hash())
	}

	s.evHandler("state: WriteNextBlock: validate: transaction signatures")

	for _, tx := range block.Transactions {
		if err := tx.VerifySignature(); err != nil {
			return zeroHash, fmt.Errorf("transaction has invalid signature, %w, tx[%v]", err, tx)
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
	s.txMempool = newTxMempool()
	s.latestBlock = SignedBlock{}
	s.balanceSheet = balanceSheet
	s.dbFile = dbFile

	// Start the peer update operation.
	s.powWorker.signalPeerUpdates()

	return nil
}

// =============================================================================

// CopyGenesis returns a copy of the genesis information.
func (s *State) CopyGenesis() Genesis {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.genesis
}

// CopyLatestBlock returns the current hash of the latest block.
func (s *State) CopyLatestBlock() SignedBlock {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.latestBlock
}

// CopyMempool returns a copy of the mempool.
func (s *State) CopyMempool() []BlockTx {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.txMempool.copy()
}

// CopyBalanceSheet returns a copy of the balance sheet.
func (s *State) CopyBalanceSheet() BalanceSheet {
	s.mu.Lock()
	defer s.mu.Unlock()

	return copyBalanceSheet(s.balanceSheet)
}

// CopyKnownPeers retrieves information about the peer for updating
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

// QueryBalances returns a copy of the set of balances by address.
func (s *State) QueryBalances(address string) BalanceSheet {
	s.mu.Lock()
	defer s.mu.Unlock()

	balanceSheet := copyBalanceSheet(s.balanceSheet)
	for _, tx := range s.txMempool {
		applyTransactionToBalance(balanceSheet, tx)
	}

	for addr := range balanceSheet {
		if address != addr {
			balanceSheet.remove(addr)
		}
	}

	return balanceSheet
}

// QueryMempoolLength returns the current length of the mempool.
func (s *State) QueryMempoolLength() int {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.txMempool.count()
}

// QueryBlocksByNumber returns the set of blocks based on block numbers. This
// function reads the blockchain from disk first.
func (s *State) QueryBlocksByNumber(from uint64, to uint64) []SignedBlock {
	blocks, err := loadBlocksFromDisk(s.dbPath)
	if err != nil {
		return nil
	}

	if from == QueryLastest {
		from = blocks[len(blocks)-1].Header.Number
		to = from
	}

	var out []SignedBlock
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
func (s *State) QueryBlocksByAddress(address string) []SignedBlock {
	blocks, err := loadBlocksFromDisk(s.dbPath)
	if err != nil {
		return nil
	}

	var out []SignedBlock
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
func (s *State) MineNewBlock(ctx context.Context) (SignedBlock, time.Duration, error) {
	var nb Block
	var balanceSheet BalanceSheet

	s.evHandler("worker: runMiningOperation: MINING: create new block")

	// Execute this code inside a lock.
	if err := func() error {
		s.mu.Lock()
		defer s.mu.Unlock()

		// Are there enough transactions in the pool.
		if s.txMempool.count() < s.genesis.TransPerBlock {
			return ErrNotEnoughTransactions
		}

		// Create a new block which owns it's own copy of the transactions.
		nb = newBlock(s.MinerName(), s.genesis.Difficulty, s.genesis.TransPerBlock, s.latestBlock.Block, s.txMempool)

		// Get a copy of the balance sheet.
		balanceSheet = copyBalanceSheet(s.balanceSheet)

		return nil
	}(); err != nil {
		return SignedBlock{}, 0, ErrNotEnoughTransactions
	}

	s.evHandler("worker: runMiningOperation: MINING: update copy of balance sheet")

	// Process the transactions against the balance sheet.
	for _, tx := range nb.Transactions {

		// Apply the balance changes based on this transaction. Set status
		// information for other nodes to process this correctly.
		if err := applyTransactionToBalance(balanceSheet, tx); err != nil {
			s.evHandler("worker: runMiningOperation: MINING: WARNING : %s", err)
			continue
		}

		// Apply the miner tip and gas fee for this transaction.
		applyMiningFeeToBalance(balanceSheet, s.MinerName(), tx)

		// Update the total gas and tip fees.
		nb.Header.TotalGas += tx.Gas
		nb.Header.TotalTip += tx.Tip
	}

	// Apply the miner reward for this block.
	applyMiningRewardToBalance(balanceSheet, s.MinerName(), s.genesis.MiningReward)

	// Attempt to create a new BlockFS by solving the POW puzzle.
	// This can be cancelled.
	blockFS, duration, err := performPOW(ctx, s.genesis.Difficulty, nb, s.privateKey, s.evHandler)
	if err != nil {
		return SignedBlock{}, duration, err
	}

	// Just check one more time we were not cancelled.
	if ctx.Err() != nil {
		return SignedBlock{}, duration, ctx.Err()
	}

	s.evHandler("worker: runMiningOperation: MINING: marshal block for write")

	// Marshal the block for writing to disk.
	blockFSJson, err := json.Marshal(blockFS)
	if err != nil {
		return SignedBlock{}, duration, err
	}

	// Execute this code inside a lock.
	if err := func() error {
		s.mu.Lock()
		defer s.mu.Unlock()

		s.evHandler("worker: runMiningOperation: MINING: write to disk")

		// Write the new block to the chain on disk.
		if _, err := s.dbFile.Write(append(blockFSJson, '\n')); err != nil {
			return err
		}

		s.evHandler("worker: runMiningOperation: MINING: apply new balance sheet")

		s.balanceSheet = balanceSheet
		s.latestBlock = blockFS.SignedBlock

		// Remove the transactions from this block.
		for _, tx := range nb.Transactions {
			s.evHandler("worker: runMiningOperation: MINING: remove from mempool: tx[%s]", tx.Hash())
			s.txMempool.delete(tx)
		}

		return nil
	}(); err != nil {
		return SignedBlock{}, duration, err
	}

	return blockFS.SignedBlock, duration, nil
}
