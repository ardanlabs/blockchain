package storage

import (
	"bytes"
	"crypto/ecdsa"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"time"

	"github.com/ardanlabs/blockchain/foundation/blockchain/signature"
)

// =============================================================================

// UserTx is the transactional data submitted by a user.
type UserTx struct {
	Nonce uint      `json:"nonce"` // Unique id for the transaction supplied by the user.
	ToID  AccountID `json:"to"`    // Account receiving the benefit of the transaction.
	Value uint      `json:"value"` // Monetary value received from this transaction.
	Tip   uint      `json:"tip"`   // Tip offered by the sender as an incentive to mine this transaction.
	Data  []byte    `json:"data"`  // Extra data related to the transaction.
}

// NewUserTx constructs a new user transaction.
func NewUserTx(nonce uint, toID AccountID, value uint, tip uint, data []byte) (UserTx, error) {
	if !toID.IsAccountID() {
		return UserTx{}, fmt.Errorf("to account is not properly formatted")
	}

	userTx := UserTx{
		Nonce: nonce,
		ToID:  toID,
		Value: value,
		Tip:   tip,
		Data:  data,
	}

	return userTx, nil
}

// Sign uses the specified private key to sign the user transaction.
func (tx UserTx) Sign(privateKey *ecdsa.PrivateKey) (SignedTx, error) {

	// Validate the to account incase the UserTx value was hand constructed.
	if !tx.ToID.IsAccountID() {
		return SignedTx{}, fmt.Errorf("to account is not properly formatted")
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

// SignedTx is a signed version of the user transaction used internally by
// the blockchain.
type SignedTx struct {
	UserTx
	V *big.Int `json:"v"` // Recovery identifier, either 29 or 30 with ardanID.
	R *big.Int `json:"r"` // First coordinate of the ECDSA signature.
	S *big.Int `json:"s"` // Second coordinate of the ECDSA signature.
}

// Validate verifies the transaction has a proper signature that conforms to our
// standards and is associated with the data claimed to be signed. It also
// checks the format of the to account.
func (tx SignedTx) Validate() error {
	if err := signature.VerifySignature(tx.UserTx, tx.V, tx.R, tx.S); err != nil {
		return err
	}

	if !tx.ToID.IsAccountID() {
		return errors.New("invalid account for to account")
	}

	return nil
}

// FromAccount extracts the account that signed the transaction.
func (tx SignedTx) FromAccount() (AccountID, error) {
	address, err := signature.FromAddress(tx.UserTx, tx.V, tx.R, tx.S)
	return AccountID(address), err
}

// SignatureString returns the signature as a string.
func (tx SignedTx) SignatureString() string {
	return signature.SignatureString(tx.V, tx.R, tx.S)
}

// String implements the fmt.Stringer interface for logging.
func (tx SignedTx) String() string {
	from, err := tx.FromAccount()
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

// Hash implements the merkle Hashable interface for providing a hash
// of a block transaction.
func (tx BlockTx) Hash() ([]byte, error) {
	str := signature.Hash(tx)
	return hex.DecodeString(str)
}

// Equals implements the merkle Hashable interface for providing an equality
// check between two block transactions. If the nonce and signatures are the
// same, the two blocks are the same.
func (tx BlockTx) Equals(otherTx BlockTx) bool {
	txSig := signature.ToSignatureBytes(tx.V, tx.R, tx.S)
	otherTxSig := signature.ToSignatureBytes(otherTx.V, otherTx.R, otherTx.S)

	return tx.Nonce == otherTx.Nonce && bytes.Equal(txSig, otherTxSig)
}
