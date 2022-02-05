// Package public maintains the group of handlers for public access.
package public

import (
	"context"
	"fmt"
	"net/http"

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

// SubmitWalletTransaction adds new user transactions to the mempool.
func (h Handlers) SubmitWalletTransaction(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	v, err := web.GetValues(ctx)
	if err != nil {
		return web.NewShutdownError("web value missing from context")
	}

	var signedTx blockchain.SignedTx
	if err := web.Decode(r, &signedTx); err != nil {
		return fmt.Errorf("unable to decode payload: %w", err)
	}

	h.Log.Infow("add user tran", "traceid", v.TraceID, "tx", signedTx)
	if err := h.BC.SubmitWalletTransaction(signedTx); err != nil {
		return v1.NewRequestError(err, http.StatusBadRequest)
	}

	resp := struct {
		Status string `json:"status"`
	}{
		Status: "transactions added to mempool",
	}

	return web.Respond(ctx, w, resp, http.StatusOK)
}

// Genesis returns the genesis information.
func (h Handlers) Genesis(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	gen := h.BC.CopyGenesis()
	return web.Respond(ctx, w, gen, http.StatusOK)
}

// Mempool returns the set of uncommitted transactions.
func (h Handlers) Mempool(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	txs := h.BC.CopyMempool()
	return web.Respond(ctx, w, txs, http.StatusOK)
}

// Balances returns the current balances for all users.
func (h Handlers) Balances(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	address := web.Param(r, "address")
	var balanceSheet blockchain.BalanceSheet

	if address == "" {
		balanceSheet = h.BC.CopyBalanceSheet()
	} else {
		balanceSheet = h.BC.QueryBalances(address)
	}

	bals := make([]balance, 0, len(balanceSheet))
	for addr, dbBal := range balanceSheet {
		bal := balance{
			Address: addr,
			Balance: dbBal,
		}
		bals = append(bals, bal)
	}

	balances := Balances{
		LastestBlock: h.BC.CopyLatestBlock().Hash(),
		Uncommitted:  len(h.BC.CopyMempool()),
		Balances:     bals,
	}

	return web.Respond(ctx, w, balances, http.StatusOK)
}

// BlocksByAddress returns all the blocks and their details.
func (h Handlers) BlocksByAddress(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	address := web.Param(r, "address")

	dbBlocks := h.BC.QueryBlocksByAddress(address)
	if len(dbBlocks) == 0 {
		return web.Respond(ctx, w, nil, http.StatusNoContent)
	}

	return web.Respond(ctx, w, dbBlocks, http.StatusOK)
}

// SignalMining signals to start a mining operation.
func (h Handlers) SignalMining(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	h.BC.SignalMining()

	resp := struct {
		Status string `json:"status"`
	}{
		Status: "mining signalled",
	}

	return web.Respond(ctx, w, resp, http.StatusOK)
}
