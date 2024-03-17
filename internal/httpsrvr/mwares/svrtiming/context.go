// Package svrtiming is a middleware for gin that adds the Server-Timing header to the response.
// functionality - https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Server-Timing
// ref - https://github.com/mitchellh/go-server-timing
package svrtiming

import (
	"context"

	"github.com/gin-gonic/gin"
)

// NewContext returns a new Context that carries the Header value h.
func NewContext(ctx context.Context, h *Header) context.Context {
	return context.WithValue(ctx, contextKey, h)
}

// FromContext returns the *Header in the context, if any. If no Header
// value exists, nil is returned.
// It checks for Header in gin.Context and then in context.Context
func FromContext(ctx context.Context) *Header {
	ginCtx, ok := ctx.(*gin.Context)
	if ok {
		ctx = ginCtx.Request.Context()
	}
	h, _ := ctx.Value(contextKey).(*Header)
	return h
}

type contextKeyType struct{}

// The key where the header value is stored. This is globally unique since
// it uses a custom unexported type. The struct{} costs zero allocations.
var contextKey = contextKeyType(struct{}{})
