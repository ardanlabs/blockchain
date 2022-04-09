// Package state is the core API for the blockchain and implements all the
// business rules and processing.
package state

import (
	"sync"

	"github.com/ardanlabs/blockchain/foundation/blockchain/accounts"
	"github.com/ardanlabs/blockchain/foundation/blockchain/genesis"
	"github.com/ardanlabs/blockchain/foundation/blockchain/mempool"
	"github.com/ardanlabs/blockchain/foundation/blockchain/peer"
	"github.com/ardanlabs/blockchain/foundation/blockchain/storage"
)

/*
	-- Web Application
	Then add a graphical way of seeing the data.

	-- Blockchain
	Use Merkle tree to validate after a new block is created, if a transaction from a client exists.
	Find a way to quickly find transactions for an account in the blockchain.
	Batch new transactions to send across the network. Must maintain mining sync.
	Create a file for each block and then fix the way we handle a forked chain.
	Add integration test for the state package.
	Try a different POW cryptographic problem to solve that could provide more consistency time.
*/

// =============================================================================

// EventHandler defines a function that is called when events
// occur in the processing of persisting blocks.
type EventHandler func(v string, args ...any)

// Worker interface represents the behavior required to be implemented by any
// package providing support for mining, peer updates, and transaction sharing.
type Worker interface {
	Shutdown()
	SignalStartMining()
	SignalCancelMining() (done func())
	SignalShareTx(blockTx storage.BlockTx)
}

// =============================================================================

// Config represents the configuration required to start
// the blockchain node.
type Config struct {
	MinerAccount   storage.Account
	Host           string
	DBPath         string
	SelectStrategy string
	KnownPeers     *peer.PeerSet
	EvHandler      EventHandler
}

// State manages the blockchain database.
type State struct {
	minerAccount storage.Account
	host         string
	dbPath       string
	evHandler    EventHandler
	latestBlock  storage.Block
	mu           sync.Mutex

	knownPeers *peer.PeerSet
	genesis    genesis.Genesis
	mempool    *mempool.Mempool
	storage    *storage.Storage
	accounts   *accounts.Accounts

	Worker Worker
}

// New constructs a new blockchain for data management.
func New(cfg Config) (*State, error) {

	// Build a safe event handler function for use.
	ev := func(v string, args ...any) {
		if cfg.EvHandler != nil {
			cfg.EvHandler(v, args...)
		}
	}

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
	blocks, err := strg.ReadAllBlocks(ev, true)
	if err != nil {
		return nil, err
	}

	// Keep the latest block from the blockchain.
	var latestBlock storage.Block
	if len(blocks) > 0 {
		latestBlock = blocks[len(blocks)-1]
	}

	// Create a new accounts value to manage accounts who transact on
	// the blockchain and apply the genesis information and blocks.
	accounts := accounts.New(genesis, blocks)

	// Construct a mempool with the specified sort strategy.
	mempool, err := mempool.NewWithStrategy(cfg.SelectStrategy)
	if err != nil {
		return nil, err
	}

	// Create the State to provide support for managing the blockchain.
	state := State{
		minerAccount: cfg.MinerAccount,
		host:         cfg.Host,
		dbPath:       cfg.DBPath,
		evHandler:    ev,
		latestBlock:  latestBlock,

		knownPeers: cfg.KnownPeers,
		genesis:    genesis,
		mempool:    mempool,
		storage:    strg,
		accounts:   accounts,
	}

	// The Worker is not set here. The call to worker.Run will assign itself
	// and start everything up and running for the node.

	return &state, nil
}

// Shutdown cleanly brings the node down.
func (s *State) Shutdown() error {

	// Make sure the database file is properly closed.
	defer func() {
		s.storage.Close()
	}()

	// Stop all blockchain writing activity.
	s.Worker.Shutdown()

	return nil
}

// Truncate resets the chain both on disk and in memory. This is used to
// correct an identified fork.
func (s *State) Truncate() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Reset the state of the database.
	s.mempool.Truncate()
	s.accounts.Reset()
	s.latestBlock = storage.Block{}
	s.storage.Reset()

	return nil
}
