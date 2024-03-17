package bc

type KaleidoEth struct {
}

var _ OpsIf = (*KaleidoEth)(nil)

func (k KaleidoEth) Sign(docId, docMd5Hash, ownerEmailMd5Hash string) (tknId string, err error) {
	// TODO implement me
	panic("implement me")
}
