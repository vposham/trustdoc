// Package main is the starting point the application
package main

import (
	"context"

	"github.com/vposham/trustdoc/config"
	"github.com/vposham/trustdoc/handler"
	"github.com/vposham/trustdoc/internal/httpsrvr"
	"github.com/vposham/trustdoc/log"
	"go.uber.org/zap"
)

func main() {

	// create a background context for a long-running application
	ctx := context.Background()

	// create a default zap sl for logging app startup operations.
	sl, _ := zap.NewProduction()
	sl = sl.With(zap.String("action", "app startup"))

	// Initially, always load all app properties
	handleStartUpErr(sl, config.Load(ctx, "./config"))

	// load log config
	handleStartUpErr(sl, log.Load(ctx))

	// load handler which exposes all the endpoints
	handleStartUpErr(sl, handler.Load(ctx))

	// load http server
	handleStartUpErr(sl, httpsrvr.Load(ctx))

	port := config.GetAll().MustGetString("app.port")

	// start http server
	httpsrvr.Start(ctx, sl, port)
}

// handleStartUpErr makes sure that app fails to start in case of
// invalid app configurations
func handleStartUpErr(l *zap.Logger, err error) {
	if err != nil {
		l.Panic("failed to start application", zap.NamedError("appCrashed", err))
	}
}
