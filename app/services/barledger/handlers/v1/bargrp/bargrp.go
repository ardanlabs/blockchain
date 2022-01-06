// Package bargrp maintains the group of handlers for bar ledger access.
package bargrp

import (
	"context"
	"fmt"
	"net/http"
	"time"

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
	if err := h.DB.AddMempool(dbTx); err != nil {
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

// Persist writes the existing transactions in the mempool to a block on disk.
func (h Handlers) Persist(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	if err := h.DB.Persist(); err != nil {
		err = fmt.Errorf("persist failed, %w", err)
		return v1.NewRequestError(err, http.StatusBadRequest)
	}

	resp := struct {
		Status string
	}{
		Status: "mempool persisted",
	}

	return web.Respond(ctx, w, resp, http.StatusOK)
}

// QueryGenesis returns the genesis information.
func (h Handlers) QueryGenesis(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	gen := h.DB.Genesis()
	return web.Respond(ctx, w, gen, http.StatusOK)
}

// QueryUncommitted returns the set of uncommitted transactions.
func (h Handlers) QueryUncommitted(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	txs := h.DB.UncommittedTransactions()
	return web.Respond(ctx, w, txs, http.StatusOK)
}

// QueryBalances returns the current balances for all users.
func (h Handlers) QueryBalances(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	acct := web.Param(r, "acct")

	var bals []balance
	for act, dbBal := range h.DB.Balances(acct) {
		bal := balance{
			Account: act,
			Balance: dbBal,
		}
		bals = append(bals, bal)
	}

	balances := balances{
		LastestBlock: fmt.Sprintf("%x", h.DB.LastestBlock()),
		Uncommitted:  len(h.DB.UncommittedTransactions()),
		Balances:     bals,
	}

	return web.Respond(ctx, w, balances, http.StatusOK)
}

// QueryBlocks returns all the blocks and their details.
func (h Handlers) QueryBlocks(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	acct := web.Param(r, "acct")
	blocks := h.DB.Blocks(acct)

	var out []block
	for _, orgBlock := range blocks {
		hash, err := orgBlock.Hash()
		if err != nil {
			return err
		}

		newBlock := block{
			Header: blockHeader{
				PrevBlock: fmt.Sprintf("%x", orgBlock.Header.PrevBlock),
				ThisBlock: fmt.Sprintf("%x", hash),
				Time:      time.Unix(int64(orgBlock.Header.Time), 0),
			},
			Transactions: toTxs(orgBlock.Transactions),
		}
		out = append(out, newBlock)
	}

	return web.Respond(ctx, w, out, http.StatusOK)
}
