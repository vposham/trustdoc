package httpsrvr

import (
	"context"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/vposham/trustdoc/config"
	"github.com/vposham/trustdoc/handler"
	"github.com/vposham/trustdoc/internal/db/sqlc/dbtx"
	"github.com/vposham/trustdoc/log"
	"go.uber.org/zap"
)

func TestCreateServerInit(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctx := context.Background()
	t.Setenv("appEnv", "test")
	_ = config.Load(ctx, "../../config")
	_ = log.Load(ctx)
	_ = dbtx.Load(ctx)
	_ = handler.Load(ctx)

	assert.NoError(t, Load(ctx))
	go Start(ctx, zap.NewNop(), "9999")
}

func Test_randSeq(t *testing.T) {
	desiredLen := 8
	if got, err := randSeq(desiredLen); err != nil || len(got) != desiredLen {
		t.Errorf("randSeq() = got str of len %d, want str of len %d", len(got), desiredLen)
	}
}

func Test_generateRandomBytes(t *testing.T) {
	_, err := generateRandomBytes(-1)
	assert.Error(t, err)
}
