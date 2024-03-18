package handler

import (
	"io"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/vposham/trustdoc/log"
	"github.com/vposham/trustdoc/pkg/rest"
)

func (d *DocH) Download(c *gin.Context) {
	logger := log.GetLogger(c)
	logger.Info("download request received")

	var req rest.DownloadReq
	err := c.BindUri(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "req validation failed - " + err.Error()})
		return
	}

	doc, err := d.Blob.Get(c, req.DocId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "unable to find file" + err.Error()})
		return
	}

	c.Header("Content-Disposition", "attachment; filename="+req.DocId)
	c.Header("Content-Type", c.Request.Header.Get("Content-Type"))
	_, err = io.Copy(c.Writer, doc)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "unable to find file" + err.Error()})
		return
	}
}
