// This program performs administrative tasks for the garage sale service.
package main

import (
	"fmt"
	"os"

	"github.com/ardanlabs/blockchain/app/tooling/admin/commands"
	"github.com/ardanlabs/blockchain/business/sys/database"
	"github.com/ardanlabs/blockchain/foundation/logger"
	"go.uber.org/zap"
)

// build is the git version of this program. It is set using build flags in the makefile.
var build = "develop"

func main() {

	// Construct the application logger.
	log, err := logger.New("ADMIN")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer log.Sync()

	// Perform the startup and shutdown sequence.
	if err := run(log); err != nil {
		log.Errorw("startup", "ERROR", err)
		log.Sync()
		os.Exit(1)
	}
}

func run(log *zap.SugaredLogger) error {
	db, err := database.New("zblock/blocks.db")
	if err != nil {
		return err
	}
	defer db.Close()

	return processCommands(os.Args, db)
}

// processCommands handles the execution of the commands specified on
// the command line.
func processCommands(args []string, db *database.DB) error {
	switch args[1] {
	case "bals":
		if err := commands.Balances(args, db); err != nil {
			return fmt.Errorf("getting balances: %w", err)
		}
	case "trans":
		if err := commands.Transactions(args, db); err != nil {
			return fmt.Errorf("getting transaction: %w", err)
		}
	}

	return nil
}
