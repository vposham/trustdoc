package log

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vposham/trustdoc/config"
	"github.com/vposham/trustdoc/internal/httpsrvr/mwares/reqlogger"
	"go.uber.org/zap"
)

func TestLogConf(t *testing.T) {

	t.Run("Successful load of log", func(t *testing.T) {
		t.Setenv("appEnv", "test")

		ctx := context.Background()
		assert.NoError(t, config.Load(ctx, "./../resources"))
		err := Load(ctx)
		assert.NoError(t, err, "err not expected")

		l := GetLogger(nil)
		assert.NotEqual(t, zap.NewNop(), l)
	})

	t.Run("GetLoggerWithNotFoundInCtxt", func(t *testing.T) {
		t.Setenv("appEnv", "test")

		ctx := context.Background()
		err := Load(ctx)
		assert.NoError(t, err, "err not expected")
		assert.Equal(t, GetConfiguredLogger(), GetLogger(ctx))
	})

}

func TestGetLoggerWithCustomTypeCtxt(t *testing.T) {
	newRootCtx := context.WithValue(context.Background(),
		correlationLoggerKey, zap.NewNop())
	l := GetLogger(newRootCtx)
	assert.Equal(t, zap.NewNop(), l)
}

func TestGetLoggerWithStringCtxt(t *testing.T) {
	newRootCtx := context.WithValue(context.Background(),
		reqlogger.CorrelationLoggerKeyStr, zap.NewNop())
	l := GetLogger(newRootCtx)
	assert.Equal(t, zap.NewNop(), l)
}
