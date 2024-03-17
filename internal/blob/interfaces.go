package blob

import (
	"context"
	"io"
)

// OpsIf is the interface for blob store operations
type OpsIf interface {
	// Put is used to put a document in blob store
	Put(ctx context.Context, doc io.Reader, size int64) (docId string, err error)

	// Get is used to get a document from blob store
	Get(ctx context.Context, docId string) (doc io.Reader, err error)
}
