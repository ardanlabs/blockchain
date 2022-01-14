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
	node      *Node
	wg        sync.WaitGroup
	shut      chan struct{}
	ticker    time.Ticker
	evHandler EventHandler
	worker    chan bool
	trans     chan tranWork
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
		ticker:    *time.NewTicker(interval),
		shut:      make(chan struct{}),
		worker:    make(chan bool),
		trans:     make(chan tranWork),
		evHandler: ev,
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
				bw.manageBlockchain()
			case <-bw.worker:
				bw.manageBlockchain()
			case tw := <-bw.trans:
				bw.manageTransactions(tw)
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

// =============================================================================

// SignalAddTransactions signals the worker to add the specified transactions
// to the mempool. It waits for confirmation the signal has been received.
func (bw *bcWorker) SignalAddTransactions(ctx context.Context, txs []Tx) error {
	tw := tranWork{
		action: "add",
		txs:    txs,
	}

	select {
	case bw.trans <- tw:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// SignalBlockWork signals the worker to perform blockchain work
// and waits to have confirmation that the worker has started.
func (bw *bcWorker) SignalBlockWork(ctx context.Context) error {
	select {
	case bw.worker <- true:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// =============================================================================

// manageTransactions performs all the transaction work that needs
// to be performed on the node outside of the manageBlockchain code.
func (bw *bcWorker) manageTransactions(tw tranWork) {
	bw.evHandler("bcWorker: manageTransactions: started")
	defer bw.evHandler("bcWorker: manageTransactions: completed")

	switch tw.action {
	case twAdd:
		bw.evHandler(fmt.Sprintf("bcWorker: manageTransactions: addTransaction: %v", tw.txs))
		bw.node.addTransactions(tw.txs)
	}

	// Publish new transactions to the peer.
	for ipPort := range bw.node.CopyKnownPeersList() {
		if err := bw.publishNewTransactions(ipPort); err != nil {
			bw.evHandler(fmt.Sprintf("bcWorker: manageBlockchain: publishNewTransactions: ERROR %s", err))
		}
	}
}

// publishNewTransactions sends any new transactions to the specified
// peer for their mempool.
func (bw *bcWorker) publishNewTransactions(ipPort string) error {
	bw.evHandler("bcWorker: manageTransactions: publishNewTransactions: started")
	defer bw.evHandler("bcWorker: manageTransactions: publishNewTransactions: completed")

	// Extract the record part of the tx for delivery.
	txs := bw.node.QueryMempool(TxStatusNew)
	records := make([]TxRecord, len(txs))
	for i, tx := range txs {
		if tx.Status == TxStatusNew {
			records[i] = tx.Record
		}
	}

	// Marshal the transactions to send.
	txJson, err := json.Marshal(txs)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("http://%s/v1/tx/add", ipPort)
	b := bytes.NewReader(txJson)

	var client http.Client
	req, err := http.NewRequest(http.MethodPost, url, b)
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

	bw.node.updateTransactions(txs, TxStatusPublished)
	bw.evHandler(fmt.Sprintf("bcWorker: manageTransactions: %d transaction published", len(txs)))

	return nil
}

// =============================================================================

// manageBlockchain performs all the blockchain work that needs
// to be performed on the blockchain.
func (bw *bcWorker) manageBlockchain() {
	bw.evHandler("bcWorker: manageBlockchain: started")
	defer bw.evHandler("bcWorker: manageBlockchain: completed")

	// Perform peer related work first.
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
			if err := bw.writePeerBlocks(ipPort); err != nil {
				bw.evHandler(fmt.Sprintf("bcWorker: manageBlockchain: writePeerBlocks: ERROR %s", err))
			}
		}
	}

	// Mine a new block based on current transactions.
	bw.mineNewBlock()
}

// queryPeerStatus looks for new nodes on the blockchain by asking
// known nodes for their peer list. New nodes are added to the list.
func (bw *bcWorker) queryPeerStatus(ipPort string) (peerStatus, error) {
	bw.evHandler("bcWorker: manageBlockchain: queryPeerStatus: started")
	defer bw.evHandler("bcWorker: manageBlockchain: queryPeerStatus: cempleted")

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

	bw.evHandler(fmt.Sprintf("bcWorker: manageBlockchain: queryPeerStatus: node[%s]: latest-blknum[%d]: peer-list[%s]", ipPort, peer.LatestBlockNumber, peer.KnownPeers))

	return peer, nil
}

// addNewPeers takes the set of known peers and makes sure they are included
// in the nodes list of know peers.
func (bw *bcWorker) addNewPeers(knownPeers map[string]struct{}) error {
	bw.evHandler("bcWorker: manageBlockchain: addNewPeers: started")
	defer bw.evHandler("bcWorker: manageBlockchain: addNewPeers: cempleted")

	for ipPort := range knownPeers {
		if err := bw.node.addPeerNode(ipPort); err != nil {
			return err
		}
		bw.evHandler(fmt.Sprintf("bcWorker: manageBlockchain: add peer nodes: adding node %s", ipPort))
	}

	return nil
}

// writePeerBlocks queries the specified node asking for blocks this
// node does not have.
func (bw *bcWorker) writePeerBlocks(ipPort string) error {
	bw.evHandler("bcWorker: manageBlockchain: writePeerBlocks: started")
	defer bw.evHandler("bcWorker: manageBlockchain: writePeerBlocks: cempleted")

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
		bw.evHandler("bcWorker: manageBlockchain: writePeerBlocks: no new block")
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
		bw.evHandler(fmt.Sprintf("bcWorker: manageBlockchain: writePeerBlocks: prevBlk[%s]: newBlk[%s]: numTrans[%d]", pb.Header.PrevBlock, pb.Header.ThisBlock, len(pb.Transactions)))

		_, err := bw.node.writeNewBlockFromPeer(pb)
		if err != nil {
			return err
		}
	}

	return nil
}

// mineNewBlock takes all the transactions from the mempool and writes a
// new block to the database.
func (bw *bcWorker) mineNewBlock() {
	bw.evHandler("bcWorker: manageBlockchain: mineNewBlock: started")
	defer bw.evHandler("bcWorker: manageBlockchain: mineNewBlock: cempleted")

	block, err := bw.node.writeNewBlockFromTransactions()
	if err != nil {
		if errors.Is(err, ErrNoTransactions) {
			bw.evHandler("bcWorker: manageBlockchain: mineNewBlock: no transactions in mempool")
			return
		}
		bw.evHandler(fmt.Sprintf("bcWorker: manageBlockchain: mineNewBlock: ERROR %s", err))
		return
	}

	bw.evHandler(fmt.Sprintf("bcWorker: manageBlockchain: mineNewBlock: prevBlk[%x]: newBlk[%x]: numTrans[%d]", block.Header.PrevBlock, block.Hash(), len(block.Transactions)))
}
