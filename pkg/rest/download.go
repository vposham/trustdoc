package rest

type DownloadReq struct {
	DocId string `uri:"docId" binding:"required,uuid"`
}
