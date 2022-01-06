// Package bargrp maintains the group of handlers for bar ledger access.
package bargrp

import (
	"context"
	"fmt"
	"net/http"

	v1 "github.com/ardanlabs/blockchain/business/web/v1"
	"github.com/ardanlabs/blockchain/foundation/database"
	"github.com/ardanlabs/blockchain/foundation/web"
	"go.uber.org/zap"
)

// Handlers manages the set of bar ledger endpoints.
type Handlers struct {
	Log *zap.SugaredLogger
	DB  *database.DB
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

	dbTx := database.NewTx(tx.From, tx.To, tx.Value, tx.Data)
	if err := h.DB.AddTransaction(dbTx); err != nil {
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

// CreateBlock writes the existing transactions in the mempool to a block on disk.
func (h Handlers) CreateBlock(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	dbBlock, err := h.DB.CreateBlock()
	if err != nil {
		err = fmt.Errorf("create block failed, %w", err)
		return v1.NewRequestError(err, http.StatusBadRequest)
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
	gen := h.DB.QueryGenesis()
	return web.Respond(ctx, w, gen, http.StatusOK)
}

// QueryMempool returns the set of uncommitted transactions.
func (h Handlers) QueryMempool(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	txs := h.DB.QueryMempool()
	return web.Respond(ctx, w, txs, http.StatusOK)
}

// QueryBalances returns the current balances for all users.
func (h Handlers) QueryBalances(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	acct := web.Param(r, "acct")

	dbBals := h.DB.QueryBalances(acct)
	bals := make([]balance, 0, len(dbBals))

	for act, dbBal := range dbBals {
		bal := balance{
			Account: act,
			Balance: dbBal,
		}
		bals = append(bals, bal)
	}

	balances := balances{
		LastestBlock: fmt.Sprintf("%x", h.DB.QueryLastestBlock()),
		Uncommitted:  len(h.DB.QueryMempool()),
		Balances:     bals,
	}

	return web.Respond(ctx, w, balances, http.StatusOK)
}

// QueryBlocks returns all the blocks and their details.
func (h Handlers) QueryBlocks(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	acct := web.Param(r, "acct")
	dbBlocks := h.DB.QueryBlocks(acct)

	out := make([]block, len(dbBlocks))
	for i := range dbBlocks {
		out[i] = toBlock(dbBlocks[i])
	}

	return web.Respond(ctx, w, out, http.StatusOK)
}
