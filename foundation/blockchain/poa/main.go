package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
)

func init() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds)
}

func main() {
	sim := newSimulation()

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)
	<-shutdown

	log.Println("******* End Simulation *******")
	sim.shutdown()
}
