package bc

type OpsIf interface {
	Sign(docId, docHash, ownerName string) (signature string, err error)
	Verify(docId, docHash, ownerName string) (valid bool, err error)
}
