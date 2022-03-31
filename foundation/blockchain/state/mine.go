package state

import (
	"context"
	"time"

	"github.com/ardanlabs/blockchain/foundation/blockchain/storage"
)

// MineNewBlock attempts to create a new block with a proper hash that can become
// the next block in the chain.
func (s *State) MineNewBlock(ctx context.Context) (storage.Block, time.Duration, error) {
	s.evHandler("state: MineNewBlock: MINING: check mempool count")

	// Are there enough transactions in the pool.
	if s.mempool.Count() < s.genesis.TransPerBlock {
		return storage.Block{}, 0, ErrNotEnoughTransactions
	}

	s.evHandler("state: MineNewBlock: MINING: create new block: pick %d", s.genesis.TransPerBlock)

	// Create a new block which owns it's own copy of the transactions.
	trans := s.mempool.PickBest(s.genesis.TransPerBlock)
	b := storage.NewBlock(s.minerAccount, s.genesis.Difficulty, s.genesis.TransPerBlock, s.RetrieveLatestBlock(), trans)

	s.evHandler("state: MineNewBlock: MINING: perform POW")

	// Attempt to create a new BlockFS by solving the POW puzzle. This can be cancelled.
	blockFS, duration, err := b.PerformPOW(ctx, s.genesis.Difficulty, s.evHandler)
	if err != nil {
		return storage.Block{}, duration, err
	}

	// Just check one more time we were not cancelled.
	if ctx.Err() != nil {
		return storage.Block{}, duration, ctx.Err()
	}

	s.evHandler("state: MineNewBlock: MINING: update local state")

	if err := s.updateLocalState(blockFS); err != nil {
		return storage.Block{}, duration, err
	}

	return blockFS.Block, duration, nil
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
	done := s.worker.signalCancelMining()
	defer func() {
		s.evHandler("state: MinePeerBlock: signal runMiningOperation to terminate")
		done()
	}()

	hash, err := block.ValidateBlock(s.latestBlock, s.evHandler)
	if err != nil {
		return err
	}

	blockFS := storage.BlockFS{
		Hash:  hash,
		Block: block,
	}

	return s.updateLocalState(blockFS)
}

// =============================================================================

// updateLocalState takes the blockFS and updates the current state of the
// chain, including adding the block to disk.
func (s *State) updateLocalState(blockFS storage.BlockFS) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.evHandler("state: updateLocalState: write to disk")

	// Write the new block to the chain on disk.
	if err := s.storage.Write(blockFS); err != nil {
		return err
	}
	s.latestBlock = blockFS.Block

	s.evHandler("state: updateLocalState: update accounts and remove from mempool")

	// Process the transactions and update the accounts.
	for _, tx := range blockFS.Block.Transactions {
		s.evHandler("state: updateLocalState: tx[%s] update and remove", tx)

		// Apply the balance changes based on this transaction.
		if err := s.accounts.ApplyTransaction(blockFS.Block.Header.MinerAccount, tx); err != nil {
			s.evHandler("state: updateLocalState: WARNING : %s", err)
			continue
		}

		// Remove this transaction from the mempool.
		s.mempool.Delete(tx)
	}

	s.evHandler("state: updateLocalState: apply mining reward")

	// Apply the mining reward for this block.
	s.accounts.ApplyMiningReward(blockFS.Block.Header.MinerAccount)

	return nil
}
