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
	node      *Node
	wg        sync.WaitGroup
	shut      chan struct{}
	ticker    time.Ticker
	evHandler EventHandler
	worker    chan bool
}

// newBCWorker creates a blockWriter for writing transactions
// from the mempool to a new block.
func newBCWorker(node *Node, interval time.Duration, evHandler EventHandler) *bcWorker {
	ev := func(v string) {
		if evHandler != nil {
			evHandler(v)
		}
	}

	bw := bcWorker{
		node:      node,
		shut:      make(chan struct{}),
		ticker:    *time.NewTicker(interval),
		evHandler: ev,
		worker:    make(chan bool),
	}

	// This G allows all work on the blockchain to be single theaded
	// to reduce the concurrency complexity.
	bw.wg.Add(1)
	go func() {
		defer bw.wg.Done()
	down:
		for {
			select {
			case <-bw.ticker.C:
				bw.performWork()
			case <-bw.worker:
				bw.performWork()
			case <-bw.shut:
				break down
			}
		}
	}()

	return &bw
}

// shutdown terminates the goroutine performing work.
func (bw *bcWorker) shutdown() {
	bw.evHandler("bcWorker: shutdown: stop timer")
	bw.ticker.Stop()

	bw.evHandler("bcWorker: shutdown: terminate goroutine")
	close(bw.shut)
	bw.wg.Wait()

	bw.evHandler("bcWorker: shutdown: off")
}

// PerformWork signals the worker G to perform work and waits to
// have confirmation from the worker G based on the context.
func (bw *bcWorker) PerformWork(ctx context.Context) error {
	select {
	case bw.worker <- true:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// performWork performs all the work that needs to be performed against
// the blockchain.
func (bw *bcWorker) performWork() {
	bw.evHandler("bcWorker: performWork: started")
	defer bw.evHandler("bcWorker: performWork: completed")

	// Perform peer related work first.
	for ipPort := range bw.node.CopyKnownPeersList() {

		// Retrieve the status of this peer.
		peer, err := bw.queryPeerStatus(ipPort)
		if err != nil {
			bw.evHandler(fmt.Sprintf("bcWorker: performWork: queryPeerStatus: ERROR: %s", err))
		}

		// Add new peers to this nodes list.
		if err := bw.addNewPeers(peer.KnownPeers); err != nil {
			bw.evHandler(fmt.Sprintf("bcWorker: performWork: addNewPeers: ERROR: %s", err))
		}

		// If this peer has blocks we don't have, we need to add them.
		if peer.LatestBlockNumber > bw.node.CopyLatestBlock().Header.Number {
			if err := bw.writePeerBlocks(ipPort); err != nil {
				bw.evHandler(fmt.Sprintf("bcWorker: performWork: writePeerBlocks: ERROR %s", err))
			}
		}

		// Publish new transactions to the peer.
		if err := bw.publishNewTransactions(ipPort); err != nil {
			bw.evHandler(fmt.Sprintf("bcWorker: performWork: publishNewTransactions: ERROR %s", err))
		}
	}

	// Mine a new block based on current transactions.
	bw.mineNewBlock()
}

// queryPeerStatus looks for new nodes on the blockchain by asking
// known nodes for their peer list. New nodes are added to the list.
func (bw *bcWorker) queryPeerStatus(ipPort string) (peerStatus, error) {
	bw.evHandler("bcWorker: performWork: queryPeerStatus: started")
	defer bw.evHandler("bcWorker: performWork: queryPeerStatus: cempleted")

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

	bw.evHandler(fmt.Sprintf("bcWorker: performWork: queryPeerStatus: node[%s]: latest-blknum[%d]: peer-list[%s]", ipPort, peer.LatestBlockNumber, peer.KnownPeers))

	return peer, nil
}

// addNewPeers takes the set of known peers and makes sure they are included
// in the nodes list of know peers.
func (bw *bcWorker) addNewPeers(knownPeers map[string]struct{}) error {
	bw.evHandler("bcWorker: performWork: addNewPeers: started")
	defer bw.evHandler("bcWorker: performWork: addNewPeers: cempleted")

	for ipPort := range knownPeers {
		if err := bw.node.AddPeerNode(ipPort); err != nil {
			return err
		}
		bw.evHandler(fmt.Sprintf("bcWorker: performWork: add peer nodes: adding node %s", ipPort))
	}

	return nil
}

// writePeerBlocks queries the specified node asking for blocks this
// node does not have.
func (bw *bcWorker) writePeerBlocks(ipPort string) error {
	bw.evHandler("bcWorker: performWork: writePeerBlocks: started")
	defer bw.evHandler("bcWorker: performWork: writePeerBlocks: cempleted")

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
		bw.evHandler("bcWorker: performWork: writePeerBlocks: no new block")
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
		bw.evHandler(fmt.Sprintf("bcWorker: performWork: writePeerBlocks: prevBlk[%s]: newBlk[%s]: numTrans[%d]", pb.Header.PrevBlock, pb.Header.ThisBlock, len(pb.Transactions)))

		_, err := bw.node.writeNewBlockFromPeer(pb)
		if err != nil {
			return err
		}
	}

	return nil
}

// publishNewTransactions sends any new transactions to the specified
// peer for their mempool.
func (bw *bcWorker) publishNewTransactions(ipPort string) error {
	return nil
}

// mineNewBlock takes all the transactions from the mempool and writes a
// new block to the database.
func (bw *bcWorker) mineNewBlock() {
	block, err := bw.node.writeNewBlockFromTransactions()
	if err != nil {
		if errors.Is(err, ErrNoTransactions) {
			bw.evHandler("bcWorker: performWork: mineNewBlock: no transactions in mempool")
			return
		}
		bw.evHandler(fmt.Sprintf("bcWorker: mineNewBlock: writeMempoolBlock: ERROR %s", err))
		return
	}

	bw.evHandler(fmt.Sprintf("bcWorker: mineNewBlock: writeMempoolBlock: prevBlk[%x]: newBlk[%x]: numTrans[%d]", block.Header.PrevBlock, block.Hash(), len(block.Transactions)))
}
