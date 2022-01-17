package node

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
)

// maxTxShareRequests represents the max number of pending tx network share
// requests that can be outstanding before share requests are dropped. To keep
// this simple, a buffered channel of this arbitrary number is being used. If
// the channel does become full, requests for new transactions to be shared
// will not be accepted.
const maxTxShareRequests = 100

// =============================================================================

// peerStatus represents information about the status
// of any given peer.
type peerStatus struct {
	Hash              string              `json:"hash"`
	LatestBlockNumber uint64              `json:"latest_block_number"`
	KnownPeers        map[string]struct{} `json:"known_peers"`
}

// =============================================================================

// bcWorker manages a goroutine that executes a write block
// call on a timer.
type bcWorker struct {
	node         *Node
	wg           sync.WaitGroup
	ticker       time.Ticker
	shut         chan struct{}
	startMining  chan bool
	cancelMining chan bool
	txSharing    chan []Tx
	evHandler    EventHandler
	baseURL      string
}

// newBCWorker creates a blockWriter for writing transactions
// from the mempool to a new block.
func newBCWorker(node *Node, evHandler EventHandler) *bcWorker {
	bw := bcWorker{
		node:         node,
		ticker:       *time.NewTicker(10 * time.Second),
		shut:         make(chan struct{}),
		startMining:  make(chan bool, 1),
		cancelMining: make(chan bool, 1),
		txSharing:    make(chan []Tx, maxTxShareRequests),
		evHandler:    evHandler,
		baseURL:      "http://%s/v1/node",
	}

	// Before anything, update the peer list and make/ sure this
	// node's blockchain is up to date.
	bw.runPeerOperation()

	// Load the set of operations we need to run.
	operations := []func(){
		bw.peerOperations,
		bw.miningOperations,
		bw.shareTxOperations,
	}

	// Set waitgroup to match the number of G's we need for the set
	// of operations we have.
	g := len(operations)
	bw.wg.Add(g)

	// We don't want to return until we know all the G's are up and running.
	hasStarted := make(chan bool)

	// Start all the operational G's.
	for _, op := range operations {
		go func(op func()) {
			defer bw.wg.Done()
			hasStarted <- true
			op()
		}(op)
	}

	// Wait for the G's to report they are running.
	for i := 0; i < g; i++ {
		<-hasStarted
	}

	return &bw
}

// shutdown terminates the goroutine performing work.
func (bw *bcWorker) shutdown() {
	bw.evHandler("bcWorker: shutdown: started")
	defer bw.evHandler("bcWorker: shutdown: completed")

	bw.evHandler("bcWorker: shutdown: stop ticker")
	bw.ticker.Stop()

	bw.evHandler("bcWorker: shutdown: signal cancel mining")
	bw.signalCancelMining()

	bw.evHandler("bcWorker: shutdown: terminate goroutines")
	close(bw.shut)
	bw.wg.Wait()
}

// =============================================================================

// peerOperations handles finding new peers.
func (bw *bcWorker) peerOperations() {
	bw.evHandler("bcWorker: peerOperations: G started")
	defer bw.evHandler("bcWorker: peerOperations: G completed")

down:
	for {
		select {
		case <-bw.ticker.C:
			bw.runPeerOperation()
		case <-bw.shut:
			bw.evHandler("bcWorker: peerOperations: received shut signal")
			break down
		}
	}
}

// miningOperations handles mining.
func (bw *bcWorker) miningOperations() {
	bw.evHandler("bcWorker: miningOperations: G started")
	defer bw.evHandler("bcWorker: miningOperations: G completed")

down:
	for {
		select {
		case <-bw.startMining:
			bw.runMiningOperation()
		case <-bw.shut:
			bw.evHandler("bcWorker: miningOperations: received shut signal")
			break down
		}
	}
}

// shareTxOperations handles sharing new user transactions.
func (bw *bcWorker) shareTxOperations() {
	bw.evHandler("bcWorker: shareTxOperations: G started")
	defer bw.evHandler("bcWorker: shareTxOperations: G completed")

down:
	for {
		select {
		case txs := <-bw.txSharing:
			bw.runShareTxOperation(txs)
		case <-bw.shut:
			bw.evHandler("bcWorker: shareTxOperations: received shut signal")
			break down
		}
	}
}

// =============================================================================

// signalStartMining starts a mining operation.
func (bw *bcWorker) signalStartMining() {
	select {
	case bw.startMining <- true:
	default:
	}
	bw.evHandler("bcWorker: signalStartMining: mining signaled")
}

// signalCancelMining cancels a mining operation.
func (bw *bcWorker) signalCancelMining() {
	select {
	case bw.cancelMining <- true:
	default:
	}
	bw.evHandler("bcWorker: signalCancelMining: cancel signaled")
}

// signalShareTransactions queues up a share transaction operation.
func (bw *bcWorker) signalShareTransactions(txs []Tx) {
	select {
	case bw.txSharing <- txs:
		bw.evHandler("bcWorker: signalShareTransactions: share signaled")
	default:
		bw.evHandler("bcWorker: signalShareTransactions: queue full, transactions won't be shared.")
	}
}

// =============================================================================

// runShareTxOperation updates the peer list and sync's up the database.
func (bw *bcWorker) runShareTxOperation(txs []Tx) {
	bw.evHandler("bcWorker: runShareTxOperation: started")
	defer bw.evHandler("bcWorker: runShareTxOperation: completed")

	for ipPort := range bw.node.CopyKnownPeersList() {
		url := fmt.Sprintf("%s/tx/add", fmt.Sprintf(bw.baseURL, ipPort))
		if err := send(http.MethodPost, url, txs, nil); err != nil {
			bw.evHandler("bcWorker: runShareTxOperation: ERROR: %s", err)
		}
	}
}

// =============================================================================

// runMiningOperation takes all the transactions from the mempool and writes a
// new block to the database.
func (bw *bcWorker) runMiningOperation() {
	bw.evHandler("bcWorker: runMiningOperation: mining started")
	defer bw.evHandler("bcWorker: runMiningOperation: mining completed")

	// Drain the cancel mining channel before starting.
	select {
	case <-bw.cancelMining:
		bw.evHandler("bcWorker: runMiningOperation: drained cancel channel")
	default:
	}

	// Make sure there are at least 2 transactions in the mempool.
	length := bw.node.QueryMempoolLength()
	if length < 2 {
		bw.evHandler("bcWorker: runMiningOperation: not enough transactions to mine: %d", length)
		return
	}

	// Create a context so mining can be cancelled. Mining has 2 minutes
	// to find a solution.
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	// Can't return from this function until these G's are complete.
	var wg sync.WaitGroup
	wg.Add(2)

	// This G exists to cancel the mining operation.
	go func() {
		defer func() { cancel(); wg.Done() }()

		select {
		case <-bw.cancelMining:
			bw.evHandler("bcWorker: runMiningOperation: cancelG: cancel mining")
		case <-ctx.Done():
			bw.evHandler("bcWorker: runMiningOperation: cancelG: context cancelled")
		}
	}()

	// This G is performing the mining.
	go func() {
		bw.evHandler("bcWorker: runMiningOperation: miningG: started")
		defer func() {
			bw.evHandler("bcWorker: runMiningOperation: miningG: completed")
			cancel()
			wg.Done()
		}()

		block, duration, err := bw.node.MineNewBlock(ctx)
		bw.evHandler("bcWorker: runMiningOperation: miningG: mining duration[%v]", duration)

		if err != nil {
			switch {
			case errors.Is(err, ErrNotEnoughTransactions):
				bw.evHandler("bcWorker: runMiningOperation: miningG: WARNING: not enough transactions in mempool")
			case ctx.Err() != nil:
				bw.evHandler("bcWorker: runMiningOperation: miningG: WARNING: mining cancelled")
			default:
				bw.evHandler("bcWorker: runMiningOperation: miningG: ERROR: %s", err)
			}
			return
		}

		// WOW, we mined a block.
		bw.evHandler("bcWorker: runMiningOperation: miningG: prevBlk[%s]: newBlk[%s]: numTrans[%d]", block.Header.PrevBlock, block.Hash(), len(block.Transactions))

		// TODO: SEND NEW BLOCK TO THE CHAIN!!!!
	}()

	// Wait for both G's to terminate.
	wg.Wait()
}

// =============================================================================

// runPeerOperation updates the peer list and sync's up the database.
func (bw *bcWorker) runPeerOperation() {
	bw.evHandler("bcWorker: runPeerOperation: started")
	defer bw.evHandler("bcWorker: runPeerOperation: completed")

	for ipPort := range bw.node.CopyKnownPeersList() {

		// Retrieve the status of this peer.
		peer, err := bw.queryPeerStatus(ipPort)
		if err != nil {
			bw.evHandler("bcWorker: runPeerOperation: queryPeerStatus: ERROR: %s", err)
		}

		// Add new peers to this nodes list.
		if err := bw.addNewPeers(peer.KnownPeers); err != nil {
			bw.evHandler("bcWorker: runPeerOperation: addNewPeers: ERROR: %s", err)
		}

		// If this peer has blocks we don't have, we need to add them.
		if peer.LatestBlockNumber > bw.node.CopyLatestBlock().Header.Number {
			bw.evHandler("bcWorker: runPeerOperation: writePeerBlocks: latestBlockNumber[%d]", peer.LatestBlockNumber)
			if err := bw.writePeerBlocks(ipPort); err != nil {
				bw.evHandler("bcWorker: runPeerOperation: writePeerBlocks: ERROR %s", err)
			}
		}
	}
}

// queryPeerStatus looks for new nodes on the blockchain by asking
// known nodes for their peer list. New nodes are added to the list.
func (bw *bcWorker) queryPeerStatus(ipPort string) (peerStatus, error) {
	bw.evHandler("bcWorker: runPeerOperation: queryPeerStatus: started: %s", ipPort)
	defer bw.evHandler("bcWorker: runPeerOperation: queryPeerStatus: completed: %s", ipPort)

	url := fmt.Sprintf("%s/status", fmt.Sprintf(bw.baseURL, ipPort))

	var peer peerStatus
	if err := send(http.MethodGet, url, nil, &peer); err != nil {
		return peerStatus{}, err
	}

	bw.evHandler("bcWorker: runPeerOperation: queryPeerStatus: node[%s]: latest-blknum[%d]: peer-list[%s]", ipPort, peer.LatestBlockNumber, peer.KnownPeers)

	return peer, nil
}

// addNewPeers takes the set of known peers and makes sure they are included
// in the nodes list of know peers.
func (bw *bcWorker) addNewPeers(knownPeers map[string]struct{}) error {
	bw.evHandler("bcWorker: runPeerOperation: addNewPeers: started")
	defer bw.evHandler("bcWorker: runPeerOperation: addNewPeers: completed")

	for ipPort := range knownPeers {
		if err := bw.node.addPeerNode(ipPort); err != nil {

			// It already exists, nothing to report.
			return nil
		}
		bw.evHandler("bcWorker: runPeerOperation: addNewPeers: add peer nodes: adding node %s", ipPort)
	}

	return nil
}

// writePeerBlocks queries the specified node asking for blocks this
// node does not have.
func (bw *bcWorker) writePeerBlocks(ipPort string) error {
	bw.evHandler("bcWorker: runPeerOperation: writePeerBlocks: started: %s", ipPort)
	defer bw.evHandler("bcWorker: runPeerOperation: writePeerBlocks: completed: %s", ipPort)

	from := bw.node.CopyLatestBlock().Header.Number + 1
	url := fmt.Sprintf("%s/blocks/list/%d/latest", fmt.Sprintf(bw.baseURL, ipPort), from)

	var pbs []PeerBlock
	if err := send(http.MethodGet, url, nil, &pbs); err != nil {
		return err
	}

	bw.evHandler("bcWorker: runPeerOperation: writePeerBlocks: found blocks[%d]", len(pbs))

	for _, pb := range pbs {
		bw.evHandler("bcWorker: runPeerOperation: writePeerBlocks: prevBlk[%s]: newBlk[%s]: numTrans[%d]", pb.Header.PrevBlock, pb.Header.ThisBlock, len(pb.Transactions))

		_, err := bw.node.writeNewBlockFromPeer(pb)
		if err != nil {
			return err
		}
	}

	return nil
}

// =============================================================================

// send is a helper function to send an HTTP request to a node.
func send(method string, url string, input interface{}, output interface{}) error {
	var req *http.Request

	switch {
	case input != nil:
		data, err := json.Marshal(input)
		if err != nil {
			return err
		}
		req, err = http.NewRequest(method, url, bytes.NewReader(data))
		if err != nil {
			return err
		}

	default:
		var err error
		req, err = http.NewRequest(method, url, nil)
		if err != nil {
			return err
		}
	}

	var client http.Client
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNoContent {
		return nil
	}

	if resp.StatusCode != http.StatusOK {
		msg, err := io.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		return errors.New(string(msg))
	}

	if output != nil {
		if err := json.NewDecoder(resp.Body).Decode(output); err != nil {
			return err
		}
	}

	return nil
}
