package rest

import (
	"mime/multipart"

	"github.com/vposham/trustdoc/internal/db/sqlc/dbtx"
)

// UploadReq is the request struct for the upload http endpoint
type UploadReq struct {
	OwnerEmail     string `form:"ownerEmail" json:"ownerEmail" binding:"required,email"`
	DocTitle       string `form:"docTitle" json:"docTitle" binding:"required,min=3"`
	DocDesc        string `form:"docDesc" json:"docDesc"`
	OwnerFirstName string `form:"ownerFirstName" json:"ownerFirstName" binding:"required,alpha,min=3"`
	OwnerLastName  string `form:"ownerLastName" json:"ownerLastName" binding:"required,alpha,min=3"`

	// below items not sent via client
	MpFileHeader   *multipart.FileHeader
	OwnerEmailHash string
	DocHash        string
}

// UploadResp is the response struct for the upload http endpoint
type UploadResp struct {
	Doc   *dbtx.DocMeta `json:"doc"`
	Error string        `json:"error,omitempty"`
}
