// Package public maintains the group of handlers for public access.
package public

import (
	"context"
	"fmt"
	"net/http"

	"github.com/ardanlabs/blockchain/foundation/node"
	"github.com/ardanlabs/blockchain/foundation/web"
	"go.uber.org/zap"
)

// Handlers manages the set of bar ledger endpoints.
type Handlers struct {
	Log  *zap.SugaredLogger
	Node *node.Node
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

	txs := make([]node.Tx, len(userTxs))
	for i, tx := range userTxs {
		h.Log.Infow("add user tran", "traceid", v.TraceID, "tx", tx)
		txs[i] = node.NewTx(tx.From, tx.To, tx.Value, tx.Data)
	}

	// Add these transaction and share them with the network.
	h.Node.AddTransactions(txs, true)

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

	dbBals := h.Node.QueryBalances(node.Account(acct))
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

// BlocksByAccount returns all the blocks and their details.
func (h Handlers) BlocksByAccount(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	acct := web.Param(r, "acct")

	dbBlocks := h.Node.QueryBlocksByAccount(node.Account(acct))
	if len(dbBlocks) == 0 {
		return web.Respond(ctx, w, nil, http.StatusNoContent)
	}

	return web.Respond(ctx, w, dbBlocks, http.StatusOK)
}

// SignalMining signals to start a mining operation.
func (h Handlers) SignalMining(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	h.Node.SignalMining()

	resp := struct {
		Status string `json:"status"`
	}{
		Status: "mining signalled",
	}

	return web.Respond(ctx, w, resp, http.StatusOK)
}
