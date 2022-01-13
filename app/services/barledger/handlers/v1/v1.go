// Package v1 contains the full set of handler functions and routes
// supported by the v1 web api.
package v1

import (
	"net/http"

	"github.com/ardanlabs/blockchain/app/services/barledger/handlers/v1/bargrp"
	"github.com/ardanlabs/blockchain/foundation/node"
	"github.com/ardanlabs/blockchain/foundation/web"
	"go.uber.org/zap"
)

// Config contains all the mandatory systems required by handlers.
type Config struct {
	Log  *zap.SugaredLogger
	Node *node.Node
}

// Routes binds all the version 1 routes.
func Routes(app *web.App, cfg Config) {
	const version = "v1"

	// Register user management and authentication endpoints.
	bgh := bargrp.Handlers{
		Log:  cfg.Log,
		Node: cfg.Node,
	}

	app.Handle(http.MethodGet, version, "/node/status", bgh.Status)

	app.Handle(http.MethodGet, version, "/genesis/list", bgh.Genesis)

	app.Handle(http.MethodGet, version, "/balances/list", bgh.Balances)
	app.Handle(http.MethodGet, version, "/balances/list/:acct", bgh.Balances)

	app.Handle(http.MethodGet, version, "/blocks/list", bgh.BlocksByAccount)
	app.Handle(http.MethodGet, version, "/blocks/list/latest", bgh.BlocksByNumber)
	app.Handle(http.MethodGet, version, "/blocks/list/:acct", bgh.BlocksByAccount)
	app.Handle(http.MethodGet, version, "/blocks/list/:from/:to", bgh.BlocksByNumber)

	app.Handle(http.MethodPost, version, "/tx/add", bgh.AddTransaction)
	app.Handle(http.MethodGet, version, "/tx/uncommitted/list", bgh.Mempool)
}
