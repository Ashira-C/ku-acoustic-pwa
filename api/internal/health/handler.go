package health

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// CheckFunc reports whether one dependency is reachable.
type CheckFunc func(ctx context.Context) error

type Handler struct {
	checks map[string]CheckFunc
}

func NewHandler(checks map[string]CheckFunc) *Handler {
	return &Handler{checks: checks}
}

// Live: the process is up. No dependency checks, so a slow DB never
// makes the platform think the app itself is dead.
func (h *Handler) Live(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// Ready: the app can actually serve traffic (DB and Redis reachable).
func (h *Handler) Ready(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
	defer cancel()

	status := http.StatusOK
	results := make(map[string]string, len(h.checks))
	for name, check := range h.checks {
		if err := check(ctx); err != nil {
			results[name] = "down"
			status = http.StatusServiceUnavailable
			continue
		}
		results[name] = "up"
	}

	c.JSON(status, gin.H{"checks": results})
}