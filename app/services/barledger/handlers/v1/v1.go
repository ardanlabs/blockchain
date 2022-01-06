// Package v1 contains the full set of handler functions and routes
// supported by the v1 web api.
package v1

import (
	"net/http"

	"github.com/ardanlabs/blockchain/app/services/barledger/handlers/v1/bargrp"
	"github.com/ardanlabs/blockchain/foundation/database"
	"github.com/ardanlabs/blockchain/foundation/web"
	"go.uber.org/zap"
)

// Config contains all the mandatory systems required by handlers.
type Config struct {
	Log *zap.SugaredLogger
	DB  *database.DB
}

// Routes binds all the version 1 routes.
func Routes(app *web.App, cfg Config) {
	const version = "v1"

	// Register user management and authentication endpoints.
	bgh := bargrp.Handlers{
		Log: cfg.Log,
		DB:  cfg.DB,
	}

	app.Handle(http.MethodGet, version, "/genesis/list", bgh.QueryGenesis)
	app.Handle(http.MethodGet, version, "/balances/list", bgh.QueryBalances)
	app.Handle(http.MethodGet, version, "/balances/list/:acct", bgh.QueryBalances)
	app.Handle(http.MethodGet, version, "/blocks/list", bgh.QueryBlocks)
	app.Handle(http.MethodGet, version, "/tx/uncommitted/list", bgh.QueryUncommitted)
	app.Handle(http.MethodPost, version, "/tx/add", bgh.AddTransaction)
	app.Handle(http.MethodPut, version, "/tx/persist", bgh.Persist)
}
