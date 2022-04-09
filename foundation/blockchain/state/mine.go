package state

import (
	"context"
	"errors"

	"github.com/ardanlabs/blockchain/foundation/blockchain/storage"
)

// ErrNotEnoughTransactions is returned when a block is requested to be created
// and there are not enough transactions.
var ErrNotEnoughTransactions = errors.New("not enough transactions in mempool")

// =============================================================================

// MineNewBlock attempts to create a new block with a proper hash that can become
// the next block in the chain.
func (s *State) MineNewBlock(ctx context.Context) (storage.Block, error) {
	s.evHandler("state: MineNewBlock: MINING: check mempool count")

	// Are there enough transactions in the pool.
	if s.mempool.Count() < s.genesis.TransPerBlock {
		return storage.Block{}, ErrNotEnoughTransactions
	}

	s.evHandler("state: MineNewBlock: MINING: perform POW")

	// Attempt to create a new block by solving the POW puzzle. This can be cancelled.
	trans := s.mempool.PickBest(s.genesis.TransPerBlock)
	block, err := storage.POW(ctx, s.minerAccount, s.genesis.Difficulty, s.RetrieveLatestBlock(), trans, s.evHandler)
	if err != nil {
		return storage.Block{}, err
	}

	// Just check one more time we were not cancelled.
	if ctx.Err() != nil {
		return storage.Block{}, ctx.Err()
	}

	s.evHandler("state: MineNewBlock: MINING: update local state")

	if err := s.updateLocalState(block); err != nil {
		return storage.Block{}, err
	}

	return block, nil
}

// MinePeerBlock takes a block received from a peer, validates it and
// if that passes, writes the block to disk.
func (s *State) MinePeerBlock(block storage.Block) error {
	s.evHandler("state: MinePeerBlock: started : block[%s]", block.Hash())
	defer s.evHandler("state: MinePeerBlock: completed")

	// If the runMiningOperation function is being executed it needs to stop
	// immediately. The G executing runMiningOperation will not return from the
	// function until done is called. That allows this function to complete
	// its state changes before a new mining operation takes place.
	done := s.Worker.SignalCancelMining()
	defer func() {
		s.evHandler("state: MinePeerBlock: signal runMiningOperation to terminate")
		done()
	}()

	if err := block.ValidateBlock(s.latestBlock, s.evHandler); err != nil {
		return err
	}

	return s.updateLocalState(block)
}

// =============================================================================

// updateLocalState takes the blockFS and updates the current state of the
// chain, including adding the block to disk.
func (s *State) updateLocalState(block storage.Block) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.evHandler("state: updateLocalState: write to disk")

	// Write the new block to the chain on disk.
	if err := s.storage.Write(storage.NewBlockFS(block)); err != nil {
		return err
	}
	s.latestBlock = block

	s.evHandler("state: updateLocalState: update accounts and remove from mempool")

	// Process the transactions and update the accounts.
	for _, tx := range block.Trans.Leafs {
		s.evHandler("state: updateLocalState: tx[%s] update and remove", tx.Value)

		// Apply the balance changes based on this transaction.
		if err := s.accounts.ApplyTransaction(block.Header.MinerAccount, tx.Value); err != nil {
			s.evHandler("state: updateLocalState: WARNING : %s", err)
			continue
		}

		// Remove this transaction from the mempool.
		s.mempool.Delete(tx.Value)
	}

	s.evHandler("state: updateLocalState: apply mining reward")

	// Apply the mining reward for this block.
	s.accounts.ApplyMiningReward(block.Header.MinerAccount)

	return nil
}
