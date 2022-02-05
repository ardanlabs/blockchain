package blockchain

import (
	"crypto/ecdsa"
	"encoding/hex"
	"encoding/json"
	"log"

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
	Data  string `json:"data"`  // Extra data related to the transaction.
}

// Sign uses the specified private key to sign the user transaction.
func (tx UserTx) Sign(privateKey *ecdsa.PrivateKey) (SignedTx, error) {
	data, err := json.Marshal(tx)
	if err != nil {
		log.Fatal(err)
	}
	hash := crypto.Keccak256Hash(data)

	sig, err := crypto.Sign(hash.Bytes(), privateKey)
	if err != nil {
		return SignedTx{}, err
	}

	signedTx := SignedTx{
		UserTx:    tx,
		Signature: hex.EncodeToString(sig),
	}

	return signedTx, nil
}

// SignedTx is a signed version of the user transaction.
type SignedTx struct {
	UserTx
	Signature string `json:"sig"` // Signature of the person who signed the transaction.
}

// =============================================================================

// BlockTx represents the transaction recorded inside the blockchain.
type BlockTx struct {
	SignedTx
	Gas uint `json:"gas"` // Gas fee to recover computation costs paid by the sender.
}

// From extracts the address for the account that signed the transaction.
func (tx BlockTx) From() (string, error) {
	userTx := UserTx{
		To:    tx.To,
		Value: tx.Value,
		Tip:   tx.Tip,
		Data:  tx.Data,
	}

	data, err := json.Marshal(userTx)
	if err != nil {
		return "", err
	}
	hash := crypto.Keccak256Hash(data)

	sig, err := hex.DecodeString(tx.Signature)
	if err != nil {
		return "", err
	}

	publicKey, err := crypto.SigToPub(hash.Bytes(), sig)
	if err != nil {
		return "", err
	}

	return crypto.PubkeyToAddress(*publicKey).String(), nil
}
