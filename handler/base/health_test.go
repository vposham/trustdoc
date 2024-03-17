package base

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestGetHealth(t *testing.T) {
	router := gin.New()
	gin.SetMode(gin.TestMode)
	router.GET("/health", GetHealth)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "{\"status\":\"ok\"}", w.Body.String())
}
