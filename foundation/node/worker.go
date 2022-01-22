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

// peerUpdateInterval represents the interval of finding new peer nodes
// and updating the blockchain on disk with missing blocks.
const peerUpdateInterval = time.Minute

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

// runBCWorker creates a blockWriter for writing transactions
// from the mempool to a new block.
func runBCWorker(node *Node, evHandler EventHandler) *bcWorker {
	bw := bcWorker{
		node:         node,
		ticker:       *time.NewTicker(peerUpdateInterval),
		shut:         make(chan struct{}),
		startMining:  make(chan bool, 1),
		cancelMining: make(chan bool, 1),
		txSharing:    make(chan []Tx, maxTxShareRequests),
		evHandler:    evHandler,
		baseURL:      "http://%s/v1/node",
	}

	// Before anything, update the peer list and make sure this
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

	for {
		select {
		case <-bw.ticker.C:
			bw.runPeerOperation()
		case <-bw.shut:
			bw.evHandler("bcWorker: peerOperations: received shut signal")
			return
		}
	}
}

// miningOperations handles mining.
func (bw *bcWorker) miningOperations() {
	bw.evHandler("bcWorker: miningOperations: G started")
	defer bw.evHandler("bcWorker: miningOperations: G completed")

	for {
		select {
		case <-bw.startMining:
			bw.runMiningOperation()
		case <-bw.shut:
			bw.evHandler("bcWorker: miningOperations: received shut signal")
			return
		}
	}
}

// shareTxOperations handles sharing new user transactions.
func (bw *bcWorker) shareTxOperations() {
	bw.evHandler("bcWorker: shareTxOperations: G started")
	defer bw.evHandler("bcWorker: shareTxOperations: G completed")

	for {
		select {
		case txs := <-bw.txSharing:
			bw.runShareTxOperation(txs)
		case <-bw.shut:
			bw.evHandler("bcWorker: shareTxOperations: received shut signal")
			return
		}
	}
}

// =============================================================================

// signalStartMining starts a mining operation. If there is already a signal
// pending in the channel, just return since a mining operation will start.
func (bw *bcWorker) signalStartMining() {
	select {
	case bw.startMining <- true:
	default:
	}
	bw.evHandler("bcWorker: signalStartMining: mining signaled")
}

// signalCancelMining cancels a mining operation. If there is already a signal
// pending in the channel, just return since a mining operation will cancel.
func (bw *bcWorker) signalCancelMining() {
	select {
	case bw.cancelMining <- true:
	default:
	}
	bw.evHandler("bcWorker: signalCancelMining: cancel signaled")
}

// signalShareTransactions queues up a share transaction operation. If
// maxTxShareRequests signals exist in the channel, we won't send these.
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

	for _, peer := range bw.node.CopyKnownPeers() {
		url := fmt.Sprintf("%s/tx/add", fmt.Sprintf(bw.baseURL, peer.Host))
		if err := send(http.MethodPost, url, txs, nil); err != nil {
			bw.evHandler("bcWorker: runShareTxOperation: WARNING: %s", err)
		}
	}
}

// =============================================================================

// runMiningOperation takes all the transactions from the mempool and writes a
// new block to the database.
func (bw *bcWorker) runMiningOperation() {
	bw.evHandler("bcWorker: runMiningOperation: **********: mining started")
	defer bw.evHandler("bcWorker: runMiningOperation: **********: mining completed")

	// Drain the cancel mining channel before starting.
	select {
	case <-bw.cancelMining:
		bw.evHandler("bcWorker: runMiningOperation: **********: drained cancel channel")
	default:
	}

	// Make sure there are at least 2 transactions in the mempool.
	length := bw.node.QueryMempoolLength()
	if length < 2 {
		bw.evHandler("bcWorker: runMiningOperation: **********: not enough transactions to mine: %d", length)
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
			bw.evHandler("bcWorker: runMiningOperation: **********: cancelG: cancel mining")
		case <-ctx.Done():
			bw.evHandler("bcWorker: runMiningOperation: **********: cancelG: context cancelled")
		}
	}()

	// This G is performing the mining.
	go func() {
		bw.evHandler("bcWorker: runMiningOperation: **********: miningG: started")
		defer func() {
			bw.evHandler("bcWorker: runMiningOperation: **********: miningG: completed")
			cancel()
			wg.Done()
		}()

		block, duration, err := bw.node.MineNewBlock(ctx)
		bw.evHandler("bcWorker: runMiningOperation: **********: miningG: mining duration[%v]", duration)

		if err != nil {
			switch {
			case errors.Is(err, ErrNotEnoughTransactions):
				bw.evHandler("bcWorker: runMiningOperation: **********: miningG: WARNING: not enough transactions in mempool")
			case ctx.Err() != nil:
				bw.evHandler("bcWorker: runMiningOperation: **********: miningG: WARNING: mining cancelled")
			default:
				bw.evHandler("bcWorker: runMiningOperation: **********: miningG: ERROR: %s", err)
			}
			return
		}

		// WOW, we mined a block.
		bw.evHandler("bcWorker: runMiningOperation: **********: miningG: MINED BLOCK: prevBlk[%s]: newBlk[%s]: numTrans[%d]", block.Header.PrevBlock, block.Hash(), len(block.Transactions))

		// Send the new block to the network.
		if err := bw.sendBlockToPeers(block); err != nil {
			bw.evHandler("bcWorker: runMiningOperation: **********: miningG: sendBlockToPeers: WARNING %s", err)

			// TODO: I need to potentially re-sync my blockchain on disk.
		}
	}()

	// Wait for both G's to terminate.
	wg.Wait()
}

// sendBlockToPeers takes the new mined block and sends it to all know peers.
func (bw *bcWorker) sendBlockToPeers(block Block) error {
	bw.evHandler("bcWorker: sendBlockToPeers: **********: started")
	defer bw.evHandler("bcWorker: sendBlockToPeers: **********: completed")

	for _, peer := range bw.node.CopyKnownPeers() {
		url := fmt.Sprintf("%s/block/next", fmt.Sprintf(bw.baseURL, peer.Host))

		var status struct {
			Status string `json:"status"`
			Block  Block  `json:"block"`
		}

		if err := send(http.MethodPost, url, block, &status); err != nil {
			return fmt.Errorf("%s: %s", peer.Host, err)
		}

		bw.evHandler("bcWorker: sendBlockToPeers: **********: %s: SENT", peer)
	}

	return nil
}

// =============================================================================

// runPeerOperation updates the peer list and sync's up the database.
func (bw *bcWorker) runPeerOperation() {
	bw.evHandler("bcWorker: runPeerOperation: started")
	defer bw.evHandler("bcWorker: runPeerOperation: completed")

	for _, peer := range bw.node.CopyKnownPeers() {

		// Retrieve the status of this peer.
		peerStatus, err := bw.queryPeerStatus(peer)
		if err != nil {
			bw.evHandler("bcWorker: runPeerOperation: queryPeerStatus: %s: ERROR: %s", peer.Host, err)
		}

		// Add new peers to this nodes list.
		if err := bw.addNewPeers(peerStatus.KnownPeers); err != nil {
			bw.evHandler("bcWorker: runPeerOperation: addNewPeers: %s: ERROR: %s", peer.Host, err)
		}

		// If this peer has blocks we don't have, we need to add them.
		if peerStatus.LatestBlockNumber > bw.node.CopyLatestBlock().Header.Number {
			bw.evHandler("bcWorker: runPeerOperation: writePeerBlocks: %s: latestBlockNumber[%d]", peer.Host, peerStatus.LatestBlockNumber)
			if err := bw.writePeerBlocks(peer); err != nil {
				bw.evHandler("bcWorker: runPeerOperation: writePeerBlocks: %s: ERROR %s", peer.Host, err)
			}
		}
	}
}

// queryPeerStatus looks for new nodes on the blockchain by asking
// known nodes for their peer list. New nodes are added to the list.
func (bw *bcWorker) queryPeerStatus(peer Peer) (PeerStatus, error) {
	bw.evHandler("bcWorker: runPeerOperation: queryPeerStatus: started: %s", peer)
	defer bw.evHandler("bcWorker: runPeerOperation: queryPeerStatus: completed: %s", peer)

	url := fmt.Sprintf("%s/status", fmt.Sprintf(bw.baseURL, peer.Host))

	var ps PeerStatus
	if err := send(http.MethodGet, url, nil, &ps); err != nil {
		return PeerStatus{}, err
	}

	bw.evHandler("bcWorker: runPeerOperation: queryPeerStatus: peer-node[%s]: latest-blknum[%d]: peer-list[%s]", peer, ps.LatestBlockNumber, ps.KnownPeers)

	return ps, nil
}

// addNewPeers takes the list of known peers and makes sure they are included
// in the nodes list of know peers.
func (bw *bcWorker) addNewPeers(knownPeers []Peer) error {
	bw.evHandler("bcWorker: runPeerOperation: addNewPeers: started")
	defer bw.evHandler("bcWorker: runPeerOperation: addNewPeers: completed")

	for _, peer := range knownPeers {
		if err := bw.node.addPeerNode(peer); err != nil {

			// It already exists, nothing to report.
			return nil
		}
		bw.evHandler("bcWorker: runPeerOperation: addNewPeers: add peer nodes: adding peer-node %s", peer)
	}

	return nil
}

// writePeerBlocks queries the specified node asking for blocks this
// node does not have, then writes them to disk.
func (bw *bcWorker) writePeerBlocks(peer Peer) error {
	bw.evHandler("bcWorker: runPeerOperation: writePeerBlocks: **********: started: %s", peer)
	defer bw.evHandler("bcWorker: runPeerOperation: writePeerBlocks: **********: completed: %s", peer)

	from := bw.node.CopyLatestBlock().Header.Number + 1
	url := fmt.Sprintf("%s/block/list/%d/latest", fmt.Sprintf(bw.baseURL, peer.Host), from)

	var blocks []Block
	if err := send(http.MethodGet, url, nil, &blocks); err != nil {
		return err
	}

	bw.evHandler("bcWorker: runPeerOperation: writePeerBlocks: **********: found blocks[%d]", len(blocks))

	for _, block := range blocks {
		bw.evHandler("bcWorker: runPeerOperation: writePeerBlocks: **********: prevBlk[%s]: newBlk[%s]: numTrans[%d]", block.Header.PrevBlock, block.Hash(), len(block.Transactions))

		if err := bw.node.WriteNextBlock(block); err != nil {
			return err
		}
	}

	return nil
}

// =============================================================================

// send is a helper function to send an HTTP request to a node.
func send(method string, url string, dataSend interface{}, dataRecv interface{}) error {
	var req *http.Request

	switch {
	case dataSend != nil:
		data, err := json.Marshal(dataSend)
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

	if dataRecv != nil {
		if err := json.NewDecoder(resp.Body).Decode(dataRecv); err != nil {
			return err
		}
	}

	return nil
}
