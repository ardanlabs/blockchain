package state

import "github.com/ardanlabs/blockchain/foundation/blockchain/storage"

// UpsertWalletTransaction accepts a transaction from a wallet for inclusion.
func (s *State) UpsertWalletTransaction(signedTx storage.SignedTx) error {
	if err := s.validateTransaction(signedTx); err != nil {
		return err
	}

	tx := storage.NewBlockTx(signedTx, s.genesis.GasPrice)

	if err := s.mempool.Upsert(tx); err != nil {
		return err
	}

	s.Worker.SignalShareTx(tx)
	s.Worker.SignalStartMining()

	return nil
}

// UpsertNodeTransaction accepts a transaction from a node for inclusion.
func (s *State) UpsertNodeTransaction(tx storage.BlockTx) error {
	if err := s.validateTransaction(tx.SignedTx); err != nil {
		return err
	}

	if err := s.mempool.Upsert(tx); err != nil {
		return err
	}

	s.Worker.SignalStartMining()

	return nil
}

// =============================================================================

// validateTransaction takes the signed transaction and validates it has
// a proper signature and other aspects of the data.
func (s *State) validateTransaction(signedTx storage.SignedTx) error {
	if err := signedTx.Validate(); err != nil {
		return err
	}

	if err := s.db.ValidateNonce(signedTx); err != nil {
		return err
	}

	return nil
}
