package commands

import (
	"context"
	"fmt"
	"time"

	"github.com/ardanlabs/blockchain/foundation/node"
)

// Transactions returns the current set of transactions.
func Transactions(args []string, n *node.Node) error {
	var sub string
	if len(args) > 2 {
		sub = args[2]
	}

	switch sub {
	case "seed":
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		var txs []node.Tx
		txs = append(txs, node.NewTx(0, "bill_kennedy", "bill_kennedy", 3, node.TxDataReward))
		txs = append(txs, node.NewTx(0, "bill_kennedy", "bill_kennedy", 703, node.TxDataReward))

		if err := n.SignalAddTransactions(ctx, txs); err != nil {
			return err
		}
		if err := n.SignalBlockWork(ctx); err != nil {
			return err
		}
		if err := waitForBlock(n, 1, ctx); err != nil {
			return err
		}

		txs = []node.Tx{}
		txs = append(txs, node.NewTx(0, "bill_kennedy", "babayaga", 2000, ""))
		txs = append(txs, node.NewTx(0, "bill_kennedy", "bill_kennedy", 100, node.TxDataReward))
		txs = append(txs, node.NewTx(0, "babayaga", "bill_kennedy", 1, ""))
		txs = append(txs, node.NewTx(0, "babayaga", "ceasar", 1000, ""))
		txs = append(txs, node.NewTx(0, "babayaga", "bill_kennedy", 50, ""))
		txs = append(txs, node.NewTx(0, "bill_kennedy", "bill_kennedy", 600, node.TxDataReward))

		if err := n.SignalAddTransactions(ctx, txs); err != nil {
			return err
		}
		if err := n.SignalBlockWork(ctx); err != nil {
			return err
		}
		if err := waitForBlock(n, 2, ctx); err != nil {
			return err
		}

	default:
		var acct string
		if len(args) == 3 {
			acct = args[2]
		}

		fmt.Println("-----------------------------------------------------------------------------------------")
		for i, block := range n.QueryBlocksByAccount(acct) {
			fmt.Println("Block:", i)
			fmt.Printf("Prev Block: %s\n", block.Header.PrevBlock)
			fmt.Printf("Block: %s\n", block.Hash())
			for _, tx := range block.Transactions {
				fmt.Printf("From: %s  To: %s  Value: %d  Data: %s\n",
					tx.Record.From, tx.Record.To, tx.Record.Value, tx.Record.Data)
			}
			fmt.Println("-----------------------------------------------------------------------------------------")
		}
	}

	return nil
}

func waitForBlock(node *node.Node, number uint64, ctx context.Context) error {
	for {
		fmt.Printf("waiting for block %d ...\n", number)

		block := node.CopyLatestBlock()
		if block.Header.Number == number {
			fmt.Printf("Block %d Persisted\n", number)
			fmt.Printf("BlockHash: %s\n\n", block.Hash())
			return nil
		}

		if ctx.Err() != nil {
			return ctx.Err()
		}
		time.Sleep(100 * time.Millisecond)
	}
}
