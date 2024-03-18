// Package main is the starting point the application
package main

import (
	"context"

	"github.com/vposham/trustdoc/internal/httpsrvr/mwares/reqlogger"
	"go.uber.org/zap"

	"github.com/vposham/trustdoc/config"
	"github.com/vposham/trustdoc/handler"
	"github.com/vposham/trustdoc/internal/httpsrvr"
	"github.com/vposham/trustdoc/log"
)

func main() {

	// create a background context for a long-running application
	ctx := context.Background()

	// create a default zap sl for logging app startup operations.
	sl, _ := zap.NewProduction()
	sl = sl.With(zap.String("action", "app startup"))
	ctx = context.WithValue(ctx, reqlogger.CorrelationLoggerKeyStr, sl)

	// Initially, always load all app properties
	handleStartUpErr(ctx, config.Load(ctx, "./config"))

	// load log config
	handleStartUpErr(ctx, log.Load(ctx))

	// load handler which exposes all the endpoints
	handleStartUpErr(ctx, handler.Load(ctx))

	// load http server
	handleStartUpErr(ctx, httpsrvr.Load(ctx))

	port := config.GetAll().MustGetString("app.port")

	// start http server
	httpsrvr.Start(ctx, sl, port)
}

// handleStartUpErr makes sure that app fails to start in case of
// invalid app configurations
func handleStartUpErr(ctx context.Context, err error) {
	if err != nil {
		logger := log.GetLogger(ctx)
		logger.Panic("failed to start application", zap.NamedError("appCrashed", err))
	}
}
