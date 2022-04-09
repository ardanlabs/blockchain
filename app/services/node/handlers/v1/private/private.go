// Package private maintains the group of handlers for node to node access.
package private

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	v1 "github.com/ardanlabs/blockchain/business/web/v1"
	"github.com/ardanlabs/blockchain/foundation/blockchain/peer"
	"github.com/ardanlabs/blockchain/foundation/blockchain/state"
	"github.com/ardanlabs/blockchain/foundation/blockchain/storage"
	"github.com/ardanlabs/blockchain/foundation/nameservice"
	"github.com/ardanlabs/blockchain/foundation/web"
	"go.uber.org/zap"
)

// Handlers manages the set of bar ledger endpoints.
type Handlers struct {
	Log   *zap.SugaredLogger
	State *state.State
	NS    *nameservice.NameService
}

// SubmitNodeTransaction adds new node transactions to the mempool.
func (h Handlers) SubmitNodeTransaction(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	v, err := web.GetValues(ctx)
	if err != nil {
		return web.NewShutdownError("web value missing from context")
	}

	var tx storage.BlockTx
	if err := web.Decode(r, &tx); err != nil {
		return fmt.Errorf("unable to decode payload: %w", err)
	}

	h.Log.Infow("add user tran", "traceid", v.TraceID, "from:nonce", tx, "to", tx.To, "value", tx.Value, "tip", tx.Tip)
	if err := h.State.UpsertNodeTransaction(tx); err != nil {
		return v1.NewRequestError(err, http.StatusBadRequest)
	}

	resp := struct {
		Status string `json:"status"`
	}{
		Status: "transactions added to mempool",
	}

	return web.Respond(ctx, w, resp, http.StatusOK)
}

// MinePeerBlock accepts a new mined block from a peer, validates it, then adds it
// to the block chain.
func (h Handlers) MinePeerBlock(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	var block storage.Block
	if err := web.Decode(r, &block); err != nil {
		return fmt.Errorf("unable to decode payload: %w", err)
	}

	if err := h.State.MinePeerBlock(block); err != nil {

		// More has to be thought about here. I don't think the blockchain
		// package can perform this activity because it doesn't understand
		// the application layer. All activity needs to stop after this call
		// to truncate to re-sync the state of the blockchain.
		// So the idea for now is to truncate the state here and force a
		// shutdown/restart of the service.
		if errors.Is(err, storage.ErrChainForked) {
			h.State.Truncate()
			return web.NewShutdownError(err.Error())
		}

		return v1.NewRequestError(err, http.StatusNotAcceptable)
	}

	resp := struct {
		Status string        `json:"status"`
		Block  storage.Block `json:"block"`
	}{
		Status: "accepted",
		Block:  block,
	}

	return web.Respond(ctx, w, resp, http.StatusOK)
}

// Status returns the current status of the node.
func (h Handlers) Status(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	latestBlock := h.State.RetrieveLatestBlock()

	status := peer.PeerStatus{
		LatestBlockHash:   latestBlock.Hash(),
		LatestBlockNumber: latestBlock.Header.Number,
		KnownPeers:        h.State.RetrieveKnownPeers(),
	}

	return web.Respond(ctx, w, status, http.StatusOK)
}

// BlocksByNumber returns all the blocks based on the specified to/from values.
func (h Handlers) BlocksByNumber(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	fromStr := web.Param(r, "from")
	if fromStr == "latest" || fromStr == "" {
		fromStr = fmt.Sprintf("%d", state.QueryLastest)
	}

	toStr := web.Param(r, "to")
	if toStr == "latest" || toStr == "" {
		toStr = fmt.Sprintf("%d", state.QueryLastest)
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

	blocks := h.State.QueryBlocksByNumber(from, to)
	if len(blocks) == 0 {
		return web.Respond(ctx, w, nil, http.StatusNoContent)
	}

	blocksFS := make([]storage.BlockFS, len(blocks))
	for i, block := range blocks {
		blocksFS[i] = storage.NewBlockFS(block)
	}

	return web.Respond(ctx, w, blocksFS, http.StatusOK)
}

// Mempool returns the set of uncommitted transactions.
func (h Handlers) Mempool(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	txs := h.State.RetrieveMempool()
	return web.Respond(ctx, w, txs, http.StatusOK)
}
