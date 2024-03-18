package rest

import (
	"mime/multipart"
)

type VerifyReq struct {
	OwnerEmail string `form:"ownerEmail" json:"ownerEmail" binding:"required,email"`
	DocBcTkn   string `form:"docBcTkn" json:"docBcTkn" binding:"required"`

	// below items not sent via client
	MpFileHeader      *multipart.FileHeader
	OwnerEmailMd5Hash string
	DocMd5Hash        string
}

type VerifyResp struct {
	Verified bool   `json:"verified"`
	Error    string `json:"error,omitempty"`
}
