package blockchain

import (
	"crypto/ecdsa"
	"encoding/hex"
	"encoding/json"
	"math/big"

	"github.com/ethereum/go-ethereum/crypto"
)

// Set of transaction data types.
const (
	TxDataReward = "reward"
)

// recoveryID is an arbitrary number for signing messages. This will make it
// clear that the signature comes from the Ardan blockchain. The value inside
// the signature can be 0x1d or 0x1e.
const recoveryID = 29

// This ensures that the signatures we generate cannot be used for purposes
// outside of the Ardan blockchain. Ethereum does this.
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

	r, s, v := toSignatureValues(sig)

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
	V *big.Int `json:"v"` // Recovery identifier, either 0 or 1.
	R *big.Int `json:"r"` // First number of the ECDSA signature.
	S *big.Int `json:"s"` // Second number of the ECDSA signature.
}

// VerifySignature verifies the signature conforms to our standards.
func (tx SignedTx) VerifySignature() bool {
	v := tx.V.Uint64() - recoveryID
	if v != 0 && v != 1 {
		return false
	}

	return crypto.ValidateSignatureValues(byte(v), tx.R, tx.S, true)
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

	// This first hash forces the data for the digest to be 32 bytes long.
	dataHash := crypto.Keccak256Hash(data)
	hash := crypto.Keccak256Hash([]byte(ardanSignature), dataHash.Bytes())

	publicKey, err := crypto.SigToPub(hash.Bytes(), toSignatureCryptoBytes(tx.R, tx.S, tx.V))
	if err != nil {
		return "", err
	}

	return crypto.PubkeyToAddress(*publicKey).String(), nil
}

// Signature returns the signature as a string.
func (tx BlockTx) Signature() string {
	return "0x" + hex.EncodeToString(toSignatureBytes(tx.R, tx.S, tx.V))
}

// =============================================================================

// toSignatureValues converts the signature into the r, s, v values.
func toSignatureValues(sig []byte) (r, s, v *big.Int) {
	r = new(big.Int).SetBytes(sig[:32])
	s = new(big.Int).SetBytes(sig[32:64])
	v = new(big.Int).SetBytes([]byte{sig[64] + recoveryID})

	return r, s, v
}

// toSignatureCryptoBytes converts the r, s, v values into a slice of bytes
// with the removal of the recoveryID.
func toSignatureCryptoBytes(r, s, v *big.Int) []byte {
	sig := make([]byte, crypto.SignatureLength)

	copy(sig, r.Bytes())
	copy(sig[32:], s.Bytes())
	sig[64] = byte(v.Uint64() - recoveryID)

	return sig
}

// toSignatureBytes converts the r, s, v values into a slice of bytes.
func toSignatureBytes(r, s, v *big.Int) []byte {
	sig := make([]byte, crypto.SignatureLength)

	copy(sig, r.Bytes())
	copy(sig[32:], s.Bytes())
	sig[64] = byte(v.Uint64())

	return sig
}
