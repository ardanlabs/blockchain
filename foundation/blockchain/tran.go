package blockchain

import (
	"crypto/ecdsa"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/crypto"
)

// Set of transaction data types.
const (
	TxDataReward = "reward"
)

// ardanID is an arbitrary number for signing messages. This will make it
// clear that the signature comes from the Ardan blockchain.
// Ethereum and Bitcoin do this as well, but they use the value of 27.
const ardanID = 29

// This string ensures that any account signature being generated is only valid
// for the Ardan blockchain. Ethereum does this as well.
// "\x19Ethereum Signed Message:\n" + length(message) + message
const ardanSignature = "\x19Ardan Signed Message:\n32"

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
	data, err := json.Marshal(tx)
	if err != nil {
		return SignedTx{}, err
	}

	// This first hash forces the data for the digest to be 32 bytes long.
	dataHash := crypto.Keccak256Hash(data)
	hash := crypto.Keccak256Hash([]byte(ardanSignature), dataHash.Bytes())

	sig, err := crypto.Sign(hash.Bytes(), privateKey)
	if err != nil {
		return SignedTx{}, err
	}

	v, r, s := toSignatureValues(sig)

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

	// Check the recovery id is either 0 or 1.
	v := tx.V.Uint64() - ardanID
	if v != 0 && v != 1 {
		return errors.New("invalid recovery id")
	}

	// Validate the signature values are valid.
	if !crypto.ValidateSignatureValues(byte(v), tx.R, tx.S, true) {
		return errors.New("invalid signature values")
	}

	// Marshal and hash the user data to validate the signature.
	data, err := json.Marshal(tx.UserTx)
	if err != nil {
		return fmt.Errorf("marshal data, %w", err)
	}
	dataHash := crypto.Keccak256Hash(data)
	hash := crypto.Keccak256Hash([]byte(ardanSignature), dataHash.Bytes())

	// Capture the public key associated with this data and signature.
	sig := toSignatureBytes(tx.V, tx.R, tx.S)
	sigPublicKey, err := crypto.Ecrecover(hash.Bytes(), sig)
	if err != nil {
		return fmt.Errorf("ecrecover, %w", err)
	}

	// Check that the given public key created the signature over the data.
	if !crypto.VerifySignature(sigPublicKey, hash.Bytes(), sig[:crypto.RecoveryIDOffset]) {
		return errors.New("invalid signature")
	}

	return nil
}

// =============================================================================

// BlockTx represents the transaction recorded inside the blockchain.
type BlockTx struct {
	SignedTx
	Gas uint `json:"gas"` // Gas fee to recover computation costs paid by the sender.
}

// FromAddress extracts the address for the account that signed the transaction.
func (tx BlockTx) FromAddress() (string, error) {
	data, err := json.Marshal(tx.UserTx)
	if err != nil {
		return "", err
	}
	dataHash := crypto.Keccak256Hash(data)
	hash := crypto.Keccak256Hash([]byte(ardanSignature), dataHash.Bytes())

	publicKey, err := crypto.SigToPub(hash.Bytes(), toSignatureBytes(tx.V, tx.R, tx.S))
	if err != nil {
		return "", err
	}

	return crypto.PubkeyToAddress(*publicKey).String(), nil
}

// Signature returns the signature as a string.
func (tx BlockTx) Signature() string {
	return "0x" + hex.EncodeToString(toSignatureBytesForDisplay(tx.V, tx.R, tx.S))
}

// Hash returns a unique string for the value.
func (tx BlockTx) Hash() string {
	data, err := json.Marshal(tx)
	if err != nil {
		return zeroHash
	}

	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}

// =============================================================================

// toSignatureValues converts the signature into the r, s, v values.
func toSignatureValues(sig []byte) (v, r, s *big.Int) {
	r = new(big.Int).SetBytes(sig[:32])
	s = new(big.Int).SetBytes(sig[32:64])
	v = new(big.Int).SetBytes([]byte{sig[64] + ardanID})

	return v, r, s
}

// toSignatureBytes converts the r, s, v values into a slice of bytes
// with the removal of the ardanID.
func toSignatureBytes(v, r, s *big.Int) []byte {
	sig := make([]byte, crypto.SignatureLength)

	copy(sig, r.Bytes())
	copy(sig[32:], s.Bytes())
	sig[64] = byte(v.Uint64() - ardanID)

	return sig
}

// toSignatureBytesForDisplay converts the r, s, v values into a slice of bytes.
func toSignatureBytesForDisplay(v, r, s *big.Int) []byte {
	sig := make([]byte, crypto.SignatureLength)

	copy(sig, r.Bytes())
	copy(sig[32:], s.Bytes())
	sig[64] = byte(v.Uint64())

	return sig
}
