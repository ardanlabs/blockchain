// Package public maintains the group of handlers for public access.
package public

import (
	"context"
	"fmt"
	"net/http"

	v1 "github.com/ardanlabs/blockchain/business/web/v1"
	"github.com/ardanlabs/blockchain/foundation/blockchain/accounts"
	"github.com/ardanlabs/blockchain/foundation/blockchain/state"
	"github.com/ardanlabs/blockchain/foundation/blockchain/storage"
	"github.com/ardanlabs/blockchain/foundation/nameservice"
	"github.com/ardanlabs/blockchain/foundation/web"
	"go.uber.org/zap"
)

// Handlers manages the set of bar ledger endpoints.
type Handlers struct {
	Log   *zap.SugaredLogger
	State *state.State
	NS    *nameservice.NameService
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

	h.Log.Infow("add user tran", "traceid", v.TraceID, "tx", signedTx)
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
	txs := h.State.RetrieveMempool()
	return web.Respond(ctx, w, txs, http.StatusOK)
}

// Accounts returns the current balances for all users.
func (h Handlers) Accounts(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	address := web.Param(r, "address")

	var blkActs map[storage.Address]accounts.Info
	switch address {
	case "":
		blkActs = h.State.RetrieveAccounts()

	default:
		addr, err := storage.ToAddress(address)
		if err != nil {
			return err
		}
		blkActs = h.State.QueryAccounts(addr)
	}

	acts := make([]info, 0, len(blkActs))
	for addr, blkInfo := range blkActs {
		act := info{
			Address: addr,
			Name:    h.NS.Lookup(addr),
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

// BlocksByAddress returns all the blocks and their details.
func (h Handlers) BlocksByAddress(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	address, err := storage.ToAddress(web.Param(r, "address"))
	if err != nil {
		return err
	}

	dbBlocks := h.State.QueryBlocksByAddress(address)
	if len(dbBlocks) == 0 {
		return web.Respond(ctx, w, nil, http.StatusNoContent)
	}

	blocks := make([]block, len(dbBlocks))
	for j, blk := range dbBlocks {
		trans := make([]tx, len(blk.Transactions))
		for i, tran := range blk.Transactions {
			address, _ := tran.FromAddress()
			trans[i] = tx{
				FromAddress: address,
				FromName:    h.NS.Lookup(address),
				Nonce:       tran.Nonce,
				To:          tran.To,
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
			MinerAddress: blk.Header.MinerAddress,
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
