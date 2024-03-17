package bc

type OpsIf interface {
	Sign(docId, docHash, ownerName string) (signature string, err error)
}
