package storage

import (
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"time"

	"github.com/ardanlabs/blockchain/foundation/blockchain/signature"
)

// =============================================================================

// UserTx is the transactional data submitted by a user.
type UserTx struct {
	Nonce uint   `json:"nonce"` // Unique id for the transaction supplied by the user.
	To    string `json:"to"`    // Address receiving the benefit of the transaction.
	Value uint   `json:"value"` // Monetary value received from this transaction.
	Tip   uint   `json:"tip"`   // Tip offered by the sender as an incentive to mine this transaction.
	Data  []byte `json:"data"`  // Extra data related to the transaction.
}

// NewUserTx constructs a new user transaction.
func NewUserTx(nonce uint, to string, value uint, tip uint, data []byte) (UserTx, error) {
	if !isAddress(to) {
		return UserTx{}, fmt.Errorf("to address is not properly formatted")
	}

	userTx := UserTx{
		Nonce: nonce,
		To:    to,
		Value: value,
		Tip:   tip,
		Data:  data,
	}

	return userTx, nil
}

// Sign uses the specified private key to sign the user transaction.
func (tx UserTx) Sign(privateKey *ecdsa.PrivateKey) (SignedTx, error) {

	// Validate the to address incase the UserTx value was hand constructed.
	if !isAddress(tx.To) {
		return SignedTx{}, fmt.Errorf("to address is not properly formatted")
	}

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

// SignatureString returns the signature as a string.
func (tx SignedTx) SignatureString() string {
	return signature.SignatureString(tx.V, tx.R, tx.S)
}

// String implements the fmt.Stringer interface for logging.
func (tx SignedTx) String() string {
	from, err := signature.FromAddress(tx.UserTx, tx.V, tx.R, tx.S)
	if err != nil {
		from = "unknown"
	}

	return fmt.Sprintf("%s:%d", from, tx.Nonce)
}

// =============================================================================

// BlockTx represents the transaction recorded inside the blockchain.
type BlockTx struct {
	SignedTx
	TimeStamp uint64 `json:"timestamp"` // The time the transaction was received.
	Gas       uint   `json:"gas"`       // Gas fee to recover computation costs paid by the sender.
}

// NewBlockTx constructs a new block transaction.
func NewBlockTx(signedTx SignedTx, gas uint) BlockTx {
	return BlockTx{
		SignedTx:  signedTx,
		TimeStamp: uint64(time.Now().UTC().Unix()),
		Gas:       gas,
	}
}
