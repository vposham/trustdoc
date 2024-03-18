package httpsrvr

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"net/http"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/vposham/trustdoc/config"
	"github.com/vposham/trustdoc/handler"
	"github.com/vposham/trustdoc/internal/httpsrvr/mwares/logio"
	"github.com/vposham/trustdoc/internal/httpsrvr/mwares/reqlogger"
	"github.com/vposham/trustdoc/internal/httpsrvr/mwares/svrtiming"
	"github.com/vposham/trustdoc/log"
)

var (
	onceInit      = new(sync.Once)
	concreteImpls = make(map[string]any)
)

const (
	// httpSrvrImplKey holds router config
	httpSrvrImplKey = "httpSrvrImpl"
)

// ServeConf holds all the configuration data needed for starting HTTP Gin server.
type ServeConf struct {
	RunTimeProfilingEnabled bool
	SwaggerEndpointsEnabled bool
	InfoEndpointEnabled     bool

	InterestedEndpoint  gin.HandlerFunc
	AttachRequestID     gin.HandlerFunc
	AttachRequestLogger gin.HandlerFunc
	MetricsLogger       gin.HandlerFunc
	ServerTiming        gin.HandlerFunc

	DocH *handler.DocH
}

// Load enables us inject this package as dependency from its parent
func Load(ctx context.Context) error {
	var appErr error
	onceInit.Do(func() {
		appErr = loadImpls(ctx)
	})
	return appErr
}

func loadImpls(_ context.Context) error {
	logger := log.GetConfiguredLogger()
	if concreteImpls[httpSrvrImplKey] == nil {
		randStr, _ := randSeq(6)
		p := config.GetAll()
		l := logio.LogImpl{
			HeaderBlackList: []string{"authorization"},
			LogReqBody:      p.MustGetBool("log.http.req.body"),
			LogRespBody:     p.MustGetBool("log.http.resp.body"),
			LogHeaders:      p.MustGetBool("log.http.req.headers"),
		}

		concreteImpls[httpSrvrImplKey] = ServeConf{
			RunTimeProfilingEnabled: p.MustGetBool("runtime.profiling.enabled"),
			SwaggerEndpointsEnabled: p.MustGetBool("swagger.enabled"),
			InfoEndpointEnabled:     p.MustGetBool("info.endpoint.enabled"),
			InterestedEndpoint:      reqlogger.InterestedEndpoints(),
			AttachRequestID:         reqlogger.AttachRequestID(randStr),
			AttachRequestLogger:     reqlogger.AttachRequestLogger(logger),
			MetricsLogger:           l.HttpTx(log.GetLogger),
			ServerTiming:            svrtiming.Middleware(&svrtiming.MwareOpts{DisableHeaders: false}),
			DocH:                    handler.GetHandler(),
		}
	}
	return nil
}

// Start starts http server
func Start(ctx context.Context, logger *zap.Logger, port string) {

	logger.Info("service up and listening", zap.String("port", port))

	// Create context that listens for the interrupt signal from the OS.
	ctx, stop := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	v, _ := concreteImpls[httpSrvrImplKey].(ServeConf)
	router := v.CreateServer(ctx)

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: router,
	}

	// Initializing the server in a goroutine so that
	// it won't block the graceful shutdown handling below
	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Fatal("http server listen failed", zap.Error(err))
		}
	}()

	// Listen for the interrupt signal.
	<-ctx.Done()

	// Restore default behavior on the interrupt signal and notify user of shutdown.
	stop()
	logger.Info("shutting down gracefully, press Ctrl+C again to force")

	// The context is used to inform the server it has 5 seconds to finish
	// the request it is currently handling
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatal("server forced to shutdown: ", zap.Error(err))
	}

	logger.Info("server exiting")
}

// randSeq returns a URL-safe, base64 encoded
// securely generated random string.
func randSeq(s int) (string, error) {
	b, err := generateRandomBytes(s)
	return base64.URLEncoding.EncodeToString(b)[:s], err
}

// generateRandomBytes returns securely generated random bytes.
// It will return an error if the system's secure random
// number generator fails to function correctly, in which
// case the caller should not continue.
func generateRandomBytes(n int) ([]byte, error) {
	if n < 1 {
		return nil, errors.New("invalid length")
	}
	b := make([]byte, n)
	_, err := rand.Read(b)
	return b, err
}
