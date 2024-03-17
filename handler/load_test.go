package handler

import (
	"context"
	"errors"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vposham/trustdoc/config"
	"github.com/vposham/trustdoc/internal/db/sqlc/dbtx"

	"github.com/vposham/trustdoc/log"
)

func TestMain(m *testing.M) {
	ce := os.Getenv("appEnv")
	defer func() {
		_ = os.Setenv("appEnv", ce)
	}()
	_ = os.Setenv("appEnv", "test")
	ctx := context.Background()
	_ = config.Load(ctx, "../config")
	_ = log.Load(ctx)
	_ = dbtx.Load(ctx)
	_ = Load(ctx)

	os.Exit(m.Run())
}

func TestDocHandlerLoad(t *testing.T) {
	t.Run("Load no error", func(t *testing.T) {
		got := Load(context.Background())
		if got != nil {
			assert.Equalf(t, nil, got, "Load no error")
		}
	})

	t.Run("GetHandler", func(t *testing.T) {
		got := GetHandler()
		if got == nil {
			assert.Error(t, errors.New("got nil"), "GetHandler()")
		}
	})

}

func Test_loadImpls(t *testing.T) {
	type args struct {
		in0 context.Context
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name:    "success loadImpl",
			args:    args{in0: context.Background()},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := loadImpls(tt.args.in0); (err != nil) != tt.wantErr {
				t.Errorf("loadImpls() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
