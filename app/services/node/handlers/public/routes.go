package public

import (
	"net/http"

	"github.com/ardanlabs/blockchain/foundation/blockchain/state"
	"github.com/ardanlabs/blockchain/foundation/events"
	"github.com/ardanlabs/blockchain/foundation/nameservice"
	"github.com/ardanlabs/blockchain/foundation/web"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

// Config contains all the mandatory systems required by handlers.
type Config struct {
	Log   *zap.SugaredLogger
	State *state.State
	NS    *nameservice.NameService
	Evts  *events.Events
}

// Routes binds all the public routes.
func Routes(app *web.App, cfg Config) {
	pbl := Handlers{
		Log:   cfg.Log,
		State: cfg.State,
		NS:    cfg.NS,
		WS:    websocket.Upgrader{},
		Evts:  cfg.Evts,
	}

	const version = "v1"

	app.Handle(http.MethodGet, version, "/events", pbl.Events)
	app.Handle(http.MethodGet, version, "/genesis/list", pbl.Genesis)
	app.Handle(http.MethodGet, version, "/accounts/list", pbl.Accounts)
	app.Handle(http.MethodGet, version, "/accounts/list/:account", pbl.Accounts)
	app.Handle(http.MethodGet, version, "/blocks/list", pbl.BlocksByAccount)
	app.Handle(http.MethodGet, version, "/blocks/list/:account", pbl.BlocksByAccount)
	app.Handle(http.MethodGet, version, "/tx/uncommitted/list", pbl.Mempool)
	app.Handle(http.MethodGet, version, "/tx/uncommitted/list/:account", pbl.Mempool)
	app.Handle(http.MethodPost, version, "/tx/submit", pbl.SubmitWalletTransaction)
	app.Handle(http.MethodPost, version, "/tx/proof/:block/", pbl.SubmitWalletTransaction)
}
