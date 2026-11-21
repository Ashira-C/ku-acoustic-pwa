package part

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/Ashira-C/ku-acoustic-pwa/api/internal/httpx"
)

type lister interface {
	List(ctx context.Context) ([]Part, error)
}

type Handler struct {
	svc    lister
	logger *slog.Logger
}

func NewHandler(svc lister, logger *slog.Logger) *Handler {
	return &Handler{svc: svc, logger: logger}
}

func (h *Handler) RegisterRoutes(rg *gin.RouterGroup) {
	rg.GET("/parts", h.list)
}

func (h *Handler) list(c *gin.Context) {
	parts, err := h.svc.List(c.Request.Context())
	if err != nil {
		httpx.InternalError(c, h.logger, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": parts})
}