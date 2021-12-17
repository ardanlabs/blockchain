package commands

import (
	"fmt"
	"strconv"

	"github.com/ardanlabs/blockchain/business/core/chain"
)

// Transactions returns the current set of transactions.
func Transactions(args []string, db *chain.Chain) error {
	var sub string
	if len(args) > 2 {
		sub = args[2]
	}

	switch sub {
	case "add":
		from := args[3]
		to := args[4]
		value, _ := strconv.Atoi(args[5])
		data := args[6]
		tx := chain.NewTx(from, to, uint(value), data)
		if err := db.Add(tx); err != nil {
			return err
		}
		fmt.Println("transaction added")

	default:
		for _, tx := range db.TxMempool {
			fmt.Printf("ID: %s  From: %s  To: %s  Value: %d  Data: %s\n",
				tx.ID, tx.From, tx.To, tx.Value, tx.Data)
		}
	}

	return nil
}
