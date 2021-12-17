package commands

import (
	"fmt"

	"github.com/ardanlabs/blockchain/business/core/chain"
)

// Balances returns the current set of balances.
func Balances(args []string, db *chain.Chain) error {
	var onlyAct string
	if len(args) == 3 {
		onlyAct = args[2]
	}

	for act, bal := range db.Balances {
		if onlyAct == "" || onlyAct == act {
			fmt.Printf("Account: %s  Balance: %d\n", act, bal)
		}
	}

	return nil
}
