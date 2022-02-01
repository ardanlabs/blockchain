// Package v1 contains the full set of handler functions and routes
// supported by the v1 web api.
package v1

import (
	"net/http"

	"github.com/ardanlabs/blockchain/app/services/node/handlers/v1/private"
	"github.com/ardanlabs/blockchain/app/services/node/handlers/v1/public"
	"github.com/ardanlabs/blockchain/foundation/blockchain"
	"github.com/ardanlabs/blockchain/foundation/web"
	"go.uber.org/zap"
)

const version = "v1"

// Config contains all the mandatory systems required by handlers.
type Config struct {
	Log *zap.SugaredLogger
	BC  *blockchain.State
}

// PublicRoutes binds all the version 1 public routes.
func PublicRoutes(app *web.App, cfg Config) {
	pbl := public.Handlers{
		Log: cfg.Log,
		BC:  cfg.BC,
	}

	app.Handle(http.MethodGet, version, "/genesis/list", pbl.Genesis)
	app.Handle(http.MethodGet, version, "/balances/list", pbl.Balances)
	app.Handle(http.MethodGet, version, "/balances/list/:account", pbl.Balances)
	app.Handle(http.MethodGet, version, "/blocks/list", pbl.BlocksByAccount)
	app.Handle(http.MethodGet, version, "/blocks/list/:account", pbl.BlocksByAccount)
	app.Handle(http.MethodGet, version, "/mining/signal", pbl.SignalMining)
	app.Handle(http.MethodGet, version, "/tx/uncommitted/list", pbl.Mempool)
	app.Handle(http.MethodPost, version, "/tx/add", pbl.AddTransactions)
	app.Handle(http.MethodPost, version, "/tx/send", pbl.SendTransactions)
}

// PrivateRoutes binds all the version 1 private routes.
func PrivateRoutes(app *web.App, cfg Config) {
	prv := private.Handlers{
		Log: cfg.Log,
		BC:  cfg.BC,
	}

	app.Handle(http.MethodGet, version, "/node/status", prv.Status)
	app.Handle(http.MethodGet, version, "/node/block/list/:from/:to", prv.BlocksByNumber)
	app.Handle(http.MethodPost, version, "/node/block/next", prv.AddNextBlock)
	app.Handle(http.MethodPost, version, "/node/tx/add", prv.AddTransactions)
}
