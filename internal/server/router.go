package server

import (
	"crypto/sha256"
	"crypto/subtle"
	"net/http"

	"go-wordpress/api"
	"go-wordpress/internal/apikey"
	"go-wordpress/internal/config"
	"go-wordpress/internal/extract"
	"go-wordpress/internal/health"
	"go-wordpress/internal/metrics"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// RegisterRoutes wires every endpoint:
//
//	POST /api/v1/extract                      API key in X-API-Key
//	GET  /health, /ready                      public
//	GET  /swagger/*, /openapi.yaml, /metrics  operator credentials (Basic)
func RegisterRoutes(engine *gin.Engine, h *health.Health, ex *extract.Handler, keys *apikey.Store, m *metrics.Metrics, cfg *config.Config) {
	engine.GET("/health", h.Handle)
	engine.GET("/ready", h.Ready)

	v1 := engine.Group("/api/v1")
	v1.POST("/extract", apikey.Middleware(keys, func() { m.Reject("unauthorized") }), ex.Handle)

	ops := engine.Group("/", OpsAuth(cfg))
	ops.GET("/openapi.yaml", func(c *gin.Context) {
		c.Data(http.StatusOK, "application/yaml; charset=utf-8", api.OpenAPI)
	})
	ops.GET("/metrics", func(c *gin.Context) {
		c.Header("Content-Type", "text/plain; version=0.0.4; charset=utf-8")
		c.Status(http.StatusOK)
		m.Write(c.Writer)
	})
	ops.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler, ginSwagger.URL("/openapi.yaml")))

	engine.NoRoute(notFound)
}

func notFound(c *gin.Context) {
	c.AbortWithStatusJSON(http.StatusNotFound, gin.H{
		"code":    "NOT_FOUND",
		"message": "The requested endpoint does not exist.",
	})
}

// OpsAuth protects operational endpoints with HTTP Basic authentication.
// The password is checked against APP_OPS_PASSWORD_HASH, a keyed hash made
// with the API key pepper. When either setting is missing the endpoints
// return 404. Client API keys are never accepted here.
func OpsAuth(cfg *config.Config) gin.HandlerFunc {
	wantUser := sha256.Sum256([]byte(cfg.OpsUser))
	return func(c *gin.Context) {
		if !cfg.OpsEnabled() {
			notFound(c)
			return
		}
		user, pass, ok := c.Request.BasicAuth()
		gotUser := sha256.Sum256([]byte(user))
		userOK := subtle.ConstantTimeCompare(gotUser[:], wantUser[:]) == 1
		passOK := apikey.HashEqual(cfg.APIKeyPepper, pass, cfg.OpsPasswordHash)
		if !ok || !userOK || !passOK {
			c.Header("WWW-Authenticate", `Basic realm="operations", charset="UTF-8"`)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"code": "UNAUTHORIZED", "message": "Operator credentials are required."})
			return
		}
		c.Next()
	}
}
