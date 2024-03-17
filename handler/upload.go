package handler

import (
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/vposham/trustdoc/internal/db/sqlc/dbtx"
	"github.com/vposham/trustdoc/pkg/rest"
)

func (d *DocH) Upload(c *gin.Context) {

	// get the request
	req, err := uploadReq(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "validation failed - " + err.Error()})
		return
	}

	// store the file in blob store
	docId, err := d.Blob.Put(c, *req.File, req.MpFileHeader.Size)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "unable to store in blob store - " + err.Error()})
		return
	}

	// do something with blockchain here.

	// store the metadata in db
	err = d.Db.SaveDocMeta(c, dbtx.DocMeta{
		DocId:          docId,
		OwnerEmail:     req.OwnerEmail,
		DocTitle:       req.DocTitle,
		DocDesc:        req.DocDesc,
		DocName:        req.MpFileHeader.Filename,
		OwnerFirstName: req.OwnerFirstName,
		OwnerLastName:  req.OwnerLastName,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "unable to persist to db" + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "file uploaded successfully"})
}

func uploadReq(c *gin.Context) (*rest.UploadReq, error) {
	var req rest.UploadReq

	err := c.Bind(&req)
	if err != nil {
		return &req, fmt.Errorf("unable to parse req - %w", err)
	}

	// Get the file from the request
	req.MpFileHeader, err = c.FormFile("doc")
	if err != nil {
		return &req, fmt.Errorf("unable to read file - %w", err)
	}

	f, err := req.MpFileHeader.Open()
	if err != nil {
		return &req, fmt.Errorf("unable to open file - %w", err)
	}
	req.File = &f
	return &req, nil
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
