// Package v1 contains the full set of handler functions and routes
// supported by the v1 web api.
package v1

import (
	"github.com/ardanlabs/blockchain/business/sys/database"
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
}
