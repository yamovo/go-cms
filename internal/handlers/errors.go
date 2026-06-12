package handlers

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/vortexcms/go-cms/internal/errs"
	"gorm.io/gorm"
)

// handleServiceError maps service errors to HTTP responses.
func handleServiceError(c *gin.Context, err error) {
	// Check for AppError first.
	var appErr *errs.AppError
	if ok := errs.Is(err, &appErr); ok {
		resp := gin.H{
			"code":    appErr.Code,
			"error":   appErr.Message,
		}
		c.JSON(appErr.StatusCode(), resp)
		return
	}

	// Check for statusCoder interface (legacy).
	type statusCoder interface {
		StatusCode() int
	}
	if sc, ok := err.(statusCoder); ok {
		c.JSON(sc.StatusCode(), gin.H{"error": err.Error()})
		return
	}

	// Check for common errors.
	if err == gorm.ErrRecordNotFound {
		c.JSON(http.StatusNotFound, gin.H{
			"code":  "NOT_FOUND",
			"error": "Resource not found",
		})
		return
	}

	// Log the unexpected error with context.
	slog.Error("unhandled service error",
		"error", err,
		"path", c.Request.URL.Path,
		"method", c.Request.Method,
	)

	// Return generic error to client.
	c.JSON(http.StatusInternalServerError, gin.H{
		"code":  "INTERNAL_ERROR",
		"error": "Internal server error",
	})
}
