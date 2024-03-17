package rest

import "mime/multipart"

type UploadReq struct {
	OwnerEmail     string `form:"ownerEmail" json:"ownerEmail" binding:"required,email"`
	DocTitle       string `form:"docTitle" json:"docTitle" binding:"required,min=3"`
	DocDesc        string `form:"docDesc" json:"docDesc"`
	OwnerFirstName string `form:"ownerFirstName" json:"ownerFirstName" binding:"required,alpha,min=3"`
	OwnerLastName  string `form:"ownerLastName" json:"ownerLastName" binding:"required,alpha,min=3"`

	MpFileHeader *multipart.FileHeader
	File         *multipart.File
}
