package handler

import (
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/vposham/trustdoc/internal/db/sqlc/dbtx"
)

func (d *DocH) Upload(c *gin.Context) {
	// Get the file from the request
	fileHeader, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "unable to find file" + err.Error()})
		return
	}

	f, err := fileHeader.Open()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "unable to read file" + err.Error()})
		return
	}

	docId, err := d.Blob.Put(c, f, fileHeader.Size)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "unable to save file" + err.Error()})
		return
	}

	// do something with blockchain here.

	err = d.Db.SaveDocMeta(c, dbtx.DocMeta{DocId: docId})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "unable to persist to db" + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "File uploaded successfully"})
}

func (d *DocH) Download(c *gin.Context) {
	docId := c.Param("docId")

	doc, err := d.Blob.Get(c, docId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "unable to find file" + err.Error()})
		return
	}

	c.Header("Content-Disposition", "attachment; filename="+docId)
	c.Header("Content-Type", "application/octet-stream")
	_, err = io.Copy(c.Writer, doc)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "unable to find file" + err.Error()})
		return
	}
}
