package handler

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (d *DocH) Verify(c *gin.Context) {
	docId := c.Param("docId")

	fmt.Println(docId)
	// do something with blockchain here.

	c.JSON(http.StatusOK, gin.H{"message": "file verified successfully"})
}
