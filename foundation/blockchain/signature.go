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

// ardanID is an arbitrary number for signing messages. This will make it
// clear that the signature comes from the Ardan blockchain.
// Ethereum and Bitcoin do this as well, but they use the value of 27.
const ardanID = 29

// Hash returns a unique string for the value.
func Hash(value interface{}) string {
	data, err := json.Marshal(value)
	if err != nil {
		return zeroHash
	}

	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}

// Sign uses the specified private key to sign the user transaction.
func Sign(value interface{}, privateKey *ecdsa.PrivateKey) (v, r, s *big.Int, err error) {

	// Prepare the transaction for signing.
	data, err := Stamp(value)
	if err != nil {
		return nil, nil, nil, err
	}

	// Sign the hash with the private key to produce a signature.
	sig, err := crypto.Sign(data, privateKey)
	if err != nil {
		return nil, nil, nil, err
	}

	// Convert the 65 byte signature into the [R|S|V] format.
	v, r, s = toSignatureValues(sig)

	return v, r, s, nil
}

// Stamp returns a hash of 32 bytes that represents this user
// transaction with the Ardan stamp embedded into the final hash.
func Stamp(value interface{}) ([]byte, error) {

	// Marshal the data.
	data, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}

	// Hash the transaction data into a 32 byte array. This will provide
	// a data length consistency with all transactions.
	txHash := crypto.Keccak256Hash(data)

	// Convert the stamp into a slice of bytes. This stamp is
	// used so signatures we produce when signing transactions
	// are always unique to the Ardan blockchain.
	stamp := []byte("\x19Ardan Signed Message:\n32")

	// Hash the stamp and txHash together in a final 32 byte array
	// that represents the transaction data.
	tran := crypto.Keccak256Hash(stamp, txHash.Bytes())

	return tran.Bytes(), nil
}

// VerifySignature verifies the signature conforms to our standards and
// is associated with the data claimed to be signed.
func VerifySignature(value interface{}, v, r, s *big.Int) error {

	// Check the recovery id is either 0 or 1.
	uintV := v.Uint64() - ardanID
	if uintV != 0 && uintV != 1 {
		return errors.New("invalid recovery id")
	}

	// Check the signature values are valid.
	if !crypto.ValidateSignatureValues(byte(uintV), r, s, false) {
		return errors.New("invalid signature values")
	}

	// Prepare the transaction for recovery and validation.
	tran, err := Stamp(value)
	if err != nil {
		return err
	}

	// Convert the [R|S|V] format into the original 65 bytes.
	sig := toSignatureBytes(v, r, s)

	// Capture the uncompressed public key associated with this signature.
	sigPublicKey, err := crypto.Ecrecover(tran, sig)
	if err != nil {
		return fmt.Errorf("ecrecover, %w", err)
	}

	// Check that the given public key created the signature over the data.
	rs := sig[:crypto.RecoveryIDOffset]
	if !crypto.VerifySignature(sigPublicKey, tran, rs) {
		return errors.New("invalid signature")
	}

	return nil
}

// FromAddress extracts the address for the account that signed the transaction.
func FromAddress(value interface{}, v, r, s *big.Int) (string, error) {

	// Prepare the transaction for public key extraction.
	tran, err := Stamp(value)
	if err != nil {
		return "", err
	}

	// Convert the [R|S|V] format into the original 65 bytes.
	sig := toSignatureBytes(v, r, s)

	// Capture the public key associated with this signature.
	publicKey, err := crypto.SigToPub(tran, sig)
	if err != nil {
		return "", err
	}

	// Extract the account address from the public key.
	return crypto.PubkeyToAddress(*publicKey).String(), nil
}

// SignatureString returns the signature as a string.
func SignatureString(v, r, s *big.Int) string {
	return "0x" + hex.EncodeToString(toSignatureBytesForDisplay(v, r, s))
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
