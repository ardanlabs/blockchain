package node

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
)

// Represents the different types of transaction work that can be performed.
const (
	twAdd = "add"
)

// tranWork is signaled to the worker goroutine to perform transaction work.
type tranWork struct {
	action string
	txs    []Tx
}

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
	evHandler    EventHandler
}

// newBCWorker creates a blockWriter for writing transactions
// from the mempool to a new block.
func newBCWorker(node *Node, evHandler EventHandler) *bcWorker {
	ev := func(v string) {
		if evHandler != nil {
			evHandler(v)
		}
	}

	bw := bcWorker{
		node:         node,
		ticker:       *time.NewTicker(10 * time.Second),
		shut:         make(chan struct{}),
		startMining:  make(chan bool, 1),
		cancelMining: make(chan bool, 1),
		evHandler:    ev,
	}

	// Let's update the peer list and blocks.
	bw.updatePeersAndBlocks()

	// Add the two G's we are about to create.
	g := 2
	bw.wg.Add(g)
	started := make(chan bool)

	// This G handles finding new peers.
	go func() {
		bw.evHandler("bcWorker: updatePeersAndBlocks: G started")
		started <- true
		defer func() {
			bw.evHandler("bcWorker: updatePeersAndBlocks: G completed")
			bw.wg.Done()
		}()
	down:
		for {
			select {
			case <-bw.ticker.C:
				bw.evHandler("bcWorker: updatePeersAndBlocks: received ticker signal")
				bw.updatePeersAndBlocks()
			case <-bw.shut:
				bw.evHandler("bcWorker: updatePeersAndBlocks: received shut signal")
				break down
			}
		}
	}()

	// This G handles mining.
	go func() {
		bw.evHandler("bcWorker: mineNewBlock: G started")
		started <- true
		defer func() {
			bw.evHandler("bcWorker: mineNewBlock: G completed")
			bw.wg.Done()
		}()
	down:
		for {
			select {
			case <-bw.startMining:
				bw.evHandler("bcWorker: mineNewBlock: received start signal")
				bw.mineNewBlock()
			case <-bw.shut:
				bw.evHandler("bcWorker: mineNewBlock: received shut signal")
				break down
			}
		}
	}()

	// Wait for the two G's to report they are running.
	for i := 0; i < g; i++ {
		<-started
	}

	return &bw
}

// shutdown terminates the goroutine performing work.
func (bw *bcWorker) shutdown() {
	bw.evHandler("bcWorker: shutdown: stop timer")
	bw.ticker.Stop()

	bw.signalCancelMining()

	bw.evHandler("bcWorker: shutdown: terminate goroutine")
	close(bw.shut)
	bw.wg.Wait()

	bw.evHandler("bcWorker: shutdown: off")
}

// =============================================================================

// signalStartMining starts a mining operation.
func (bw *bcWorker) signalStartMining() error {
	bw.evHandler("bcWorker: signalStartMining: started")
	defer bw.evHandler("bcWorker: signalStartMining: completed")

	select {
	case bw.startMining <- true:
		return nil
	default:
		return errors.New("mining already pending")
	}
}

// signalCancelMining cancels a mining operation.
func (bw *bcWorker) signalCancelMining() error {
	bw.evHandler("bcWorker: signalCancelMining: started")
	defer bw.evHandler("bcWorker: signalCancelMining: completed")

	select {
	case bw.cancelMining <- true:
		return nil
	default:
		return errors.New("cancel already pending")
	}
}

// =============================================================================

// updatePeersAndBlocks updates the peer list and sync's up the database.
func (bw *bcWorker) updatePeersAndBlocks() {
	bw.evHandler("bcWorker: updatePeersAndBlocks: started")
	defer bw.evHandler("bcWorker: updatePeersAndBlocks: completed")

	for ipPort := range bw.node.CopyKnownPeersList() {

		// Retrieve the status of this peer.
		peer, err := bw.queryPeerStatus(ipPort)
		if err != nil {
			bw.evHandler(fmt.Sprintf("bcWorker: manageBlockchain: queryPeerStatus: ERROR: %s", err))
		}

		// Add new peers to this nodes list.
		if err := bw.addNewPeers(peer.KnownPeers); err != nil {
			bw.evHandler(fmt.Sprintf("bcWorker: manageBlockchain: addNewPeers: ERROR: %s", err))
		}

		// If this peer has blocks we don't have, we need to add them.
		if peer.LatestBlockNumber > bw.node.CopyLatestBlock().Header.Number {
			bw.evHandler(fmt.Sprintf("bcWorker: manageBlockchain: writePeerBlocks: latestBlockNumber[%d]", peer.LatestBlockNumber))
			if err := bw.writePeerBlocks(ipPort); err != nil {
				bw.evHandler(fmt.Sprintf("bcWorker: manageBlockchain: writePeerBlocks: ERROR %s", err))
			}
		}
	}
}

// =============================================================================

// mineNewBlock takes all the transactions from the mempool and writes a
// new block to the database.
func (bw *bcWorker) mineNewBlock() {
	bw.evHandler("bcWorker: mineNewBlock: started")
	defer bw.evHandler("bcWorker: mineNewBlock: completed")

	// Make sure there are at least 2 transactions in the mempool.
	length := bw.node.QueryMempoolLength()
	if length < 2 {
		bw.evHandler(fmt.Sprintf("bcWorker: mineNewBlock: not enough transactions to mine: %d", length))
		return
	}

	// Create a context so mining can be cancelled.
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Can't return from this function until these G's are complete.
	var wg sync.WaitGroup
	wg.Add(2)

	// This G exists to cancel the mining operation.
	go func() {
		defer func() { cancel(); wg.Done() }()

		select {
		case <-bw.cancelMining:
			bw.evHandler("bcWorker: mineNewBlock: cancelG: cancelling mining")
		case <-ctx.Done():
			bw.evHandler("bcWorker: mineNewBlock: cancelG: context cancelled")
		}
	}()

	var block Block
	var err error

	// This G is performing the mining.
	go func() {
		defer func() { cancel(); wg.Done() }()

		bw.evHandler("bcWorker: mineNewBlock: miningG: started")
		block, err = bw.node.MineNewBlock(ctx)
	}()

	// Wait for both G's to terminate.
	wg.Wait()

	// Evaluate the result of mining.
	if err != nil {
		switch {
		case errors.Is(err, ErrNotEnoughTransactions):
			bw.evHandler("bcWorker: mineNewBlock: miningG: not enough transactions in mempool")
		case ctx.Err() != nil:
			bw.evHandler("bcWorker: mineNewBlock: miningG: cancelled")
		default:
			bw.evHandler(fmt.Sprintf("bcWorker: mineNewBlock: miningG: ERROR: %s", err))
		}
		return
	}

	// WOW, we mined a block.
	bw.evHandler(fmt.Sprintf("bcWorker: mineNewBlock: prevBlk[%s]: newBlk[%s]: numTrans[%d]", block.Header.PrevBlock, block.Hash(), len(block.Transactions)))

	// TODO: SEND NEW BLOCK TO THE CHAIN!!!!
}

// =============================================================================

// queryPeerStatus looks for new nodes on the blockchain by asking
// known nodes for their peer list. New nodes are added to the list.
func (bw *bcWorker) queryPeerStatus(ipPort string) (peerStatus, error) {
	bw.evHandler("bcWorker: updatePeersAndBlocks: queryPeerStatus: started")
	defer bw.evHandler("bcWorker: updatePeersAndBlocks: queryPeerStatus: cempleted")

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

	bw.evHandler(fmt.Sprintf("bcWorker: updatePeersAndBlocks: queryPeerStatus: node[%s]: latest-blknum[%d]: peer-list[%s]", ipPort, peer.LatestBlockNumber, peer.KnownPeers))

	return peer, nil
}

// addNewPeers takes the set of known peers and makes sure they are included
// in the nodes list of know peers.
func (bw *bcWorker) addNewPeers(knownPeers map[string]struct{}) error {
	bw.evHandler("bcWorker: updatePeersAndBlocks: addNewPeers: started")
	defer bw.evHandler("bcWorker: updatePeersAndBlocks: addNewPeers: cempleted")

	for ipPort := range knownPeers {
		if err := bw.node.addPeerNode(ipPort); err != nil {
			return err
		}
		bw.evHandler(fmt.Sprintf("bcWorker: updatePeersAndBlocks: addNewPeers: add peer nodes: adding node %s", ipPort))
	}

	return nil
}

// writePeerBlocks queries the specified node asking for blocks this
// node does not have.
func (bw *bcWorker) writePeerBlocks(ipPort string) error {
	bw.evHandler("bcWorker: updatePeersAndBlocks: writePeerBlocks: started")
	defer bw.evHandler("bcWorker: updatePeersAndBlocks: writePeerBlocks: cempleted")

	from := bw.node.CopyLatestBlock().Header.Number + 1
	url := fmt.Sprintf("http://%s/v1/blocks/list/%d/latest", ipPort, from)

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

	if resp.StatusCode == http.StatusNoContent {
		bw.evHandler("bcWorker: updatePeersAndBlocks: writePeerBlocks: no new block")
		return nil
	}

	if resp.StatusCode != http.StatusOK {
		msg, err := io.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		return errors.New(string(msg))
	}

	var pbs []PeerBlock
	if err := json.NewDecoder(resp.Body).Decode(&pbs); err != nil {
		return err
	}

	for _, pb := range pbs {
		bw.evHandler(fmt.Sprintf("bcWorker: updatePeersAndBlocks: writePeerBlocks: prevBlk[%s]: newBlk[%s]: numTrans[%d]", pb.Header.PrevBlock, pb.Header.ThisBlock, len(pb.Transactions)))

		_, err := bw.node.writeNewBlockFromPeer(pb)
		if err != nil {
			return err
		}
	}

	return nil
}
