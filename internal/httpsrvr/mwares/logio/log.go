package logio

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

const (
	headersKey           = "reqHeaders"
	urlPathKey           = "path"
	statusCodeKey        = "statusCode"
	respBodyKey          = "respBody"
	bytesWrittenCountKey = "bytesWritten"
	reqBodyKey           = "reqBody"
	latencyMillisKey     = "latencyMillis"
	clientIPKey          = "clientIP"
	methodKey            = "method"
	userAgentKey         = "userAgent"
	authUserKey          = "authUser"
)

// HttpTx logs the metric of a http request, one per request.
func (l LogImpl) HttpTx(fn getLoggerFn) gin.HandlerFunc {
	return func(c *gin.Context) {
		startTime := time.Now()
		logger := fn(c)

		// blw is only used when logRespBody is true
		blw := &bodyLogWriter{body: &bytes.Buffer{}, ResponseWriter: c.Writer}

		logFields := make([]zap.Field, 0, 10)

		if l.LogHeaders {
			logFields = append(logFields, getHeaders(c, l.HeaderBlackList))
		}

		if l.LogReqBody {
			logFields = append(logFields, getReqBody(c))
		}

		if l.LogRespBody {
			c.Writer = blw
		}

		// process the request
		c.Next()

		if l.LogRespBody {
			logFields = append(logFields, getRespJson(blw.body))
		}

		latencyTime := time.Since(startTime)
		statusCode := c.Writer.Status()
		path := c.Request.URL.Path
		bytesWritten := c.Writer.Size()
		clientIP := c.ClientIP()
		method := c.Request.Method
		userAgent := c.Request.UserAgent()
		authUser, _ := c.Get(gin.AuthUserKey)
		authUserStr := fmt.Sprint(authUser)

		logFields = append(logFields, zap.String(urlPathKey, path), zap.String(methodKey, method),
			zap.Int(statusCodeKey, statusCode), zap.Int64(latencyMillisKey, latencyTime.Milliseconds()),
			zap.Int(bytesWrittenCountKey, bytesWritten), zap.String(clientIPKey, clientIP),
			zap.String(userAgentKey, userAgent), zap.String(authUserKey, authUserStr))

		logger.Info("request/response cycle completed", logFields...)
	}
}

// bodyLogWriter is used to record data written into response writer
type bodyLogWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

// Write overrides gin's response writer to also write in our recorder
func (w bodyLogWriter) Write(b []byte) (n int, err error) {
	// err is always nil for Write method - please refer bytes module,
	// Buffer type's Write method documentation.
	n, _ = w.body.Write(b)
	n, err = w.ResponseWriter.Write(b)
	return
}

// getReqBody - if req body is logged, we have to restore the request body in request
func getReqBody(c *gin.Context) zap.Field {
	// restore the request in request body
	blw := &bodyLogWriter{body: &bytes.Buffer{}, ResponseWriter: c.Writer}
	reqBody, _ := io.ReadAll(c.Request.Body)
	c.Request.Body = io.NopCloser(bytes.NewBuffer(reqBody))
	c.Writer = blw
	return zap.Any(reqBodyKey, jsonBytesToMap(reqBody))
}

func getHeaders(c *gin.Context, blackList []string) zap.Field {
	cReq := c.Request.Clone(c)
	for hk := range cReq.Header {
		for _, bhk := range blackList {
			if strings.EqualFold(hk, bhk) {
				cReq.Header.Del(hk)
				continue
			}
		}
	}
	reqH, _ := json.Marshal(cReq.Header)
	return zap.Any(headersKey, string(reqH))
}

// getRespJson provides the map built from response even though the message is non json
func getRespJson(body *bytes.Buffer) zap.Field {
	var respObj map[string]any
	if body != nil {
		respObj = jsonBytesToMap(body.Bytes())
	}
	return zap.Any(respBodyKey, respObj)
}

func jsonBytesToMap(b []byte) map[string]any {
	m := make(map[string]any)
	if !json.Valid(b) {
		m["processingErr"] = "unable to parse response body as json"
		m["contents"] = string(b)
	}
	_ = json.Unmarshal(b, &m)
	return m
}
