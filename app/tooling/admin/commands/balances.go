package commands

import (
	"fmt"

	"github.com/ardanlabs/blockchain/foundation/node"
)

// Balances returns the current set of balances.
func Balances(args []string, n *node.Node) error {
	var onlyAct string
	if len(args) == 3 {
		onlyAct = args[2]
	}

	fmt.Printf("LastestBlockHash: %x\n\n", n.LatestBlock().Hash())

	for act, bal := range n.Balances(onlyAct) {
		fmt.Printf("Account: %s  Balance: %d\n", act, bal)
	}

	return nil
}
