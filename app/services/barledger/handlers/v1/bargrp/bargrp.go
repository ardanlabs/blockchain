// Package bargrp maintains the group of handlers for bar ledger access.
package bargrp

import (
	"context"
	"errors"
	"fmt"
	"net/http"

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

// QueryStatus returns the current status of the node.
func (h Handlers) QueryStatus(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	status := status{
		Hash:   fmt.Sprintf("%x", h.Node.QueryLatestBlock().Hash()),
		Number: h.Node.QueryLatestBlock().Header.Number,
	}

	return web.Respond(ctx, w, status, http.StatusOK)
}

// AddTransaction adds a new transaction to the mempool.
func (h Handlers) AddTransaction(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	v, err := web.GetValues(ctx)
	if err != nil {
		return web.NewShutdownError("web value missing from context")
	}

	var tx newTx
	if err := web.Decode(r, &tx); err != nil {
		return fmt.Errorf("unable to decode payload: %w", err)
	}

	h.Log.Infow("add tran", "traceid", v.TraceID, "data", tx)

	dbTx := node.NewTx(tx.From, tx.To, tx.Value, tx.Data)
	if err := h.Node.AddTransaction(dbTx); err != nil {
		err = fmt.Errorf("transaction failed, %w", err)
		return v1.NewRequestError(err, http.StatusBadRequest)
	}

	resp := struct {
		Status string `json:"status"`
		newTx
	}{
		Status: "transaction added to mempool",
		newTx:  tx,
	}

	return web.Respond(ctx, w, resp, http.StatusOK)
}

// WriteNewBlock writes the existing transactions in the mempool to a block on disk.
func (h Handlers) WriteNewBlock(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	dbBlock, err := h.Node.WriteNewBlock()
	if err != nil {
		switch {
		case errors.Is(err, node.ErrNoTransactions):
			return v1.NewRequestError(err, http.StatusBadRequest)
		default:
			err = fmt.Errorf("create block failed, %w", err)
			return v1.NewRequestError(err, http.StatusBadRequest)
		}
	}

	resp := struct {
		Status      string `json:"status"`
		NumberTrans int    `json:"num_trans"`
		Block       block  `json:"block"`
	}{
		Status:      "new block created",
		NumberTrans: len(dbBlock.Transactions),
		Block:       toBlock(dbBlock),
	}

	return web.Respond(ctx, w, resp, http.StatusOK)
}

// QueryGenesis returns the genesis information.
func (h Handlers) QueryGenesis(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	gen := h.Node.QueryGenesis()
	return web.Respond(ctx, w, gen, http.StatusOK)
}

// QueryMempool returns the set of uncommitted transactions.
func (h Handlers) QueryMempool(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	txs := h.Node.QueryMempool()
	return web.Respond(ctx, w, txs, http.StatusOK)
}

// QueryBalances returns the current balances for all users.
func (h Handlers) QueryBalances(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
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
		LastestBlock: fmt.Sprintf("%x", h.Node.QueryLatestBlock()),
		Uncommitted:  len(h.Node.QueryMempool()),
		Balances:     bals,
	}

	return web.Respond(ctx, w, balances, http.StatusOK)
}

// QueryBlocks returns all the blocks and their details.
func (h Handlers) QueryBlocks(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	acct := web.Param(r, "acct")
	dbBlocks := h.Node.QueryBlocks(acct)

	out := make([]block, len(dbBlocks))
	for i := range dbBlocks {
		out[i] = toBlock(dbBlocks[i])
	}

	return web.Respond(ctx, w, out, http.StatusOK)
}
