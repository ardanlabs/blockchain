package storage

import (
	"crypto/ecdsa"
	"fmt"
	"math/big"

	"github.com/ardanlabs/blockchain/foundation/blockchain/signature"
)

// =============================================================================

// UserTx is the transactional data submitted by a user.
type UserTx struct {
	ID    uint   `json:"id"`    // Unique id for the transaction supplied by the user.
	To    string `json:"to"`    // Address receiving the benefit of the transaction.
	Value uint   `json:"value"` // Monetary value received from this transaction.
	Tip   uint   `json:"tip"`   // Tip offered by the sender as an incentive to mine this transaction.
	Data  []byte `json:"data"`  // Extra data related to the transaction.
}

// NewUserTx constructs a new user transaction.
func NewUserTx(id uint, to string, value uint, tip uint, data []byte) UserTx {
	return UserTx{
		ID:    id,
		To:    to,
		Value: value,
		Tip:   tip,
		Data:  data,
	}
}

// Sign uses the specified private key to sign the user transaction.
func (tx UserTx) Sign(privateKey *ecdsa.PrivateKey) (SignedTx, error) {

	// Sign the hash with the private key to produce a signature.
	v, r, s, err := signature.Sign(tx, privateKey)
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

// =============================================================================

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
	return signature.VerifySignature(tx.UserTx, tx.V, tx.R, tx.S)
}

// FromAddress extracts the address for the account that signed the transaction.
func (tx SignedTx) FromAddress() (string, error) {
	return signature.FromAddress(tx.UserTx, tx.V, tx.R, tx.S)
}

// Signature returns the signature as a string.
func (tx SignedTx) SignatureString() string {
	return signature.SignatureString(tx.V, tx.R, tx.S)
}

// UniqueKey is used to generate a unique key for mempool activity. The key is
// generated based on the UserTx ID field and from address.
func (tx SignedTx) UniqueKey() string {
	from, err := signature.FromAddress(tx.UserTx, tx.V, tx.R, tx.S)
	if err != nil {
		from = "unknown"
	}

	return fmt.Sprintf("%s:%d", from, tx.ID)
}

// =============================================================================

// BlockTx represents the transaction recorded inside the blockchain.
type BlockTx struct {
	SignedTx
	Gas uint `json:"gas"` // Gas fee to recover computation costs paid by the sender.
}

// NewBlockTx constructs a new block transaction.
func NewBlockTx(signedTx SignedTx, gas uint) BlockTx {
	return BlockTx{
		SignedTx: signedTx,
		Gas:      gas,
	}
}
