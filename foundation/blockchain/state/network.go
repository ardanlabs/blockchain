package state

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/ardanlabs/blockchain/foundation/blockchain/database"
	"github.com/ardanlabs/blockchain/foundation/blockchain/database/storage"
	"github.com/ardanlabs/blockchain/foundation/blockchain/peer"
)

const baseURL = "http://%s/v1/node"

// NetSendBlockToPeers takes the new mined block and sends it to all know peers.
func (s *State) NetSendBlockToPeers(block database.Block) error {
	s.evHandler("state: NetSendBlockToPeers: started")
	defer s.evHandler("state: NetSendBlockToPeers: completed")

	for _, peer := range s.RetrieveKnownPeers() {
		url := fmt.Sprintf("%s/block/propose", fmt.Sprintf(baseURL, peer.Host))

		var status struct {
			Status string `json:"status"`
		}

		if err := send(http.MethodPost, url, storage.NewBlock(block), &status); err != nil {
			return fmt.Errorf("%s: %s", peer.Host, err)
		}

		s.evHandler("state: NetSendBlockToPeers: sent to peer[%s]", peer)
	}

	return nil
}

// NetSendTxToPeers shares a new block transaction with the known peers.
func (s *State) NetSendTxToPeers(tx database.BlockTx) {
	s.evHandler("state: NetSendTxToPeers: started")
	defer s.evHandler("state: NetSendTxToPeers: completed")

	// CORE NOTE: Bitcoin does not send the full transaction immediately to save
	// on bandwidth. A node will send the transaction's mempool key first so the
	// receiving node can check if they already have the transaction or not. If
	// the receiving node doesn't have it, then it will request the transaction
	// based on the mempool key it received.

	// For now, the Ardan blockchain just sends the full transaction.
	for _, peer := range s.RetrieveKnownPeers() {
		url := fmt.Sprintf("%s/tx/submit", fmt.Sprintf(baseURL, peer.Host))
		if err := send(http.MethodPost, url, tx, nil); err != nil {
			s.evHandler("state: NetSendTxToPeers: WARNING: %s", err)
		}
	}
}

// NetRequestPeerStatus looks for new nodes on the blockchain by asking
// known nodes for their peer list. New nodes are added to the list.
func (s *State) NetRequestPeerStatus(pr peer.Peer) (peer.PeerStatus, error) {
	s.evHandler("state: NetRequestPeerStatus: started: %s", pr)
	defer s.evHandler("state: NetRequestPeerStatus: completed: %s", pr)

	url := fmt.Sprintf("%s/status", fmt.Sprintf(baseURL, pr.Host))

	var ps peer.PeerStatus
	if err := send(http.MethodGet, url, nil, &ps); err != nil {
		return peer.PeerStatus{}, err
	}

	s.evHandler("state: NetRequestPeerStatus: peer-node[%s]: latest-blknum[%d]: peer-list[%s]", pr, ps.LatestBlockNumber, ps.KnownPeers)

	return ps, nil
}

// NetRequestPeerMempool asks the peer for the transactions in their mempool.
func (s *State) NetRequestPeerMempool(pr peer.Peer) ([]database.BlockTx, error) {
	s.evHandler("state: NetRequestPeerMempool: started: %s", pr)
	defer s.evHandler("state: NetRequestPeerMempool: completed: %s", pr)

	url := fmt.Sprintf("%s/tx/list", fmt.Sprintf(baseURL, pr.Host))

	var mempool []database.BlockTx
	if err := send(http.MethodGet, url, nil, &mempool); err != nil {
		return nil, err
	}

	s.evHandler("state: sync: NetRequestPeerMempool: len[%d]", len(mempool))

	return mempool, nil
}

// NetRequestPeerBlocks queries the specified node asking for blocks this node does
// not have, then writes them to disk.
func (s *State) NetRequestPeerBlocks(pr peer.Peer) error {
	s.evHandler("worker: NetRequestPeerBlocks: started: %s", pr)
	defer s.evHandler("worker: NetRequestPeerBlocks: completed: %s", pr)

	// CORE NOTE: Ideally you want to start by pulling just block headers and
	// performing the cryptographic audit so you know your're not being attacked.
	// After that you can start pulling the full block data for each block header
	// if you are a full node and maybe only the last 1000 full blocks if you
	// are a pruned node. That can be done in the background. Remember, you
	// only need block headers to validate new blocks.

	// Currently the Ardan blockchain is a full node only system and needs the
	// transactions to have a complete account database. The cryptographic audit
	// does take place as each full block is downloaded from peers.

	from := s.RetrieveLatestBlock().Header.Number + 1
	url := fmt.Sprintf("%s/block/list/%d/latest", fmt.Sprintf(baseURL, pr.Host), from)

	var blocks []storage.Block
	if err := send(http.MethodGet, url, nil, &blocks); err != nil {
		return err
	}

	s.evHandler("worker: NetRequestPeerBlocks: found blocks[%d]", len(blocks))

	for _, block := range blocks {
		if err := s.ProcessProposedBlock(block); err != nil {
			return err
		}
	}

	return nil
}

// =============================================================================

// send is a helper function to send an HTTP request to a node.
func send(method string, url string, dataSend any, dataRecv any) error {
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
