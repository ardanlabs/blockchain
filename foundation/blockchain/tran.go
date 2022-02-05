package blockchain

import (
	"crypto/ecdsa"
	"encoding/hex"
	"encoding/json"
	"log"
	"sort"

	"github.com/ethereum/go-ethereum/crypto"
)

// Set of transaction data types.
const (
	TxDataReward = "reward"
)

// =============================================================================

// TxError represents an error on a transaction.
type TxError struct {
	Tx  Tx
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

// SignedTx provides a signature from the sender for a transaction.
type SignedTx struct {
	UserTx
	Signature string `json:"sig"` // Signature of the person who signed the transaction.
}

// =============================================================================

// Tx represents the transaction recorded inside the blockchain.
type Tx struct {
	SignedTx
	Gas uint `json:"gas"` // Gas fee to recover computation costs paid by the sender.
}

// From extracts the address for the account that signed the transaction.
func (tx Tx) From() (string, error) {
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

// =============================================================================

// TxMempool represents a cache of transactions each with a unique id.
type TxMempool map[string]Tx

// NewTxMempool constructs a new info set to manage node peer information.
func NewTxMempool() TxMempool {
	return make(TxMempool)
}

// Count returns the current number of transaction in the pool.
func (tm TxMempool) Count() int {
	return len(tm)
}

// Add adds a new transaction to the mempool.
func (tm TxMempool) Add(sig string, tx Tx) {
	if _, exists := tm[sig]; !exists {
		tm[sig] = tx
	}
}

// Delete removed a transaction from the mempool.
func (tm TxMempool) Delete(sig string) {
	delete(tm, sig)
}

// Copy returns a list of the current transaction in the pool.
func (tm TxMempool) Copy() []Tx {
	cpy := make([]Tx, 0, len(tm))
	for _, tx := range tm {
		cpy = append(cpy, tx)
	}
	return cpy
}

// CopyBestByTip returns a list of the best transactions for the next
// mining operation based on the offered tip. The caller specifies
// how many transaction they want.
func (tm TxMempool) CopyBestByTip(amount int) []Tx {
	txs := tm.Copy()
	sort.Sort(byTip(txs))

	cpy := make([]Tx, amount)
	for i := 0; i < amount; i++ {
		cpy[i] = txs[i]
	}
	return cpy
}

// =============================================================================

// byTip provides sorting support by the transaction tip value.
type byTip []Tx

// Len returns the number of transactions in the list.
func (bt byTip) Len() int {
	return len(bt)
}

// Less returns true or false based on the tip value between two transactions.
func (bt byTip) Less(i, j int) bool {
	return bt[j].Tip < bt[i].Tip
}

// Swap moves transactions in the order of the tip value.
func (bt byTip) Swap(i, j int) {
	bt[i], bt[j] = bt[j], bt[i]
}
