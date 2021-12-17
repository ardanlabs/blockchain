package commands

import (
	"fmt"

	"github.com/ardanlabs/blockchain/business/core/chain"
)

// Transactions returns the current set of transactions.
func Transactions(args []string, db *chain.Chain) error {
	for _, tx := range db.TxMempool {
		fmt.Printf("ID: %s  From: %s  To: %s  Value: %d  Data: %s\n",
			tx.ID, tx.From, tx.To, tx.Value, tx.Data)
	}

	return nil
}
