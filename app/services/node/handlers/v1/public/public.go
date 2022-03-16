// Package public maintains the group of handlers for public access.
package public

import (
	"context"
	"fmt"
	"net/http"
	"time"

	v1 "github.com/ardanlabs/blockchain/business/web/v1"
	"github.com/ardanlabs/blockchain/foundation/blockchain/accounts"
	"github.com/ardanlabs/blockchain/foundation/blockchain/state"
	"github.com/ardanlabs/blockchain/foundation/blockchain/storage"
	"github.com/ardanlabs/blockchain/foundation/events"
	"github.com/ardanlabs/blockchain/foundation/nameservice"
	"github.com/ardanlabs/blockchain/foundation/web"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

// Handlers manages the set of bar ledger endpoints.
type Handlers struct {
	Log   *zap.SugaredLogger
	State *state.State
	NS    *nameservice.NameService
	WS    websocket.Upgrader
	Evts  *events.Events
}

// Events handles a web socket to provide events to a client.
func (h Handlers) Events(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	v, err := web.GetValues(ctx)
	if err != nil {
		return web.NewShutdownError("web value missing from context")
	}

	h.WS.CheckOrigin = func(r *http.Request) bool { return true }

	c, err := h.WS.Upgrade(w, r, nil)
	if err != nil {
		return err
	}
	defer c.Close()

	ch := h.Evts.Acquire(v.TraceID)
	defer h.Evts.Release(v.TraceID)

	ticker := time.NewTicker(time.Second)

	for {
		select {
		case msg, wd := <-ch:
			if !wd {
				return nil
			}

			if err := c.WriteMessage(websocket.TextMessage, []byte(msg)); err != nil {
				return err
			}

		case <-ticker.C:
			if err := c.WriteMessage(websocket.PingMessage, []byte("ping")); err != nil {
				return nil
			}
		}
	}
}

// SubmitWalletTransaction adds new user transactions to the mempool.
func (h Handlers) SubmitWalletTransaction(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	v, err := web.GetValues(ctx)
	if err != nil {
		return web.NewShutdownError("web value missing from context")
	}

	var signedTx storage.SignedTx
	if err := web.Decode(r, &signedTx); err != nil {
		return fmt.Errorf("unable to decode payload: %w", err)
	}

	h.Log.Infow("add user tran", "traceid", v.TraceID, "from:nonce", signedTx, "to", signedTx.To, "value", signedTx.Value, "tip", signedTx.Tip)
	if err := h.State.SubmitWalletTransaction(signedTx); err != nil {
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
	gen := h.State.RetrieveGenesis()
	return web.Respond(ctx, w, gen, http.StatusOK)
}

// Mempool returns the set of uncommitted transactions.
func (h Handlers) Mempool(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	acct := web.Param(r, "account")

	mempool := h.State.RetrieveMempool()

	trans := make([]tx, len(mempool))
	for i, tran := range mempool {
		account, _ := tran.FromAccount()

		if acct != "" && ((acct != string(account)) && (acct != string(tran.To))) {
			continue
		}

		trans[i] = tx{
			FromAccount: account,
			FromName:    h.NS.Lookup(account),
			To:          tran.To,
			ToName:      h.NS.Lookup(tran.To),
			Nonce:       tran.Nonce,
			Value:       tran.Value,
			Tip:         tran.Tip,
			Data:        tran.Data,
			TimeStamp:   tran.TimeStamp,
			Gas:         tran.Gas,
			Sig:         tran.SignatureString(),
		}
	}

	return web.Respond(ctx, w, trans, http.StatusOK)
}

// Accounts returns the current balances for all users.
func (h Handlers) Accounts(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	account := web.Param(r, "account")

	var blkAccounts map[storage.Account]accounts.Info
	switch account {
	case "":
		blkAccounts = h.State.RetrieveAccounts()

	default:
		account, err := storage.ToAccount(account)
		if err != nil {
			return err
		}
		blkAccounts = h.State.QueryAccounts(account)
	}

	acts := make([]info, 0, len(blkAccounts))
	for account, blkInfo := range blkAccounts {
		act := info{
			Account: account,
			Name:    h.NS.Lookup(account),
			Balance: blkInfo.Balance,
			Nonce:   blkInfo.Nonce,
		}
		acts = append(acts, act)
	}

	ai := actInfo{
		LastestBlock: h.State.RetrieveLatestBlock().Hash(),
		Uncommitted:  len(h.State.RetrieveMempool()),
		Accounts:     acts,
	}

	return web.Respond(ctx, w, ai, http.StatusOK)
}

// BlocksByAccount returns all the blocks and their details.
func (h Handlers) BlocksByAccount(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	account, err := storage.ToAccount(web.Param(r, "account"))
	if err != nil {
		return err
	}

	dbBlocks := h.State.QueryBlocksByAccount(account)
	if len(dbBlocks) == 0 {
		return web.Respond(ctx, w, nil, http.StatusNoContent)
	}

	blocks := make([]block, len(dbBlocks))
	for j, blk := range dbBlocks {
		trans := make([]tx, len(blk.Transactions))
		for i, tran := range blk.Transactions {
			account, _ := tran.FromAccount()
			trans[i] = tx{
				FromAccount: account,
				FromName:    h.NS.Lookup(account),
				To:          tran.To,
				ToName:      h.NS.Lookup(tran.To),
				Nonce:       tran.Nonce,
				Value:       tran.Value,
				Tip:         tran.Tip,
				Data:        tran.Data,
				TimeStamp:   tran.TimeStamp,
				Gas:         tran.Gas,
				Sig:         tran.SignatureString(),
			}
		}

		b := block{
			ParentHash:   blk.Header.ParentHash,
			MinerAccount: blk.Header.MinerAccount,
			Difficulty:   blk.Header.Difficulty,
			Number:       blk.Header.Number,
			TotalTip:     blk.Header.TotalTip,
			TotalGas:     blk.Header.TotalGas,
			TimeStamp:    blk.Header.TimeStamp,
			Nonce:        blk.Header.Nonce,
			Transactions: trans,
		}

		blocks[j] = b
	}

	return web.Respond(ctx, w, blocks, http.StatusOK)
}
