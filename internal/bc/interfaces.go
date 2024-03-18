package bc

import "context"

type OpsIf interface {
	MintDocTkn(ctx context.Context, docId, docMd5Hash, ownerEmailMd5Hash string) (tknId string, err error)
}
