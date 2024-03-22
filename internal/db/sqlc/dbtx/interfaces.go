package dbtx

import (
	"context"
	"database/sql"

	"github.com/vposham/trustdoc/internal/db/sqlc/raw"
)

// DBConn allows the db connection to be abstracted
type DBConn interface {
	BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error)
	Close() error
}

// Queries is an interface for raw.queries which can be overridden for unit testing
type Queries interface {
	WithTx(tx *sql.Tx) *raw.Queries
	getTxInterface(tx *sql.Tx) Queries

	raw.Querier
}

// StoreIf interface provides all the valid business DB transactions
type StoreIf interface {
	// SaveDocMeta stores the document meta in the db
	SaveDocMeta(ctx context.Context, in DocMeta) error
	// GetDocMetaByHash gets the document meta from the db
	GetDocMetaByHash(ctx context.Context, docHash string) (DocMeta, error)
}
