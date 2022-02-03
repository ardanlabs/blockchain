package blockchain

import (
	"encoding/json"
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

// Tx represents the basic unit of record for the things of value being recorded.
type Tx struct {
	ID        string `json:"id"`        // Unique ID for the transaction to help with mempool lookups.
	To        string `json:"to"`        // Address receiving the benefit of the transaction.
	Value     uint   `json:"value"`     // Monetary value received from this transaction.
	Tip       uint   `json:"tip"`       // Tip offered by the sender as an incentive to mine this transaction.
	Gas       uint   `json:"gas"`       // Gas fee to recover computation costs paid by the sender.
	Data      string `json:"data"`      // Extra data related to the transaction.
	Signature []byte `json:"signature"` // Signature of the person who signed transaction.
}

// From extracts the address for the account that signed the transaction.
func (tx Tx) From() string {
	walletTx := WalletTx{
		To:    tx.To,
		Value: tx.Value,
		Tip:   tx.Tip,
		Data:  tx.Data,
	}

	data, err := json.Marshal(walletTx)
	if err != nil {
		return ""
	}
	hash := crypto.Keccak256Hash(data)
	publicKey, err := crypto.SigToPub(hash.Bytes(), tx.Signature)
	if err != nil {
		return ""
	}

	return crypto.PubkeyToAddress(*publicKey).String()
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
func (tm TxMempool) Add(id string, tx Tx) {
	if _, exists := tm[id]; !exists {
		tm[id] = tx
	}
}

// Delete removed a transaction from the mempool.
func (tm TxMempool) Delete(id string) {
	delete(tm, id)
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
