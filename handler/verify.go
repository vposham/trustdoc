package handler

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/vposham/trustdoc/log"
	"github.com/vposham/trustdoc/pkg/rest"
)

func (d *DocH) Verify(c *gin.Context) {
	logger := log.GetLogger(c)
	logger.Info("document verification request received")

	// parse the request
	req, err := d.verifyReq(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, verifyResp(false, err))
		return
	}
	err = d.Bc.VerifyDocTkn(c, req.DocBcTkn, req.DocMd5Hash, req.OwnerEmailMd5Hash)
	if err != nil {
		c.JSON(http.StatusInternalServerError,
			verifyResp(false, fmt.Errorf("unable to verify in blockchain - %w", err)))
		return
	}

	c.JSON(http.StatusOK, verifyResp(true, nil))
}

func (d *DocH) verifyReq(c *gin.Context) (*rest.VerifyReq, error) {
	var req rest.VerifyReq

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

func verifyResp(verified bool, err error) *rest.VerifyResp {
	if err != nil {
		return &rest.VerifyResp{Error: err.Error(), Verified: verified}
	}
	return &rest.VerifyResp{Verified: verified}
}
