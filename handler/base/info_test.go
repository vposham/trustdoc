package base

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestGetInfo(t *testing.T) {
	router := gin.New()
	gin.SetMode(gin.TestMode)
	router.GET("/info", GetInfo)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/info", nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}
