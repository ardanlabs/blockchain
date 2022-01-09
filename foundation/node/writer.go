package node

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
)

// blockWriter manages a goroutine that executes a write block
// call on a timer.
type blockWriter struct {
	node      *Node
	wg        sync.WaitGroup
	shut      chan struct{}
	ticker    time.Ticker
	evHandler EventHandler
}

// newBlockWriter creates a blockWriter for writing transactions
// from the mempool to a new block.
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
				bw.writeBlocks()
			case <-bw.shut:
				break down
			}
		}
	}()

	return &bw
}

// shutdown terminates the goroutine performing work.
func (bw *blockWriter) shutdown() {
	bw.evHandler("block writer: shutdown: stop timer")
	bw.ticker.Stop()

	bw.evHandler("block writer: shutdown: terminate goroutine")
	close(bw.shut)
	bw.wg.Wait()

	bw.evHandler("block writer: shutdown: off")
}

// writeBlock performs the work to create a new block from transactions
// in the mempool.
func (bw *blockWriter) writeBlocks() {
	bw.evHandler("block writer: started")
	defer bw.evHandler("block writer: completed")

	// Find new nodes and update the known peer list.
	if err := bw.findNewNodes(); err != nil {
		bw.evHandler(fmt.Sprintf("block writer: writeBlocks: findNewNodes: ERROR: %s", err))
	}

	// Write a new block based on the mempool.
	bw.writeMempoolBlock()
}

// findNewNodes looks for new nodes on the blockchain by asking
// known nodes for their peer list. New nodes are added to the list.
func (bw *blockWriter) findNewNodes() error {
	for ipPort := range bw.node.QueryKnownPeers() {
		url := fmt.Sprintf("http://%s/v1/node/peers", ipPort)

		var client http.Client
		req, err := http.NewRequest(http.MethodGet, url, nil)
		if err != nil {
			return err
		}

		resp, err := client.Do(req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			msg, err := io.ReadAll(resp.Body)
			if err != nil {
				return err
			}
			return errors.New(string(msg))
		}

		peers := make(map[string]struct{})
		if err := json.NewDecoder(resp.Body).Decode(&peers); err != nil {
			return err
		}

		bw.evHandler(fmt.Sprintf("block writer: findNewNodes: node %s sent peer list %s", ipPort, peers))

		for ipPort := range peers {
			if err := bw.node.AddPeerNode(ipPort); err == nil {
				bw.evHandler(fmt.Sprintf("block writer: findNewNodes: adding node %s", ipPort))
			}
		}
	}

	return nil
}

// writeMempoolBlock takes all the transactions from the mempool
// and writes a new block to the database.
func (bw *blockWriter) writeMempoolBlock() {
	block, err := bw.node.WriteNewBlock()
	if err != nil {
		if errors.Is(err, ErrNoTransactions) {
			bw.evHandler("block writer: writeMempoolBlock: no transactions in mempool")
			return
		}
		bw.evHandler(fmt.Sprintf("block writer: writeMempoolBlock: ERROR %s", err))
		return
	}

	hash := fmt.Sprintf("%x", block.Hash())

	bw.evHandler(fmt.Sprintf("block writer: writeMempoolBlock: prevBlk[%x]: newBlk[%x]: numTrans[%d]", block.Header.PrevBlock, hash, len(block.Transactions)))
}
