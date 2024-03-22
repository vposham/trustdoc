package handler

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/vposham/trustdoc/log"
	"github.com/vposham/trustdoc/pkg/rest"
)

// Verify handler takes in a document, owner email and blockchain tokenId and verifies if the document is valid
// it checks the provided doc hash and owner email hash with the one stored in blockchain
func (d *DocH) Verify(c *gin.Context) {
	logger := log.GetLogger(c)
	logger.Info("document verification request received")

	// parse the request
	req, err := d.verifyReq(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, verifyResp(false, err))
		return
	}
	err = d.Bc.VerifyDocTkn(c, req.DocBcTkn, req.DocHash, req.OwnerEmailHash)
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

	// generate  hash for the doc and owner email id
	docHash, e1 := d.H.Hash(c, f)
	ownerEmailIdHash, e2 := d.H.Hash(c, strings.NewReader(req.OwnerEmail))
	if e1 != nil || e2 != nil {
		return &req, fmt.Errorf("unable to generate hash. docHashErr - %w. emailHashErr - %w", e1, e2)
	}
	req.DocHash = docHash
	req.OwnerEmailHash = ownerEmailIdHash
	return &req, nil
}

func verifyResp(verified bool, err error) *rest.VerifyResp {
	if err != nil {
		return &rest.VerifyResp{Error: err.Error(), Verified: verified}
	}
	return &rest.VerifyResp{Verified: verified}
}
