// Package public maintains the group of handlers for public access.
package public

import (
	"context"
	"fmt"
	"net/http"

	v1 "github.com/ardanlabs/blockchain/business/web/v1"
	"github.com/ardanlabs/blockchain/foundation/blockchain"
	"github.com/ardanlabs/blockchain/foundation/nameservice"
	"github.com/ardanlabs/blockchain/foundation/web"
	"go.uber.org/zap"
)

// Handlers manages the set of bar ledger endpoints.
type Handlers struct {
	Log *zap.SugaredLogger
	BC  *blockchain.State
	NS  *nameservice.NameService
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

	h.Log.Infow("add user tran", "traceid", v.TraceID, "tx", signedTx.SignatureString()[:16])
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
	var balanceSheet map[string]uint

	if address == "" {
		balanceSheet = h.BC.CopyBalanceSheet()
	} else {
		balanceSheet = h.BC.QueryBalances(address)
	}

	bals := make([]balance, 0, len(balanceSheet))
	for addr, dbBal := range balanceSheet {
		bal := balance{
			Address: addr,
			Name:    h.NS.Lookup(addr),
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

// BlocksByAddress returns all the blocks and their details.
func (h Handlers) BlocksByAddress(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	address := web.Param(r, "address")

	dbBlocks := h.BC.QueryBlocksByAddress(address)
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
				To:          tran.To,
				Value:       tran.Value,
				Tip:         tran.Tip,
				Data:        tran.Data,
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
