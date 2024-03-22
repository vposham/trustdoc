package rest

// DownloadReq to download a document for the download http endpoint
type DownloadReq struct {
	DocId string `uri:"docId" binding:"required,uuid"`
}
