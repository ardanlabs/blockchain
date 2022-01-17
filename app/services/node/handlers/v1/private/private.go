// Package private maintains the group of handlers for node to node access.
package private

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	v1 "github.com/ardanlabs/blockchain/business/web/v1"
	"github.com/ardanlabs/blockchain/foundation/node"
	"github.com/ardanlabs/blockchain/foundation/web"
	"go.uber.org/zap"
)

// Handlers manages the set of bar ledger endpoints.
type Handlers struct {
	Log  *zap.SugaredLogger
	Node *node.Node
}

// AddTransactions adds new node transactions to the mempool.
func (h Handlers) AddTransactions(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	v, err := web.GetValues(ctx)
	if err != nil {
		return web.NewShutdownError("web value missing from context")
	}

	var txs []node.Tx
	if err := web.Decode(r, &txs); err != nil {
		return fmt.Errorf("unable to decode payload: %w", err)
	}

	for _, tx := range txs {
		h.Log.Infow("add node tran", "traceid", v.TraceID, "tx", tx)
	}

	// Add these transaction but don't share them, since they were
	// shared with us already.
	h.Node.AddTransactions(txs, false)

	resp := struct {
		Status string `json:"status"`
		Total  int
	}{
		Status: "transactions added to mempool",
		Total:  len(txs),
	}

	return web.Respond(ctx, w, resp, http.StatusOK)
}

// Status returns the current status of the node.
func (h Handlers) Status(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	latestBlock := h.Node.CopyLatestBlock()

	status := struct {
		Hash              string              `json:"hash"`
		LatestBlockNumber uint64              `json:"latest_block_number"`
		KnownPeers        map[string]struct{} `json:"known_peers"`
	}{
		Hash:              latestBlock.Hash(),
		LatestBlockNumber: latestBlock.Header.Number,
		KnownPeers:        h.Node.CopyKnownPeersList(),
	}

	return web.Respond(ctx, w, status, http.StatusOK)
}

// BlocksByNumber returns all the blocks based on the specified to/from values.
func (h Handlers) BlocksByNumber(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	fromStr := web.Param(r, "from")
	if fromStr == "latest" || fromStr == "" {
		fromStr = fmt.Sprintf("%d", node.QueryLastest)
	}

	toStr := web.Param(r, "to")
	if toStr == "latest" || toStr == "" {
		toStr = fmt.Sprintf("%d", node.QueryLastest)
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

	dbBlocks := h.Node.QueryBlocksByNumber(from, to)
	if len(dbBlocks) == 0 {
		return web.Respond(ctx, w, nil, http.StatusNoContent)
	}

	return web.Respond(ctx, w, node.BlocksToPeerBlocks(dbBlocks), http.StatusOK)
}
