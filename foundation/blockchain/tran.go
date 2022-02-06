package blockchain

import (
	"crypto/ecdsa"
	"encoding/json"
	"math/big"

	"github.com/ethereum/go-ethereum/crypto"
)

// Set of transaction data types.
const (
	TxDataReward = "reward"
)

// =============================================================================

// TxError represents an error on a transaction.
type TxError struct {
	Tx  BlockTx
	Err error
}

// Error implements the error interface.
func (txe *TxError) Error() string {
	return txe.Err.Error()
}

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
	hash := crypto.Keccak256Hash(data)

	sig, err := crypto.Sign(hash.Bytes(), privateKey)
	if err != nil {
		return SignedTx{}, err
	}

	signedTx := SignedTx{
		UserTx: tx,
		V:      (&big.Int{}).SetUint64(uint64(sig[64])),
		R:      (&big.Int{}).SetBytes(sig[:32]),
		S:      (&big.Int{}).SetBytes(sig[32:64]),
	}

	return signedTx, nil
}

// SignedTx is a signed version of the user transaction.
type SignedTx struct {
	UserTx
	V *big.Int `json:"v"` // Last byte of the signature.
	R *big.Int `json:"r"` // First 32 bytes of the signature.
	S *big.Int `json:"s"` // Next 32 bytes of the signature.
}

// VerifySignature verifies the signature conforms to the Secp256k1 standard.
func (tx SignedTx) VerifySignature() bool {
	return crypto.ValidateSignatureValues(byte(tx.V.Uint64()), tx.R, tx.S, true)
}

// =============================================================================

// BlockTx represents the transaction recorded inside the blockchain.
type BlockTx struct {
	SignedTx
	Gas uint `json:"gas"` // Gas fee to recover computation costs paid by the sender.
}

// From extracts the address for the account that signed the transaction.
func (tx BlockTx) From() (string, error) {
	data, err := json.Marshal(tx.UserTx)
	if err != nil {
		return "", err
	}
	hash := crypto.Keccak256Hash(data)

	sig := make([]byte, 65)
	copy(sig, tx.R.Bytes())
	copy(sig[32:], tx.S.Bytes())
	sig[64] = byte(tx.V.Uint64())

	publicKey, err := crypto.SigToPub(hash.Bytes(), sig)
	if err != nil {
		return "", err
	}

	return crypto.PubkeyToAddress(*publicKey).String(), nil
}
