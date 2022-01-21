package node

import (
	"github.com/google/uuid"
)

// Set of transaction data types.
const (
	TxDataReward = "reward"
)

// Set of transaction status types.
const (
	TxStatusAccepted  = "accepted"
	TxStatusError     = "error"
	TxStatusNew       = "new"
	TxStatusPublished = "published"
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

// ID represents a unique ID in the system.
type ID string

// Tx represents a transaction in the block.
type Tx struct {
	ID          ID     `json:"id"`
	FromAccount string `json:"from"`
	ToAccount   string `json:"to"`
	Value       uint   `json:"value"`
	Data        string `json:"data"`
	Status      string `json:"status"`
	StatusInfo  string `json:"status_info"`
}

// NewTx constructs a new TxRecord.
func NewTx(fromAccount string, toAccount string, value uint, data string) Tx {
	return Tx{
		ID:          ID(uuid.New().String()),
		FromAccount: fromAccount,
		ToAccount:   toAccount,
		Value:       value,
		Data:        data,
		Status:      TxStatusNew,
	}
}
