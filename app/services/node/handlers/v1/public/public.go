// Package public maintains the group of handlers for public access.
package public

import (
	"context"
	"fmt"
	"net/http"
	"time"

	v1 "github.com/ardanlabs/blockchain/business/web/v1"
	"github.com/ardanlabs/blockchain/foundation/blockchain/database"
	"github.com/ardanlabs/blockchain/foundation/blockchain/merkle"
	"github.com/ardanlabs/blockchain/foundation/blockchain/state"
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

	// Need this to handle CORS on the websocket.
	h.WS.CheckOrigin = func(r *http.Request) bool { return true }

	// This upgrades the HTTP connection to a websocket connection.
	c, err := h.WS.Upgrade(w, r, nil)
	if err != nil {
		return err
	}
	defer c.Close()

	// This provides a channel for receiving events from the blockchain.
	ch := h.Evts.Acquire(v.TraceID)
	defer h.Evts.Release(v.TraceID)

	// Starting a ticker to send a ping message over the websocket.
	ticker := time.NewTicker(time.Second)

	// Block waiting for events from the blockchain or ticker.
	for {
		select {
		case msg, wd := <-ch:

			// If the channel is closed, release the websocket.
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

// SubmitWalletTransaction adds new transactions to the mempool.
func (h Handlers) SubmitWalletTransaction(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	v, err := web.GetValues(ctx)
	if err != nil {
		return web.NewShutdownError("web value missing from context")
	}

	// Decode the JSON in the post call into a Signed transaction.
	var signedTx database.SignedTx
	if err := web.Decode(r, &signedTx); err != nil {
		return fmt.Errorf("unable to decode payload: %w", err)
	}

	h.Log.Infow("add tran", "traceid", v.TraceID, "from:nonce", signedTx, "to", signedTx.ToID, "value", signedTx.Value, "tip", signedTx.Tip)

	// Ask the state package to add this transaction to the mempool. Only the
	// checks are the transaction signature and the recipient account format.
	// It's up to the wallet to make sure the account has a proper balance and
	// nonce. Fees will be taken if this transaction is mined into a block.
	if err := h.State.UpsertWalletTransaction(signedTx); err != nil {
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

	trans := []tx{}
	for _, tran := range mempool {
		account, _ := tran.FromAccount()
		if acct != "" && ((acct != string(account)) && (acct != string(tran.ToID))) {
			continue
		}

		trans = append(trans, tx{
			FromAccount: account,
			FromName:    h.NS.Lookup(account),
			To:          tran.ToID,
			ToName:      h.NS.Lookup(tran.ToID),
			ChainID:     tran.ChainID,
			Nonce:       tran.Nonce,
			Value:       tran.Value,
			Tip:         tran.Tip,
			Data:        tran.Data,
			TimeStamp:   tran.TimeStamp,
			GasPrice:    tran.GasPrice,
			GasUnits:    tran.GasUnits,
			Sig:         tran.SignatureString(),
		})
	}

	return web.Respond(ctx, w, trans, http.StatusOK)
}

// Accounts returns the current balances for all users.
func (h Handlers) Accounts(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	accountStr := web.Param(r, "account")

	var accounts map[database.AccountID]database.Account
	switch accountStr {
	case "":
		accounts = h.State.RetrieveAccounts()

	default:
		accountID, err := database.ToAccountID(accountStr)
		if err != nil {
			return err
		}
		account, err := h.State.QueryAccounts(accountID)
		if err != nil {
			return err
		}
		accounts = map[database.AccountID]database.Account{accountID: account}
	}

	resp := make([]act, 0, len(accounts))
	for account, info := range accounts {
		act := act{
			Account: account,
			Name:    h.NS.Lookup(account),
			Balance: info.Balance,
			Nonce:   info.Nonce,
		}
		resp = append(resp, act)
	}

	ai := actInfo{
		LastestBlock: h.State.RetrieveLatestBlock().Hash(),
		Uncommitted:  len(h.State.RetrieveMempool()),
		Accounts:     resp,
	}

	return web.Respond(ctx, w, ai, http.StatusOK)
}

// BlocksByAccount returns all the blocks and their details.
func (h Handlers) BlocksByAccount(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	accountStr, err := database.ToAccountID(web.Param(r, "account"))
	if err != nil {
		return err
	}

	dbBlocks, err := h.State.QueryBlocksByAccount(accountStr)
	if err != nil {
		return err
	}
	if len(dbBlocks) == 0 {
		return web.Respond(ctx, w, nil, http.StatusNoContent)
	}

	blocks := make([]block, len(dbBlocks))
	for j, blk := range dbBlocks {
		values := blk.Trans.Values()

		trans := make([]tx, len(blk.Trans.Values()))
		for i, tran := range values {
			account, err := tran.FromAccount()
			if err != nil {
				return err
			}

			rawProof, order, err := blk.Trans.MerkleProof(tran)
			if err != nil {
				return err
			}
			proof := make([]string, len(rawProof))
			for i, rp := range rawProof {
				proof[i] = merkle.ToHex(rp)
			}

			trans[i] = tx{
				FromAccount: account,
				FromName:    h.NS.Lookup(account),
				To:          tran.ToID,
				ToName:      h.NS.Lookup(tran.ToID),
				ChainID:     tran.ChainID,
				Nonce:       tran.Nonce,
				Value:       tran.Value,
				Tip:         tran.Tip,
				Data:        tran.Data,
				TimeStamp:   tran.TimeStamp,
				GasPrice:    tran.GasPrice,
				GasUnits:    tran.GasUnits,
				Sig:         tran.SignatureString(),
				Proof:       proof,
				ProofOrder:  order,
			}
		}

		b := block{
			Number:        blk.Header.Number,
			PrevBlockHash: blk.Header.PrevBlockHash,
			TimeStamp:     blk.Header.TimeStamp,
			BeneficiaryID: blk.Header.BeneficiaryID,
			Difficulty:    blk.Header.Difficulty,
			MiningReward:  blk.Header.MiningReward,
			Nonce:         blk.Header.Nonce,
			TransRoot:     blk.Header.TransRoot,
			Transactions:  trans,
		}

		blocks[j] = b
	}

	return web.Respond(ctx, w, blocks, http.StatusOK)
}
