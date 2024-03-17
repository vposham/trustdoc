// Package httpsrvr holds the implementations to load server and start the server
package httpsrvr

import (
	"context"

	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"

	"github.com/vposham/trustdoc/handler/base"
)

// CreateServer makes the server that main will run.
// Includes all routing config for app's endpoints.
func (s ServeConf) CreateServer(c context.Context) *gin.Engine {

	router := gin.New()
	router.Use(gin.Recovery())
	s.addDefaultEndpoints(router)
	s.addProfiling(router)
	s.addSwagger(router)
	s.addBusinessEndpoints(c, router)
	return router
}

func (s ServeConf) addProfiling(router *gin.Engine) {
	if s.RunTimeProfilingEnabled {
		pprof.Register(router)
	}
}

func (s ServeConf) addSwagger(router *gin.Engine) {
	if s.SwaggerEndpointsEnabled {
		router.Static("/swaggerui", "doc/swagger-ui-dist")
		router.StaticFile("swagger.json", "doc/swagger.json")
	}
}

func (s ServeConf) addDefaultEndpoints(router *gin.Engine) {
	_ = router.SetTrustedProxies(nil)
	if s.InfoEndpointEnabled {
		router.GET("/info", base.GetInfo)
	}
	router.GET("/health", base.GetHealth)
}

func (s ServeConf) addBusinessEndpointsMiddlewares(router *gin.RouterGroup) {
	router.Use(s.InterestedEndpoint, s.AttachRequestID, s.AttachRequestLogger,
		s.MetricsLogger, s.ServerTiming)
}

func (s ServeConf) addBusinessEndpoints(c context.Context, router *gin.Engine) {
	svcRtr := router.Group("/svc")

	// add all the middlewares
	s.addBusinessEndpointsMiddlewares(svcRtr)

	// versioning to support http api request interfaces future extensibility.
	intVerRtr := svcRtr.Group("/v1")
	// create a group for all endpoints which contains business logic and
	entRtr := intVerRtr.Group("/doc")

	entRtr.GET("/upload", base.GetInfo)        // s.DocH.
	entRtr.GET("/:docId/verify", base.GetInfo) // s.DocH.
}