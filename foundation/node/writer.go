package node

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

// EventHandler defines a function that is called when events
// occur in the processing of persisting blocks.
type EventHandler func(v string)

// blockWriter manages a goroutine that executes a write block
// call on a timer.
type blockWriter struct {
	node      *Node
	wg        sync.WaitGroup
	shut      chan struct{}
	ticker    time.Ticker
	evHandler EventHandler
}

// newBlockWriter creates a persister for writing transactions
// to a block.
func newBlockWriter(n *Node, interval time.Duration, evHandler EventHandler) *blockWriter {
	ev := func(v string) {
		if evHandler != nil {
			evHandler(v)
		}
	}

	bw := blockWriter{
		node:      n,
		shut:      make(chan struct{}),
		ticker:    *time.NewTicker(interval),
		evHandler: ev,
	}

	bw.wg.Add(1)
	go func() {
		defer bw.wg.Done()
	down:
		for {
			select {
			case <-bw.ticker.C:
				bw.writeBlock()
			case <-bw.shut:
				break down
			}
		}
	}()

	return &bw
}

// shutdown terminates the goroutine performing work.
func (bw *blockWriter) shutdown() {
	bw.evHandler("block writer: stop timer")
	bw.ticker.Stop()

	bw.evHandler("block writer: terminate goroutine")
	close(bw.shut)
	bw.wg.Wait()

	bw.evHandler("block writer: off")
}

// writeBlock performs the work to create a new block from transactions
// in the mempool.
func (bw *blockWriter) writeBlock() {
	bw.evHandler("block writer: started")
	defer bw.evHandler("block writer: completed")

	block, err := bw.node.WriteBlock()
	if err != nil {
		if errors.Is(err, ErrNoTransactions) {
			bw.evHandler("block writer: no transactions in mempool")
			return
		}
		bw.evHandler(fmt.Sprintf("block writer: ERROR %s", err))
		return
	}

	hash := fmt.Sprintf("%x", block.Hash())

	bw.evHandler(fmt.Sprintf("block writer: prevBlk[%x], newBlk[%x], numTrans[%d]", block.Header.PrevBlock, hash, len(block.Transactions)))
}
