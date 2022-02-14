package blockchain

import (
	"crypto/ecdsa"
	"math/big"
)

// Set of transaction data types.
const (
	TxDataReward = "reward"
)

// =============================================================================

// UserTx is the transactional data submitted by a user.
type UserTx struct {
	To    string `json:"to"`    // Address receiving the benefit of the transaction.
	Value uint   `json:"value"` // Monetary value received from this transaction.
	Tip   uint   `json:"tip"`   // Tip offered by the sender as an incentive to mine this transaction.
	Data  []byte `json:"data"`  // Extra data related to the transaction.
}

// Sign uses the specified private key to sign the user transaction.
func (tx UserTx) Sign(privateKey *ecdsa.PrivateKey) (SignedTx, error) {

	// Sign the hash with the private key to produce a signature.
	v, r, s, err := Sign(tx, privateKey)
	if err != nil {
		return SignedTx{}, err
	}

	// Construct the signed transaction.
	signedTx := SignedTx{
		UserTx: tx,
		V:      v,
		R:      r,
		S:      s,
	}

	return signedTx, nil
}

// SignedTx is a signed version of the user transaction.
type SignedTx struct {
	UserTx
	V *big.Int `json:"v"` // Recovery identifier, either 29 or 30 with ardanID.
	R *big.Int `json:"r"` // First coordinate of the ECDSA signature.
	S *big.Int `json:"s"` // Second coordinate of the ECDSA signature.
}

// VerifySignature verifies the signature conforms to our standards and
// is associated with the data claimed to be signed.
func (tx SignedTx) VerifySignature() error {
	return VerifySignature(tx, tx.V, tx.R, tx.S)
}

// FromAddress extracts the address for the account that signed the transaction.
func (tx SignedTx) FromAddress() (string, error) {
	return FromAddress(tx, tx.V, tx.R, tx.S)
}

// Signature returns the signature as a string.
func (tx SignedTx) SignatureString() string {
	return SignatureString(tx.V, tx.R, tx.S)
}

// =============================================================================

// BlockTx represents the transaction recorded inside the blockchain.
type BlockTx struct {
	SignedTx
	Gas uint `json:"gas"` // Gas fee to recover computation costs paid by the sender.
}

// Hash returns a unique string for the value.
func (tx BlockTx) Hash() string {
	return Hash(tx)
}
