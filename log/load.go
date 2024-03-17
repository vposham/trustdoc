// Package log will have the implementations of initializing and loading of logger used as part of app startup
package log

import (
	"context"
	"sync"

	"go.uber.org/zap"

	"github.com/vposham/trustdoc/config"
)

var (
	// onceInit guarantees initialization of custom zap Logger only once
	onceInit      = new(sync.Once)
	concreteImpls = make(map[string]any)
)

const (
	configuredZapLogKey = "configuredZapLogger"
)

const (
	ctxtLogNotFound   = "logger not found in context."
	logConfDefZapProd = "defaulting to zap-prod configuration."
	ctxNil            = "input context is nil."
)

// Load implementations
func Load(ctx context.Context) error {
	var appErr error
	onceInit.Do(func() {
		appErr = loadImpls(ctx)
	})
	return appErr
}

// loadImpls calls function in od-go/v2/jlog.DefaultConfig to initialize zapBaseLogger
//
//	zapConf is modified depending on environment and properties by calling jlog.CustomizeLogger
//
// Multiple times modification of the zapConf and multi initialization of zapBaseLogger
// is avoided by the protection provided by onceInit
func loadImpls(_ context.Context) error {
	if concreteImpls[configuredZapLogKey] == nil {
		props := config.GetAll()
		zapAppConf := ZapAppConf{
			ShowCallerInLogs:       props.MustGetBool("log.show.caller"),
			ShowStackTraceInLogs:   props.MustGetBool("log.show.stacktrace"),
			UseUnstructuredLogging: props.MustGetBool("log.use.unstructured.logging"),
			AppLogLevel:            props.MustGetString("log.level"),
		}

		// get base default zap configuration
		baseZapConf := DefaultConfig()
		concreteImpls[configuredZapLogKey] = zapAppConf.CustomizeLogger(baseZapConf)
	}
	return nil
}

// GetConfiguredLogger getting zap logger
func GetConfiguredLogger() *zap.Logger {
	v := concreteImpls[configuredZapLogKey]
	return v.(*zap.Logger)
}
