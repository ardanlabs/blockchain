// Package public maintains the group of handlers for public access.
package public

import (
	"context"
	"fmt"
	"net/http"

	"github.com/ardanlabs/blockchain/foundation/blockchain"
	"github.com/ardanlabs/blockchain/foundation/web"
	"go.uber.org/zap"
)

// Handlers manages the set of bar ledger endpoints.
type Handlers struct {
	Log *zap.SugaredLogger
	BC  *blockchain.State
}

// AddTransactions adds new user transactions to the mempool.
func (h Handlers) AddTransactions(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	v, err := web.GetValues(ctx)
	if err != nil {
		return web.NewShutdownError("web value missing from context")
	}

	var userTxs []tx
	if err := web.Decode(r, &userTxs); err != nil {
		return fmt.Errorf("unable to decode payload: %w", err)
	}

	txs := make([]blockchain.Tx, len(userTxs))
	for i, tx := range userTxs {
		h.Log.Infow("add user tran", "traceid", v.TraceID, "tx", tx)
		txs[i] = h.BC.NewTx(tx.From, tx.To, tx.Value, tx.Tip, tx.Data)
	}

	// Add these transaction and share them with the network.
	h.BC.AddTransactions(txs, true)

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
	account := web.Param(r, "account")
	var balanceSheet blockchain.BalanceSheet

	if account == "" {
		balanceSheet = h.BC.CopyBalanceSheet()
	} else {
		balanceSheet = h.BC.QueryBalances(account)
	}

	bals := make([]balance, 0, len(balanceSheet))
	for act, dbBal := range balanceSheet {
		bal := balance{
			Account: act,
			Balance: dbBal,
		}
		bals = append(bals, bal)
	}

	balances := balances{
		LastestBlock: h.BC.CopyLatestBlock().Hash(),
		Uncommitted:  len(h.BC.CopyMempool()),
		Balances:     bals,
	}

	return web.Respond(ctx, w, balances, http.StatusOK)
}

// BlocksByAccount returns all the blocks and their details.
func (h Handlers) BlocksByAccount(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	account := web.Param(r, "account")

	dbBlocks := h.BC.QueryBlocksByAccount(account)
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
