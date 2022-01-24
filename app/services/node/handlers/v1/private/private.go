// Package private maintains the group of handlers for node to node access.
package private

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	v1 "github.com/ardanlabs/blockchain/business/web/v1"
	"github.com/ardanlabs/blockchain/foundation/blockchain"
	"github.com/ardanlabs/blockchain/foundation/web"
	"go.uber.org/zap"
)

// Handlers manages the set of bar ledger endpoints.
type Handlers struct {
	Log *zap.SugaredLogger
	BC  *blockchain.State
}

// AddNextBlock accepts a new mined block from a peer, validates it, then adds it
// to the block chain.
func (h Handlers) AddNextBlock(ctx context.Context, w http.ResponseWriter, r *http.Request) error {

	// If the node is mining, it needs to stop immediately.
	h.BC.SignalCancelMining()

	var block blockchain.Block
	if err := web.Decode(r, &block); err != nil {
		return fmt.Errorf("unable to decode payload: %w", err)
	}

	if err := h.BC.WriteNextBlock(block); err != nil {

		// We need to correct the fork in our chain.
		if errors.Is(err, blockchain.ErrChainForked) {
			h.BC.Truncate()
		}
		return v1.NewRequestError(err, http.StatusNotAcceptable)
	}

	resp := struct {
		Status string           `json:"status"`
		Block  blockchain.Block `json:"block"`
	}{
		Status: "accepted",
		Block:  block,
	}

	return web.Respond(ctx, w, resp, http.StatusOK)
}

// AddTransactions adds new node transactions to the mempool.
func (h Handlers) AddTransactions(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	v, err := web.GetValues(ctx)
	if err != nil {
		return web.NewShutdownError("web value missing from context")
	}

	var txs []blockchain.Tx
	if err := web.Decode(r, &txs); err != nil {
		return fmt.Errorf("unable to decode payload: %w", err)
	}

	for _, tx := range txs {
		h.Log.Infow("add node tran", "traceid", v.TraceID, "tx", tx)
	}

	// Add these transaction but don't share them, since they were
	// shared with us already.
	h.BC.AddTransactions(txs, false)

	resp := struct {
		Status string `json:"status"`
		Total  int
	}{
		Status: "added",
		Total:  len(txs),
	}

	return web.Respond(ctx, w, resp, http.StatusOK)
}

// Status returns the current status of the node.
func (h Handlers) Status(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	latestBlock := h.BC.CopyLatestBlock()

	status := blockchain.PeerStatus{
		LatestBlockHash:   latestBlock.Hash(),
		LatestBlockNumber: latestBlock.Header.Number,
		KnownPeers:        h.BC.CopyKnownPeers(),
	}

	return web.Respond(ctx, w, status, http.StatusOK)
}

// BlocksByNumber returns all the blocks based on the specified to/from values.
func (h Handlers) BlocksByNumber(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	fromStr := web.Param(r, "from")
	if fromStr == "latest" || fromStr == "" {
		fromStr = fmt.Sprintf("%d", blockchain.QueryLastest)
	}

	toStr := web.Param(r, "to")
	if toStr == "latest" || toStr == "" {
		toStr = fmt.Sprintf("%d", blockchain.QueryLastest)
	}

	from, err := strconv.ParseUint(fromStr, 10, 64)
	if err != nil {
		return v1.NewRequestError(err, http.StatusBadRequest)
	}
	to, err := strconv.ParseUint(toStr, 10, 64)
	if err != nil {
		return v1.NewRequestError(err, http.StatusBadRequest)
	}

	if from > to {
		return v1.NewRequestError(errors.New("from greater than to"), http.StatusBadRequest)
	}

	dbBlocks := h.BC.QueryBlocksByNumber(from, to)
	if len(dbBlocks) == 0 {
		return web.Respond(ctx, w, nil, http.StatusNoContent)
	}

	return web.Respond(ctx, w, dbBlocks, http.StatusOK)
}
