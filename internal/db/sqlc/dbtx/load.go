// Package dbtx will contain the interfaces and implementations of DB transactions, dependency loading functionality etc
package dbtx

import (
	"context"
	"database/sql"
	"sync"

	"github.com/vposham/trustdoc/config"
)

var (
	onceInit      = new(sync.Once)
	concreteImpls = make(map[string]any)
)

const (
	// dbExecKey implementation makes calls to postgres db
	dbExecKey = "PostgresConfiguredExecKey"
)

// Load enables us inject this package as dependency from its parent
func Load(ctx context.Context) error {
	var appErr error
	onceInit.Do(func() {
		appErr = loadImpls(ctx)
	})
	return appErr
}

func loadImpls(_ context.Context) error {
	props := config.GetAll()

	if concreteImpls[dbExecKey] == nil {
		dataSourceUrl := props.MustGetString("postgres.db.url")
		dbDriver := props.MustGetString("postgres.sql.driver")
		conn, err := sql.Open(dbDriver, dataSourceUrl)
		if err != nil {
			return err
		}
		conn.SetMaxIdleConns(props.MustGetInt("postgres.db.max.idle.conns"))
		conn.SetMaxOpenConns(props.MustGetInt("postgres.db.max.open.conns"))
		conn.SetConnMaxLifetime(props.MustGetParsedDuration("postgres.db.conn.max.dur"))
		conn.SetConnMaxIdleTime(props.MustGetParsedDuration("postgres.db.timeout.dur"))
		var dbTxStore StoreIf = NewStore(conn)
		concreteImpls[dbExecKey] = dbTxStore
	}

	return nil
}

// GetDbStore is used to get db store
func GetDbStore() StoreIf {
	targetImpl := dbExecKey
	v := concreteImpls[targetImpl]
	return v.(StoreIf)
}
