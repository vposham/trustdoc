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
)

func TestCreateServer(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctx := context.Background()
	t.Setenv("appEnv", "test")
	assert.NoError(t, config.Load(ctx, "../../config"))
	assert.NoError(t, log.Load(ctx))
	assert.NoError(t, dbtx.Load(ctx))
	assert.NoError(t, handler.Load(ctx))

	assert.NoError(t, Load(ctx))

	mock := ServeConf{
		RunTimeProfilingEnabled: true,
		SwaggerEndpointsEnabled: true,
		InfoEndpointEnabled:     true,
		InterestedEndpoint:      nil,
		AttachRequestID:         nil,
		AttachRequestLogger:     nil,
		MetricsLogger:           nil,
		ServerTiming:            nil,
		DocH:                    handler.GetHandler(),
	}
	mock.CreateServer(context.Background())
}
