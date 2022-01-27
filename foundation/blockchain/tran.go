package blockchain

import (
	"github.com/google/uuid"
)

// Set of transaction data types.
const (
	TxDataReward = "reward"
)

// Set of transaction status types.
const (
	TxStatusAccepted = "accepted" // Accepted and should be applied to the balance.
	TxStatusError    = "error"    // An error occured and should not be applied to the balance.
	TxStatusNew      = "new"      // The transaction is newly added to the mempool.
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

// Tx represents a transaction in the block.
type Tx struct {
	ID         string `json:"id"`          // Unique ID for the transaction to help with mempool lookups.
	From       string `json:"from"`        // The account this transaction is from.
	To         string `json:"to"`          // The account receiving the benefit of the transaction.
	Value      uint   `json:"value"`       // The monetary value received from this transaction.
	Data       string `json:"data"`        // Extra data related to the transaction.
	GasPrice   uint   `json:"gas_price"`   // The amount of gas offered to execute the transaction.
	GasLimit   uint   `json:"gas_limit"`   // The minimum amount of gas required to execute the transaction.
	Status     string `json:"status"`      // The final status of the transaction to help reconcile balances.
	StatusInfo string `json:"status_info"` // Extra information related to the state.
}

// NewTx constructs a new TxRecord.
func NewTx(from string, to string, value uint, data string) Tx {
	return Tx{
		ID:     uuid.New().String(),
		From:   from,
		To:     to,
		Value:  value,
		Data:   data,
		Status: TxStatusNew,
	}
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
