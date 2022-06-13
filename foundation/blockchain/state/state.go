// Package state is the core API for the blockchain and implements all the
// business rules and processing.
package state

import (
	"sync"

	"github.com/ardanlabs/blockchain/foundation/blockchain/database"
	"github.com/ardanlabs/blockchain/foundation/blockchain/genesis"
	"github.com/ardanlabs/blockchain/foundation/blockchain/mempool"
	"github.com/ardanlabs/blockchain/foundation/blockchain/peer"
)

/*
	-- Web Application
	Then add a graphical way of seeing the data.

	-- Chrome Wallet

	-- Blockchain
	On chain fork, only remove the block need to be removed and reset.

	-- Testing
	Fork Test
	Mining Test
*/

// =============================================================================

// The set of different consensus protocols that can be used.
const (
	ConsensusPOW = "POW"
	ConsensusPOA = "POA"
)

// =============================================================================

// EventHandler defines a function that is called when events
// occur in the processing of persisting blocks.
type EventHandler func(v string, args ...any)

// Worker interface represents the behavior required to be implemented by any
// package providing support for mining, peer updates, and transaction sharing.
type Worker interface {
	Shutdown()
	Sync()
	SignalStartMining()
	SignalCancelMining()
	SignalShareTx(blockTx database.BlockTx)
}

// =============================================================================

// Config represents the configuration required to start
// the blockchain node.
type Config struct {
	BeneficiaryID  database.AccountID
	Host           string
	Storage        database.Storage
	Genesis        genesis.Genesis
	SelectStrategy string
	KnownPeers     *peer.PeerSet
	EvHandler      EventHandler
	Consensus      string
}

// State manages the blockchain database.
type State struct {
	mu          sync.RWMutex
	resyncWG    sync.WaitGroup
	allowMining bool

	beneficiaryID database.AccountID
	host          string
	evHandler     EventHandler
	consensus     string

	knownPeers *peer.PeerSet
	storage    database.Storage
	genesis    genesis.Genesis
	mempool    *mempool.Mempool
	db         *database.Database

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

	// Access the storage for the blockchain.
	db, err := database.New(cfg.Genesis, cfg.Storage, ev)
	if err != nil {
		return nil, err
	}

	// Construct a mempool with the specified sort strategy.
	mempool, err := mempool.NewWithStrategy(cfg.SelectStrategy)
	if err != nil {
		return nil, err
	}

	// Create the State to provide support for managing the blockchain.
	state := State{
		beneficiaryID: cfg.BeneficiaryID,
		host:          cfg.Host,
		storage:       cfg.Storage,
		evHandler:     ev,
		consensus:     cfg.Consensus,
		allowMining:   true,

		knownPeers: cfg.KnownPeers,
		genesis:    cfg.Genesis,
		mempool:    mempool,
		db:         db,
	}

	// The Worker is not set here. The call to worker.Run will assign itself
	// and start everything up and running for the node.

	return &state, nil
}

// Shutdown cleanly brings the node down.
func (s *State) Shutdown() error {
	s.evHandler("state: shutdown: started")
	defer s.evHandler("state: shutdown: completed")

	// Make sure the database file is properly closed.
	defer func() {
		s.db.Close()
	}()

	// Stop all blockchain writing activity.
	s.Worker.Shutdown()

	// Wait for any resync to finish.
	s.resyncWG.Wait()

	return nil
}

// IsMiningAllowed identifies if we are allowed to mine blocks. This
// might be turned off if the blockchain needs to be re-synced.
func (s *State) IsMiningAllowed() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.allowMining
}

// TurnMiningOn sets the allowMining flag back to true.
func (s *State) TurnMiningOn() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.allowMining = true
}

// Reorganize corrects an identified fork. No mining is allowed to take place
// while this process is running. New transactions can be placed into the mempool.
func (s *State) Reorganize() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Don't allow mining to continue.
	s.allowMining = false

	// Reset the state of the blockchain node.
	s.db.Reset()

	// Resync the state of the blockchain.
	s.resyncWG.Add(1)
	go func() {
		s.evHandler("state: Resync: started: *****************************")
		defer func() {
			s.TurnMiningOn()
			s.evHandler("state: Resync: completed: *****************************")
			s.resyncWG.Done()
		}()

		s.Worker.Sync()
	}()

	return nil
}
