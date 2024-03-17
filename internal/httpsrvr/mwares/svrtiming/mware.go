package svrtiming

import (
	"context"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// MwareOpts are options for the Middleware.
type MwareOpts struct {
	// don’t write headers in the response. Metrics are still gathered though.
	DisableHeaders bool
}

// Middleware is a middleware that adds the Server-Timing header to the response.
func Middleware(opts *MwareOpts) gin.HandlerFunc {
	return func(c *gin.Context) {
		rhw := &respHdrWriter{ResponseWriter: c.Writer, opts: opts, h: &Header{}}

		// This places the *Header value into the request context. This
		// can be extracted again with FromContext.
		c.Request = c.Request.WithContext(NewContext(c.Request.Context(), rhw.h))

		c.Writer = rhw

		// process the request
		c.Next()

		c.Writer.WriteHeaderNow()
	}
}

// respHdrWriter is used to record data written into response writer for rewriting response headers
type respHdrWriter struct {
	gin.ResponseWriter
	h    *Header
	opts *MwareOpts
}

// WriteHeader overrides gin's response header writer
func (w *respHdrWriter) WriteHeader(statusCode int) {
	wrtHdr(w)
	w.ResponseWriter.WriteHeader(statusCode)
}

// WriteHeaderNow overrides gin's response header Now writer
func (w *respHdrWriter) WriteHeaderNow() {
	wrtHdr(w)
}

func wrtHdr(w *respHdrWriter) {
	// Grab the lock just in case there is any ongoing concurrency that
	// still has a reference and may be modifying the value.
	w.h.Lock()
	defer w.h.Unlock()

	// If there are no metrics set, or if the user opted-out writing headers,
	// do nothing
	if (w.opts != nil && w.opts.DisableHeaders) || len(w.h.Metrics) == 0 {
		return
	}

	w.Header().Add(hdrKey, w.h.String())
}

// loggerFn is a helper function type which is used to retrieve the logger stored in context
type loggerFn func(ctx context.Context) *zap.Logger
