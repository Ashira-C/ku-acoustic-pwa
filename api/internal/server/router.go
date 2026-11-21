package server

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/Ashira-C/ku-acoustic-pwa/api/internal/health"
	"github.com/Ashira-C/ku-acoustic-pwa/api/internal/httpx"
	"github.com/Ashira-C/ku-acoustic-pwa/api/internal/part"
)

type Deps struct {
	Logger *slog.Logger
	Health *health.Handler
	Parts  *part.Handler
}

func NewRouter(d Deps) *gin.Engine {
	r := gin.New()
	r.Use(httpx.RequestID(), httpx.AccessLog(d.Logger), gin.Recovery())

	r.NoRoute(func(c *gin.Context) {
		httpx.Error(c, http.StatusNotFound, "not_found", "route not found")
	})

	r.GET("/health", d.Health.Live)
	r.GET("/ready", d.Health.Ready)

	v1 := r.Group("/api/v1")
	d.Parts.RegisterRoutes(v1)

	return r
}