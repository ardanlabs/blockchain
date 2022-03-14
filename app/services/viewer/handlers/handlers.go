// Package handlers contains the full set of handler functions and routes
// supported by the web api.
package handlers

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/ardanlabs/blockchain/business/web/v1/mid"
	"github.com/ardanlabs/blockchain/foundation/web"
	"go.uber.org/zap"
)

// UIMux constructs an http.Handler with all application routes defined.
func UIMux(build string, shutdown chan os.Signal, log *zap.SugaredLogger) (*web.App, error) {
	app := web.NewApp(
		shutdown,
		mid.Logger(log),
		mid.Errors(log),
		mid.Panics(),
		mid.Cors("*"),
	)

	// Register the index page for the website.
	ig, err := newIndex()
	if err != nil {
		return nil, fmt.Errorf("loading index template: %w", err)
	}
	app.Handle(http.MethodGet, "", "/", ig.handler)

	// Register the assets.
	fs := http.FileServer(http.Dir("app/services/viewer/assets"))
	fs = http.StripPrefix("/assets/", fs)
	f := func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
		fs.ServeHTTP(w, r)
		return nil
	}
	app.Handle(http.MethodGet, "", "/assets/*", f)

	return app, nil
}
