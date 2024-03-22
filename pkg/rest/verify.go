package rest

import (
	"mime/multipart"
)

// VerifyReq is the request struct for the verify http endpoint
type VerifyReq struct {
	OwnerEmail string `form:"ownerEmail" json:"ownerEmail" binding:"required,email"`
	DocBcTkn   string `form:"docBcTkn" json:"docBcTkn" binding:"required"`

	// below items not sent via client
	MpFileHeader   *multipart.FileHeader
	OwnerEmailHash string
	DocHash        string
}

// VerifyResp is the response struct for the verify http endpoint
type VerifyResp struct {
	Verified bool   `json:"verified"`
	Error    string `json:"error,omitempty"`
}
