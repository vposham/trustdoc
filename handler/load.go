package handler

import (
	"context"
	"sync"

	"github.com/vposham/trustdoc/internal/bc"
	"github.com/vposham/trustdoc/internal/blob"
	"github.com/vposham/trustdoc/internal/db/sqlc/dbtx"
	"github.com/vposham/trustdoc/internal/hash"
)

var (
	// onceInit guarantees initialization of properties only once
	onceInit      = new(sync.Once)
	concreteImpls = make(map[string]any)
)

const (
	docHandlerImplKey = "docHandlerImpl"
)

// Load is an exported method that loads DocH depending on environment
// Load enables us switch between mocks and real implementation using configuration
func Load(ctx context.Context) error {
	var appErr error
	onceInit.Do(func() {
		appErr = loadImpls(ctx)
	})
	return appErr
}

func loadImpls(ctx context.Context) error {
	if concreteImpls[docHandlerImplKey] == nil {

		// load db layer
		if err := dbtx.Load(ctx); err != nil {
			return err
		}

		// load blob layer
		if err := blob.Load(ctx); err != nil {
			return err
		}

		// load blockchain layer
		if err := bc.Load(ctx); err != nil {
			return err
		}

		concreteImpls[docHandlerImplKey] = &DocH{
			Db:   dbtx.GetDbStore(),
			Blob: blob.GetBlobStore(),
			H:    hash.Sha256{},
			Bc:   bc.GetBc(),
		}
	}
	return nil
}

// GetHandler gets the doc handler
func GetHandler() *DocH {
	return concreteImpls[docHandlerImplKey].(*DocH)
}
