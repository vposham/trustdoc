package blob

import (
	"context"
	"io"
)

type OpsIf interface {
	Put(ctx context.Context, doc io.Reader, size int64) (docId string, err error)
	Get(ctx context.Context, docId string) (doc io.Reader, err error)
}
