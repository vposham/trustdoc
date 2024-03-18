package hash

import (
	"context"
	"io"
)

type Hasher interface {
	Hash(ctx context.Context, data io.Reader) (string, error)
}
