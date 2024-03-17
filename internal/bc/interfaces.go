package bc

type OpsIf interface {
	Sign(docId, docMd5Hash, ownerEmailMd5Hash string) (tknId string, err error)
}
