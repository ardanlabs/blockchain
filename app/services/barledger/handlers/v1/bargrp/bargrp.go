// Package bargrp maintains the group of handlers for bar ledger access.
package bargrp

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/ardanlabs/blockchain/business/sys/database"
	v1 "github.com/ardanlabs/blockchain/business/web/v1"
	"github.com/ardanlabs/blockchain/foundation/web"
	"go.uber.org/zap"
)

// Handlers manages the set of bar ledger endpoints.
type Handlers struct {
	Log *zap.SugaredLogger
	DB  *database.DB
}

// AddTransaction adds a new transaction to the database.
func (h Handlers) AddTransaction(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	from := web.Param(r, "from")
	to := web.Param(r, "to")
	valueStr := web.Param(r, "value")
	data := web.Param(r, "data")

	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return v1.NewRequestError(err, http.StatusBadRequest)
	}

	tx := database.NewTx(from, to, uint(value), data)
	if err := h.DB.Add(tx); err != nil {
		return err
	}

	return web.Respond(ctx, w, tx, http.StatusOK)
}

// Persist writes the existing transactions in the mempool to a block on disk.
func (h Handlers) Persist(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	if err := h.DB.Persist(); err != nil {
		return v1.NewRequestError(err, http.StatusInternalServerError)
	}

	resp := struct {
		Status string
	}{
		Status: "mempool persisted",
	}

	return web.Respond(ctx, w, resp, http.StatusOK)
}

// QueryBalances returns the current balances for all users.
func (h Handlers) QueryBalances(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	acct := web.Param(r, "acct")

	var bals []struct {
		Account string
		Balance uint
	}

	for act, bal := range h.DB.Balances(acct) {
		bal := struct {
			Account string
			Balance uint
		}{
			Account: act,
			Balance: bal,
		}
		bals = append(bals, bal)
	}

	return web.Respond(ctx, w, bals, http.StatusOK)
}

// QueryBlocks returns all the blocks and their details.
func (h Handlers) QueryBlocks(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	acct := web.Param(r, "acct")
	blocks := h.DB.Blocks(acct)

	var out []block
	for _, orgBlock := range blocks {
		newBlock := block{
			Header: header{
				PrevBlock: fmt.Sprintf("%x", orgBlock.Header.PrevBlock),
				Time:      orgBlock.Header.Time,
			},
			Transactions: orgBlock.Transactions,
		}
		out = append(out, newBlock)
	}

	return web.Respond(ctx, w, out, http.StatusOK)
}
