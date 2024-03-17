// Package base will have the common handler api's which are needed for running in kubernetes
package base

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// GetHealth returns status ok
func GetHealth(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}
