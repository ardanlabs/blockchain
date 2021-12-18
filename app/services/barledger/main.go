package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/ardanlabs/blockchain/business/core/chain"
	"github.com/ardanlabs/blockchain/foundation/logger"
	"go.uber.org/zap"
)

// build is the git version of this program. It is set using build flags in the makefile.
var build = "develop"

func main() {

	// Construct the application logger.
	log, err := logger.New("BAR-LED")
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

	// =========================================================================
	// App Starting

	log.Infow("starting service", "version", build)
	defer log.Infow("shutdown complete")

	if err := hacking(); err != nil {
		return err
	}

	// =========================================================================
	// Start/Stop Service

	// Make a channel to listen for an interrupt or terminate signal from the OS.
	// Use a buffered channel because the signal package requires it.
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)

	// Blocking main and waiting for shutdown.
	<-shutdown
	log.Infow("shutdown", "status", "shutdown started")
	defer log.Infow("shutdown", "status", "shutdown complete")

	return nil
}

func hacking() error {
	db, err := chain.New()
	if err != nil {
		return err
	}
	defer db.Close()

	data, _ := json.MarshalIndent(db, "", "    ")
	fmt.Println(string(data))

	tx := chain.NewTx("babayaga", "bill_kennedy", 10, "whisky drink")
	db.Add(tx)

	data, _ = json.MarshalIndent(db, "", "    ")
	fmt.Println(string(data))

	return nil
}
