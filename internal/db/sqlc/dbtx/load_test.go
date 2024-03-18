package dbtx

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/vposham/trustdoc/config"
	"github.com/vposham/trustdoc/log"
)

func TestMain(m *testing.M) {
	ce := os.Getenv("appEnv")
	defer func() {
		_ = os.Setenv("appEnv", ce)
	}()
	_ = os.Setenv("appEnv", "test")
	ctx := context.Background()
	_ = config.Load(ctx, "../../../../config")
	_ = log.Load(ctx)
	_ = Load(ctx)

	os.Exit(m.Run())
}

func TestGetDbStore(t *testing.T) {
	assert.NotNil(t, GetDbStore())
}
