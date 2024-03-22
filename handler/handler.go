// Package handler provides the endpoints implementation for this service's APIs.
package handler

import (
	"github.com/vposham/trustdoc/internal/bc"
	"github.com/vposham/trustdoc/internal/blob"
	"github.com/vposham/trustdoc/internal/db/sqlc/dbtx"
	"github.com/vposham/trustdoc/internal/hash"
)

// DocH will have all the dependencies this handler will have
type DocH struct {
	Db   dbtx.StoreIf
	Blob blob.OpsIf
	Bc   bc.OpsIf
	H    hash.Sha256
}
