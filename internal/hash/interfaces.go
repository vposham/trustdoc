// Package hash lets us hash data
package hash

import (
	"context"
	"io"
)

// Hasher is the interface that wraps the basic Hash method.
type Hasher interface {
	Hash(ctx context.Context, data io.Reader) (string, error)
}
