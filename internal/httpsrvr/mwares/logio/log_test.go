package logio

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
	yaml "gopkg.in/yaml.v3"
)

// o global variable is used to validate captured logs
var o *observer.ObservedLogs

func testLogger(_ context.Context) *zap.Logger {
	p, _ := rand.Prime(rand.Reader, 64)
	lCore, observedLogs := observer.New(zapcore.InfoLevel)
	observedLogger := zap.New(lCore)
	o = observedLogs
	return observedLogger.With(zap.String("traceId", p.String()))
}

func TestLogImpl_HttpTx_Json(t *testing.T) {
	l := LogImpl{
		HeaderBlackList: []string{"authorization"},
		LogReqBody:      true,
		LogRespBody:     true,
		LogHeaders:      true,
	}
	w := httptest.NewRecorder()
	ctx, engine := gin.CreateTestContext(w)

	engine.Use(l.HttpTx(testLogger))

	engine.POST("/test/json", jsonHandler)

	req := httptest.NewRequest(http.MethodPost, "/test/json",
		bytes.NewBuffer([]byte(`{"message":"ping"}`)))
	req.Header.Set("Authorization", "Basic dXNlcutyudGVzdDE=")
	req.Header.Set("content-type", "application/json; charset=UTF-8")

	ctx.Request = req
	engine.ServeHTTP(w, req)

	if http.StatusOK != w.Code {
		t.Fail()
	}

	if len(o.All()) != 1 {
		t.Fail()
	}

	logEntry := o.All()[0]
	if logEntry.Message != "request/response cycle completed" {
		t.Fail()
	}

	reqBody, ok := logEntry.ContextMap()[reqBodyKey]
	if !ok {
		t.FailNow()
	}

	reqBodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		t.FailNow()
	}

	if string(reqBodyBytes) != `{"message":"ping"}` {
		t.Fail()
	}

	respBody, ok := logEntry.ContextMap()[respBodyKey]
	if !ok {
		t.FailNow()
	}

	respBodyBytes, err := json.Marshal(respBody)
	if err != nil {
		t.FailNow()
	}

	if string(respBodyBytes) != `{"message":"pong"}` {
		t.Fail()
	}
}

func TestLogImpl_HttpTx_NonJson(t *testing.T) {
	l := LogImpl{
		HeaderBlackList: []string{"authorization"},
		LogReqBody:      true,
		LogRespBody:     true,
		LogHeaders:      true,
	}
	w := httptest.NewRecorder()
	ctx, engine := gin.CreateTestContext(w)

	engine.Use(l.HttpTx(testLogger))
	engine.POST("/test/other", otherHandler)

	req := httptest.NewRequest(http.MethodPost, "/test/other",
		bytes.NewBuffer([]byte(`ping`)))
	req.Header.Set("Authorization", "Basic dXNlcutyudGVzdDE=")
	req.Header.Set("content-type", "application/yaml")

	ctx.Request = req
	engine.ServeHTTP(w, req)

	if http.StatusOK != w.Code {
		t.Fail()
	}

	if len(o.All()) != 1 {
		t.Fail()
	}

	logEntry := o.All()[0]
	if logEntry.Message != "request/response cycle completed" {
		t.Fail()
	}

	reqBody, ok := logEntry.ContextMap()[reqBodyKey]
	if !ok {
		t.FailNow()
	}

	reqBodyBytes, err := yaml.Marshal(reqBody)
	if err != nil {
		t.FailNow()
	}

	reqBodyBytesStr := string(reqBodyBytes)

	if !strings.Contains(reqBodyBytesStr, `contents: ping`) {
		t.Fail()
	}

	if !strings.Contains(reqBodyBytesStr, `processingErr: unable to parse response body as json`) {
		t.Fail()
	}

	respBody, ok := logEntry.ContextMap()[respBodyKey]
	if !ok {
		t.FailNow()
	}

	respBodyBytes, err := yaml.Marshal(respBody)
	if err != nil {
		t.FailNow()
	}

	respBodyBytesStr := string(respBodyBytes)

	if !strings.Contains(respBodyBytesStr, `message: pong`) {
		t.Fail()
	}

	if !strings.Contains(respBodyBytesStr, `processingErr: unable to parse response body as json`) {
		t.Fail()
	}
}

func jsonHandler(c *gin.Context) {
	reqBody := struct {
		Message string `json:"message"`
	}{}
	err := c.ShouldBindJSON(&reqBody)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}
	if reqBody.Message == "ping" {
		c.JSON(http.StatusOK, gin.H{"message": "pong"})
		return
	}
	c.Status(http.StatusInternalServerError)
}

func otherHandler(c *gin.Context) {
	bis, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}
	bisStr := string(bis)
	if bisStr == "ping" {
		c.YAML(http.StatusOK, gin.H{"message": "pong"})
		return
	}
	c.Status(http.StatusInternalServerError)
}
