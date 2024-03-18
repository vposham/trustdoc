package handler

import (
	"database/sql"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/vposham/trustdoc/internal/db/sqlc/dbtx"
	"github.com/vposham/trustdoc/log"
	"github.com/vposham/trustdoc/pkg/rest"
	"go.uber.org/zap"
)

func (d *DocH) Upload(c *gin.Context) {

	logger := log.GetLogger(c)
	logger.Info("upload request received")

	// parse the request
	req, err := d.uploadReq(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "req validation failed - " + err.Error()})
		return
	}

	var exists bool
	doc, err := d.Db.GetDocMetaByHash(c, req.DocMd5Hash)
	if err == nil {
		exists = true
	} else {
		if !errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "unable to find doc in db - " + err.Error()})
			return
		}
		exists = false
	}
	logger.Info("doc exists check", zap.String("docId", doc.DocId), zap.Bool("docExists", exists))

	if exists {
		c.JSON(http.StatusOK, gin.H{"message": "file already exists" + doc.DocId})
		return
	}

	// store the file in blob store
	f, _ := req.MpFileHeader.Open()
	docId, err := d.Blob.Put(c, f, req.MpFileHeader.Size)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "unable to store in blob store - " + err.Error()})
		return
	}

	// mint a new tkn in blockchain
	bcTknId, err := d.Bc.MintDocTkn(c, docId, req.DocMd5Hash, req.OwnerEmailMd5Hash)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "unable to sign in blockchain - " + err.Error()})
		return
	}

	// store the metadata in db
	err = d.Db.SaveDocMeta(c, dbtx.DocMeta{
		DocId:          docId,
		OwnerEmail:     req.OwnerEmail,
		DocTitle:       req.DocTitle,
		DocDesc:        req.DocDesc,
		DocMd5Hash:     req.DocMd5Hash,
		BcTknId:        bcTknId,
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

func (d *DocH) uploadReq(c *gin.Context) (*rest.UploadReq, error) {
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

	// generate md5 hash for the doc and owner email id
	docMd5Hash, e1 := d.H.Hash(c, f)
	ownerEmailIdMd5Hash, e2 := d.H.Hash(c, strings.NewReader(req.OwnerEmail))
	if e1 != nil || e2 != nil {
		return &req, fmt.Errorf("unable to generate hash. docHashErr - %w. emailHashErr - %w", e1, e2)
	}
	req.DocMd5Hash = docMd5Hash
	req.OwnerEmailMd5Hash = ownerEmailIdMd5Hash
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
