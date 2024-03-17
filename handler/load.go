package handler

import (
	"context"
	"sync"

	"github.com/vposham/trustdoc/internal/db/sqlc/dbtx"
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

		// load blockchain layer

		concreteImpls[docHandlerImplKey] = &DocH{
			Db: dbtx.GetDbStore(),
		}
	}
	return nil
}

// GetHandler gets the doc handler
func GetHandler() *DocH {
	return concreteImpls[docHandlerImplKey].(*DocH)
}