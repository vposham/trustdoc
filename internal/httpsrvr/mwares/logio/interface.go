// Package logio logs the request and response lifecycle of a http request
package logio

import (
	"context"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Log is an interface that is implemented to log http request/response
type Log interface {
	HttpTx(getLoggerFn) gin.HandlerFunc
}

// LogImpl is the implementation struct that contains all the dependencies needed for Log
type LogImpl struct {
	HeaderBlackList []string // list of all header names which should not be logged
	LogReqBody      bool
	LogRespBody     bool
	LogHeaders      bool
}

// getLoggerFn is a helper function type which is used to retrieve the logger stored in context
type getLoggerFn func(ctx context.Context) *zap.Logger
