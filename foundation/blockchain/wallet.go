package blockchain

import (
	"crypto/ecdsa"
	"encoding/json"
	"log"

	"github.com/ethereum/go-ethereum/crypto"
)

// WalletTxSigned provides a signature from the sender for a transaction.
type WalletTxSigned struct {
	Tx        WalletTx `json:"tx"`
	Signature []byte   `json:"sig"`
}

// WalletTx is what is submitted by a wallet.
type WalletTx struct {
	To    string `json:"to"`
	Value uint   `json:"value"`
	Tip   uint   `json:"tip"`
	Data  string `json:"data"`
}

// Sign generates a Tx that is signed with the specified private key.
func (tx WalletTx) Sign(privateKey *ecdsa.PrivateKey) (WalletTxSigned, error) {
	data, err := json.Marshal(tx)
	if err != nil {
		log.Fatal(err)
	}
	hash := crypto.Keccak256Hash(data)

	sig, err := crypto.Sign(hash.Bytes(), privateKey)
	if err != nil {
		return WalletTxSigned{}, err
	}

	signedTx := WalletTxSigned{
		Tx:        tx,
		Signature: sig,
	}

	return signedTx, nil
}
