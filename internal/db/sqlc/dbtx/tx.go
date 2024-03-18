// Package dbtx consists of the DB transactions implementation which includes the executeTrx and executeTrxWithRetry
package dbtx

import (
	"context"
	"crypto/rand"
	"database/sql"
	"errors"
	"fmt"
	"math/big"
	"time"

	"github.com/lib/pq"
	"github.com/sqlc-dev/pqtype"
	"go.uber.org/zap"

	"github.com/vposham/trustdoc/config"
	"github.com/vposham/trustdoc/internal/db/sqlc/raw"
	"github.com/vposham/trustdoc/log"
)

//revive:disable:function-length,cognitive-complexity,cyclomatic

const (
	dbRetryCount    = 5
	dbRetrySleepDur = time.Millisecond * 30

	dbTxLatencyLKey  = "db-tx-latency"
	dbTxAttemptLKey  = "db-attempt"
	dbRetrySleepLKey = "db-sleep"
)

var errDbTimeout = errors.New("db timeout")

// Store struct provides all the valid business DB transactions. It implements StoreIf
type Store struct {
	Queries
	db      DBConn
	timeout time.Duration
}

// QueryBase is a wrapper for raw.queries that contains extra methods that allow mocking
type QueryBase struct {
	*raw.Queries
}

// wraps raw.queries so that it can be mocked in transaction logic tests
func (qb *QueryBase) getTxInterface(tx *sql.Tx) Queries {
	newQuery := new(QueryBase)
	newQuery.Queries = qb.Queries.WithTx(tx)
	return newQuery
}

// NewStore created a new store
func NewStore(db *sql.DB) *Store {
	props := config.GetAll()
	return &Store{
		db:      db,
		Queries: &QueryBase{raw.New(db)},
		timeout: props.MustGetParsedDuration("postgres.db.source.timeout.dur"),
	}
}

// execTx executes a function within a database transactions
// return error in case the execution takes more time than timeout duration specified
// reference - https://stackoverflow.com/questions/52799280/context-confusion-regarding-cancellation
func (store *Store) execTx(ctx context.Context, fn func(queries Queries) error) error {

	logger := log.GetLogger(ctx)
	start := time.Now()
	dbRespErrCh := make(chan error)

	dbCtx, cancel := context.WithTimeout(ctx, store.timeout)

	// this will be used in case where context is not cancelled and db responds with an err
	defer cancel()

	// run db tx in separate go routine
	go func(x context.Context) {
		tx, err := store.db.BeginTx(x, &sql.TxOptions{
			Isolation: sql.LevelSerializable,
		})
		if err != nil {
			dbRespErrCh <- err
			return
		}
		q := store.getTxInterface(tx)
		err = fn(q)
		if err != nil {
			if rbErr := tx.Rollback(); rbErr != nil {
				dbRespErrCh <- fmt.Errorf("tx err: %v, rb err: %v", err, rbErr)
				return
			}
			dbRespErrCh <- err
			return
		}
		err = tx.Commit()
		if err == nil {
			cancel() // upon success, cancel context explicitly
			return
		}
		dbRespErrCh <- err
	}(dbCtx)

	select {
	// upon completion of context
	case <-dbCtx.Done():
		switch dbCtx.Err() {

		// timeout case
		case context.DeadlineExceeded:
			logger.Error("db connectivity timed out",
				zap.Int64(dbTxLatencyLKey, time.Since(start).Milliseconds()),
				zap.Duration("dBTimeout", store.timeout))
			return errDbTimeout

		// success case
		case context.Canceled:
			logger.Info("db tx completed",
				zap.Int64(dbTxLatencyLKey, time.Since(start).Milliseconds()))
			return nil
		}

	// db responded err and context is not cancelled
	case dbTxErr := <-dbRespErrCh:
		if !errors.Is(dbTxErr, sql.ErrNoRows) {
			logger.Error("db tx failed",
				zap.Int64(dbTxLatencyLKey, time.Since(start).Milliseconds()),
				zap.Error(dbTxErr))
		}
		return dbTxErr
	}

	// should never come here
	logger.Warn("db tx timeout logic not handled correctly")
	return errors.New("bug: db tx timeout logic shouldnt come here")
}

// execTxWithRetry executes execTx for dbRetryCount times by introducing a jitter of
// random(1,7) * dbRetrySleepDur duration
func (store *Store) execTxWithRetry(ctx context.Context, fn func(Queries) error) error {
	logger := log.GetLogger(ctx)
	var err error
	for i := 0; i < dbRetryCount; i++ {
		if i > 0 {
			sleepDur := time.Duration(randomInt(1, 6)) * dbRetrySleepDur
			logger.Info("sleeping before retry",
				zap.Duration(dbRetrySleepLKey, sleepDur),
				zap.Int(dbTxAttemptLKey, i))
			time.Sleep(sleepDur)
		}
		err = store.execTx(ctx, fn)
		if err == nil {
			return nil
		}
		// we retry in case of transaction rollback
		// as we are using pessimistic locking, we can retry.
		// if this logic needs to be changed, pls also change Isolation level in execTx
		var validPqErr *pq.Error
		if errors.As(err, &validPqErr) && validPqErr.Code.Class() == "40" {
			logger.Warn("found PQ err , pq error code - %s \n",
				zap.Int("attempt", i),
				zap.String("severity", validPqErr.Severity),
				zap.String("pq-err-code", validPqErr.Code.Name()),
				zap.String("message", validPqErr.Message),
				zap.String("detail", validPqErr.Detail),
				zap.String("hint", validPqErr.Hint),
				zap.String("position", validPqErr.Position),
				zap.String("internalPosition", validPqErr.InternalPosition),
				zap.String("internalQuery", validPqErr.InternalQuery),
				zap.String("where", validPqErr.Where),
				zap.String("schema", validPqErr.Schema),
				zap.String("table", validPqErr.Table),
				zap.String("column", validPqErr.Column),
				zap.String("dataTypeName", validPqErr.DataTypeName),
				zap.String("constraint", validPqErr.Constraint),
				zap.String("file", validPqErr.File),
				zap.String("line", validPqErr.Line),
				zap.String("routine", validPqErr.Routine))
			continue
		}
		// failed with a non retry-able failure
		return err
	}
	return fmt.Errorf("failed after %d attempts, last error: %s", dbRetryCount, err)
}

// NewNullStr returns NullString with correct String and Valid fields
func NewNullStr(s *string) sql.NullString {
	if s == nil || len(*s) == 0 {
		return sql.NullString{}
	}
	return sql.NullString{
		String: *s,
		Valid:  true,
	}
}

func newNullBool(s *bool) sql.NullBool {
	if s == nil {
		return sql.NullBool{}
	}
	return sql.NullBool{
		Bool:  *s,
		Valid: true,
	}
}

func newNullInt64(i *int64) sql.NullInt64 {
	if i == nil || *i == 0 {
		return sql.NullInt64{}
	}
	return sql.NullInt64{
		Int64: *i,
		Valid: true,
	}
}

// NewNullJson returns NullRawMessage with correct RawMessage and Valid fields
func NewNullJson(j *[]byte) pqtype.NullRawMessage {
	if j == nil || len(*j) == 0 {
		return pqtype.NullRawMessage{}
	}

	return pqtype.NullRawMessage{
		RawMessage: *j,
		Valid:      true,
	}
}

// randomInt generates a random int within given limits
func randomInt(min, max int) (result int) {
	nBig, err := rand.Int(rand.Reader, big.NewInt(int64(max-min)))
	if err != nil {
		return min
	}
	return int(nBig.Int64()) + min
}
