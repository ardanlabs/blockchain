package commands

import (
	"fmt"
	"strconv"

	"github.com/ardanlabs/blockchain/foundation/database"
)

// Transactions returns the current set of transactions.
func Transactions(args []string, db *database.DB) error {
	var sub string
	if len(args) > 2 {
		sub = args[2]
	}

	switch sub {
	case "seed":
		var txs []database.Tx
		txs = append(txs, database.NewTx("bill_kennedy", "bill_kennedy", 3, database.TxDataReward))
		txs = append(txs, database.NewTx("bill_kennedy", "bill_kennedy", 703, database.TxDataReward))

		for _, tx := range txs {
			if err := db.AddMempool(tx); err != nil {
				return err
			}
		}

		if err := db.Persist(); err != nil {
			return err
		}
		fmt.Println("Block 0 Persisted")
		fmt.Printf("BlockHash: %x\n\n", db.LastestBlock())

		txs = []database.Tx{}
		txs = append(txs, database.NewTx("bill_kennedy", "babayaga", 2000, ""))
		txs = append(txs, database.NewTx("bill_kennedy", "bill_kennedy", 100, database.TxDataReward))
		txs = append(txs, database.NewTx("babayaga", "bill_kennedy", 1, ""))
		txs = append(txs, database.NewTx("babayaga", "ceasar", 1000, ""))
		txs = append(txs, database.NewTx("babayaga", "bill_kennedy", 50, ""))
		txs = append(txs, database.NewTx("bill_kennedy", "bill_kennedy", 600, database.TxDataReward))

		for _, tx := range txs {
			if err := db.AddMempool(tx); err != nil {
				return err
			}
		}

		if err := db.Persist(); err != nil {
			return err
		}
		fmt.Println("Block 1 Persisted")
		fmt.Printf("BlockHash: %x\n\n", db.LastestBlock())

	case "add":
		fmt.Printf("LastestBlockHash: %x\n\n", db.LastestBlock())

		from := args[3]
		to := args[4]
		value, _ := strconv.Atoi(args[5])
		data := args[6]
		tx := database.NewTx(from, to, uint(value), data)
		if err := db.AddMempool(tx); err != nil {
			return err
		}
		fmt.Println("Transaction added")

		if err := db.Persist(); err != nil {
			return err
		}
		fmt.Println("Transaction persisted")
		fmt.Printf("LastestBlockHash: %x\n\n", db.LastestBlock())

	default:
		var acct string
		if len(args) == 3 {
			acct = args[2]
		}

		fmt.Println("-----------------------------------------------------------------------------------------")
		for i, block := range db.Blocks(acct) {
			h, _ := block.Hash()
			fmt.Println("Block:", i)
			fmt.Printf("Prev Block: %x\n", block.Header.PrevBlock)
			fmt.Printf("Block: %x\n", h)
			for _, tx := range block.Transactions {
				fmt.Printf("From: %s  To: %s  Value: %d  Data: %s\n",
					tx.From, tx.To, tx.Value, tx.Data)
			}
			fmt.Println("-----------------------------------------------------------------------------------------")
		}
	}

	return nil
}
