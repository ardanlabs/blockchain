package chain

import (
	"fmt"
	"strings"
)

// TxError represents an error on a transaction.
type TxError struct {
	Tx  Tx
	Err error
}

// Error implements the error interface.
func (txe *TxError) Error() string {
	return txe.Err.Error()
}

// TxErrors represents a set of transaction errors.
type TxErrors map[string]TxError

// Error implements the error interface.
func (txes TxErrors) Error() string {
	var sb strings.Builder
	for k, v := range txes {
		s := fmt.Sprintf("{ID: %s, ERROR: %s}", k, v.Err)
		sb.WriteString(s)
	}

	return sb.String()
}
