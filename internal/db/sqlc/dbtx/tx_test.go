package dbtx

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"reflect"
	"testing"
	"time"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/lib/pq"
	"github.com/sqlc-dev/pqtype"
	"github.com/stretchr/testify/assert"

	"github.com/vposham/trustdoc/internal/db/sqlc/raw"
)

//revive:disable

// TestNewNullInt64 tests if newNullInt64 returns correct sql.NullInt64
func Test_newNullInt64(t *testing.T) {
	var validVal int64 = 123
	type args struct {
		i *int64
	}
	tests := []struct {
		name string
		args args
		want sql.NullInt64
	}{
		{
			name: "nil int64",
			args: args{
				i: nil,
			},
			want: sql.NullInt64{},
		},
		{
			name: "empty int64",
			args: args{
				i: new(int64),
			},
			want: sql.NullInt64{},
		},
		{
			name: "valid int64",
			args: args{
				i: &validVal,
			},
			want: sql.NullInt64{Valid: true, Int64: validVal},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := newNullInt64(tt.args.i); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("newNullInt64() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestNewNullStr tests if NewNullString returns correct sql.NullString
func Test_NewNullStr(t *testing.T) {
	validStr := "abcd"
	type args struct {
		s *string
	}
	tests := []struct {
		name string
		args args
		want sql.NullString
	}{
		{
			name: "Empty",
			args: args{
				s: new(string),
			},
			want: sql.NullString{},
		},
		{
			name: "Empty",
			args: args{
				s: &validStr,
			},
			want: sql.NullString{
				String: "abcd",
				Valid:  true,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewNullStr(tt.args.s); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewNullStr() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestNewNullBool tests if NewNullBool returns correct sql.NullBool
func Test_newNullBool(t *testing.T) {
	validBool := true
	type args struct {
		s *bool
	}
	tests := []struct {
		name string
		args args
		want sql.NullBool
	}{
		{
			name: "Empty",
			args: args{
				s: nil,
			},
			want: sql.NullBool{},
		},
		{
			name: "Valid",
			args: args{
				s: &validBool,
			},
			want: sql.NullBool{
				Bool:  true,
				Valid: true,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := newNullBool(tt.args.s); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("newNullBool() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestNewNullJson tests if NewNullJson returns correct pqtype.NullRawMessage
func TestNewNullJson(t *testing.T) {
	validByte, _ := json.Marshal("abcd")
	type args struct {
		s *[]byte
	}
	tests := []struct {
		name string
		args args
		want pqtype.NullRawMessage
	}{
		{
			name: "empty json",
			args: args{
				s: new([]byte),
			},
			want: pqtype.NullRawMessage{},
		},
		{
			name: "valid json",
			args: args{
				s: &validByte,
			},
			want: pqtype.NullRawMessage{
				RawMessage: validByte,
				Valid:      true,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewNullJson(tt.args.s); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewNullJson() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStore_execTx(t *testing.T) {
	dbTimeOut := 50 * time.Millisecond
	someDbErr := errors.New("some db err")

	fastQuery := func(queries Queries) error {
		time.Sleep(10 * time.Millisecond)
		return nil
	}

	slowQuery := func(queries Queries) error {
		time.Sleep(80 * time.Millisecond)
		return nil
	}

	errQuery := func(queries Queries) error {
		time.Sleep(20 * time.Millisecond)
		return someDbErr
	}

	type fields struct {
		timeout time.Duration
	}
	type args struct {
		ctx context.Context
		fn  func(queries Queries) error
	}
	tests := []struct {
		name      string
		fields    fields
		args      args
		ExpectErr error
	}{
		{
			name: "db success",
			fields: fields{
				timeout: dbTimeOut,
			},
			args: args{
				ctx: context.Background(),
				fn:  fastQuery,
			},
			ExpectErr: nil,
		},
		{
			name: "db responded error",
			fields: fields{
				timeout: dbTimeOut,
			},
			args: args{
				ctx: context.Background(),
				fn:  errQuery,
			},
			ExpectErr: someDbErr,
		},
		{
			name: "db timeout error",
			fields: fields{
				timeout: dbTimeOut,
			},
			args: args{
				ctx: context.Background(),
				fn:  slowQuery,
			},
			ExpectErr: errDbTimeout,
		},
	}
	for _, tt := range tests {
		db, mock, mockErr := sqlmock.New()
		if mockErr != nil {
			t.FailNow()
		}
		t.Run(tt.name, func(t *testing.T) {
			store := &Store{
				Queries: &QueryBase{&raw.Queries{}},
				db:      db,
				timeout: tt.fields.timeout,
			}

			mock.ExpectBegin()

			switch tt.name {
			case "db success":
				mock.ExpectCommit()
			case "db responded error":
				mock.ExpectRollback()
			}
			if err := store.execTx(tt.args.ctx, tt.args.fn); err != tt.ExpectErr {
				t.Errorf("execTx() expected Err = %v, recieved Err %v", tt.ExpectErr, err)
			}

			switch tt.name {
			case "db timeout error":
				return
			default:
				mock.ExpectClose()
			}
			err := db.Close()

			if err != nil {
				t.Errorf("cant close mock db - %v", err)
			}
		})
	}
}

func TestStore_execTxWithRetry(t *testing.T) {
	dbTimeOut := 500 * time.Millisecond
	someDbErr := errors.New("some db err")

	maxRetriesErr := errors.New("failed after 5 attempts, last error: pq: ")

	someErrQuery := func(queries Queries) error {
		time.Sleep(100 * time.Millisecond)
		return someDbErr
	}
	retryableErrQuery := func(queries Queries) error {
		time.Sleep(100 * time.Millisecond)
		return &pq.Error{
			Code: "40000",
		}
	}
	goodQuery := func(queries Queries) error {
		time.Sleep(1 * time.Millisecond)
		return nil
	}
	type args struct {
		ctx context.Context
		fn  func(Queries) error
	}

	tests := []struct {
		name     string
		args     args
		expected error
	}{
		{
			name: "failed with non retry able err in tx",
			args: args{
				ctx: context.Background(),
				fn:  someErrQuery,
			},
			expected: someDbErr,
		},
		{
			name: "failed with retry able err in tx",
			args: args{
				ctx: context.Background(),
				fn:  retryableErrQuery,
			},
			expected: maxRetriesErr,
		},
		{
			name: "good tx",
			args: args{
				ctx: context.Background(),
				fn:  goodQuery,
			},
			expected: nil,
		},
	}
	for _, tt := range tests {
		db, mock, mockErr := sqlmock.New()
		if mockErr != nil {
			t.FailNow()
		}

		switch tt.name {
		case "good tx":
			mock.ExpectBegin()
			mock.ExpectCommit()
		case "failed with retry able err in tx":
			for i := 0; i < 5; i++ {
				mock.ExpectBegin()
				mock.ExpectRollback()
			}
		case "failed with non retry able err in tx":
			mock.ExpectBegin()
			mock.ExpectRollback()
		}

		t.Run(tt.name, func(t *testing.T) {
			store := &Store{
				Queries: &QueryBase{&raw.Queries{}},
				db:      db,
				timeout: dbTimeOut,
			}
			if err := store.execTxWithRetry(tt.args.ctx, tt.args.fn); err != nil &&
				err.Error() != tt.expected.Error() {
				t.Errorf("execTxWithRetry() expected Err = %v, recieved Err = %v",
					err, tt.expected)
			}
			mock.ExpectClose()
			err := db.Close()

			if err != nil {
				t.Errorf("cant close mock db - %v", err)
			}
		})
	}
}

func Test_randomInt(t *testing.T) {
	type args struct {
		min int
		max int
	}
	tests := []struct {
		name      string
		args      args
		minResult int
		maxResult int
		repeat    int
	}{
		{
			name:      "valid",
			args:      args{min: 1, max: 10},
			minResult: 1,
			maxResult: 10,
			repeat:    100,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for i := 0; i < tt.repeat; i++ {
				val := randomInt(tt.args.min, tt.args.max)
				assert.GreaterOrEqual(t, val, tt.minResult, "randomInt(%v, %v) should be greater than or equal "+
					"to %v", tt.args.min, tt.args.max, tt.args.min)
				assert.LessOrEqual(t, val, tt.maxResult, "randomInt(%v, %v) should be less than or equal "+
					"to %v", tt.args.min, tt.args.max, tt.args.max)
			}
		})
	}
}
