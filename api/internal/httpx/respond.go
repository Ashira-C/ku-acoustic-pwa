package httpx

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
)

type errorBody struct {
	Error errorDetail `json:"error"`
}

type errorDetail struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// Error writes the standard error response and stops the handler chain.
func Error(c *gin.Context, status int, code, message string) {
	c.AbortWithStatusJSON(status, errorBody{Error: errorDetail{Code: code, Message: message}})
}

// InternalError logs the real cause and returns a generic 500,
// so internal details (SQL, hostnames) never leak to clients.
func InternalError(c *gin.Context, logger *slog.Logger, err error) {
	logger.ErrorContext(c.Request.Context(), "internal error",
		slog.String("request_id", RequestIDFrom(c)),
		slog.String("err", err.Error()),
	)
	Error(c, http.StatusInternalServerError, "internal_error", "something went wrong")
}