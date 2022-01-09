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

// peerStatus represents information about the status
// of any given peer.
type peerStatus struct {
	Hash              string              `json:"hash"`
	LatestBlockNumber uint64              `json:"latest_block_number"`
	KnownPeers        map[string]struct{} `json:"known_peers"`
}

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

	for ipPort := range bw.node.KnownPeersList() {

		// Retrieve the status of this peer.
		peer, err := bw.queryPeerStatus(ipPort)
		if err != nil {
			bw.evHandler(fmt.Sprintf("block writer: writeBlocks: queryPeerStatus: ERROR: %s", err))
		}
		bw.evHandler(fmt.Sprintf("block writer: writeBlocks: node %s: latest blknum: %d: peer list %s", ipPort, peer.LatestBlockNumber, peer.KnownPeers))

		// Add new peers to this nodes list.
		for ipPort := range peer.KnownPeers {
			if err := bw.node.AddPeerNode(ipPort); err == nil {
				bw.evHandler(fmt.Sprintf("block writer: findNewNodes: adding node %s", ipPort))
			}
		}

		// If this peer has blocks we don't have, we need to add it.
		if peer.LatestBlockNumber > bw.node.LatestBlock().Header.Number {
			if err := bw.addMissingBlocks(ipPort); err != nil {
				bw.evHandler(fmt.Sprintf("block writer: addMissingBlocks: ERROR %s", err))
			}
		}
	}

	// Write a new block based on the mempool.
	bw.writeMempoolBlock()
}

// queryPeerStatus looks for new nodes on the blockchain by asking
// known nodes for their peer list. New nodes are added to the list.
func (bw *blockWriter) queryPeerStatus(ipPort string) (peerStatus, error) {
	url := fmt.Sprintf("http://%s/v1/node/status", ipPort)

	var client http.Client
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return peerStatus{}, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return peerStatus{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		msg, err := io.ReadAll(resp.Body)
		if err != nil {
			return peerStatus{}, err
		}
		return peerStatus{}, errors.New(string(msg))
	}

	var peer peerStatus
	if err := json.NewDecoder(resp.Body).Decode(&peer); err != nil {
		return peerStatus{}, err
	}

	return peer, nil
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

// addMissingBlocks queries the specified node asking for blocks this
// node does not have.
func (bw *blockWriter) addMissingBlocks(ipPort string) error {

	return nil
}
