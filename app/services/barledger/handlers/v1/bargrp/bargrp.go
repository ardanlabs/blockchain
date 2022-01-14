// Package bargrp maintains the group of handlers for bar ledger access.
package bargrp

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

// AddTransactions adds a new transactions to the mempool.
func (h Handlers) AddTransactions(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	v, err := web.GetValues(ctx)
	if err != nil {
		return web.NewShutdownError("web value missing from context")
	}

	var records []node.TxRecord
	if err := web.Decode(r, &records); err != nil {
		return fmt.Errorf("unable to decode payload: %w", err)
	}

	txs := make([]node.Tx, len(records))
	for i, rec := range records {
		h.Log.Infow("add tran", "traceid", v.TraceID, "tx", rec)
		if rec.Nonce == 0 {
			txs[i] = node.NewTx(rec.Nonce, rec.From, rec.To, rec.Value, rec.Data)
		}
	}

	if err := h.Node.SignalAddTransactions(ctx, txs); err != nil {
		err = fmt.Errorf("transaction failed, %w", err)
		return v1.NewRequestError(err, http.StatusBadRequest)
	}

	resp := struct {
		Status string `json:"status"`
		Total  int
	}{
		Status: "transactions added to mempool",
		Total:  len(txs),
	}

	return web.Respond(ctx, w, resp, http.StatusOK)
}

// Genesis returns the genesis information.
func (h Handlers) Genesis(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	gen := h.Node.CopyGenesis()
	return web.Respond(ctx, w, gen, http.StatusOK)
}

// Mempool returns the set of uncommitted transactions.
func (h Handlers) Mempool(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	txs := h.Node.CopyMempool()
	return web.Respond(ctx, w, txs, http.StatusOK)
}

// Balances returns the current balances for all users.
func (h Handlers) Balances(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	acct := web.Param(r, "acct")

	dbBals := h.Node.QueryBalances(acct)
	bals := make([]balance, 0, len(dbBals))

	for act, dbBal := range dbBals {
		bal := balance{
			Account: act,
			Balance: dbBal,
		}
		bals = append(bals, bal)
	}

	balances := balances{
		LastestBlock: h.Node.CopyLatestBlock().Hash(),
		Uncommitted:  len(h.Node.CopyMempool()),
		Balances:     bals,
	}

	return web.Respond(ctx, w, balances, http.StatusOK)
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

// BlocksAcct returns all the blocks and their details.
func (h Handlers) BlocksByAccount(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	acct := web.Param(r, "acct")

	dbBlocks := h.Node.QueryBlocksByAccount(acct)
	if len(dbBlocks) == 0 {
		return web.Respond(ctx, w, nil, http.StatusNoContent)
	}

	return web.Respond(ctx, w, node.BlocksToPeerBlocks(dbBlocks), http.StatusOK)
}
