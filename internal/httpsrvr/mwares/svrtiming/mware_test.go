package svrtiming

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func someHandler(c *gin.Context) {
	h := FromContext(c)
	m := h.NewMetric("cosmos").Start()
	time.Sleep(50 * time.Millisecond)
	m.Stop()
	c.Status(http.StatusOK)
}

func TestMiddlewareWriteToHdr(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, engine := gin.CreateTestContext(w)
	engine.Use(Middleware(&MwareOpts{DisableHeaders: false}))
	engine.GET("/", someHandler)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req = req.WithContext(c)
	c.Request = req

	engine.ServeHTTP(w, c.Request)

	assert.Contains(t, w.Header().Get(hdrKey), "cosmos;dur=",
		"%s response header must be present", hdrKey)
}

func TestMiddlewareNoWriteToHdr(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, engine := gin.CreateTestContext(w)

	engine.Use(Middleware(&MwareOpts{
		DisableHeaders: true,
	}))

	engine.GET("/", someHandler)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req = req.WithContext(c)
	c.Request = req

	engine.ServeHTTP(w, c.Request)

	assert.NotContains(t, w.Header().Get(hdrKey), "cosmos;dur=",
		"%s response header must be present", hdrKey)
}
