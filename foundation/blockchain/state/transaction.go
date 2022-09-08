package state

import "github.com/ardanlabs/blockchain/foundation/blockchain/database"

// UpsertWalletTransaction accepts a transaction from a wallet for inclusion.
func (s *State) UpsertWalletTransaction(signedTx database.SignedTx) error {

	// CORE NOTE: Check the signed transaction has a proper signature, the
	// from matches the signature, and there is a valid account format for the
	// from and to fields. It's up to the wallet to make sure the account has a
	// proper balance and nonce. Fees will be taken if this transaction is mined
	// into a block and those types of validation fail.

	if err := signedTx.Validate(); err != nil {
		return err
	}

	const oneUnitOfGas = 1
	tx := database.NewBlockTx(signedTx, s.genesis.GasPrice, oneUnitOfGas)
	if err := s.mempool.Upsert(tx); err != nil {
		return err
	}

	s.Worker.SignalShareTx(tx)
	s.Worker.SignalStartMining()

	return nil
}

// UpsertNodeTransaction accepts a transaction from a node for inclusion.
func (s *State) UpsertNodeTransaction(tx database.BlockTx) error {

	// Just check the signed transaction has a proper signature and valid
	// account for the recipient.
	if err := tx.Validate(); err != nil {
		return err
	}

	if err := s.mempool.Upsert(tx); err != nil {
		return err
	}

	s.Worker.SignalStartMining()

	return nil
}
