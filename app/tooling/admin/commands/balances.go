package commands

import (
	"fmt"

	"github.com/ardanlabs/blockchain/foundation/database"
)

// Balances returns the current set of balances.
func Balances(args []string, db *database.DB) error {
	var onlyAct string
	if len(args) == 3 {
		onlyAct = args[2]
	}

	fmt.Printf("LastestBlockHash: %x\n\n", db.LastestBlock())

	for act, bal := range db.Balances(onlyAct) {
		fmt.Printf("Account: %s  Balance: %d\n", act, bal)
	}

	return nil
}
