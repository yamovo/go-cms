package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// APIResponse is the unified API response structure.
type APIResponse struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Meta    *Meta       `json:"meta,omitempty"`
}

// Meta contains pagination metadata.
type Meta struct {
	Page     int   `json:"page"`
	PageSize int   `json:"page_size"`
	Total    int64 `json:"total"`
	HasNext  bool  `json:"has_next"`
	HasPrev  bool  `json:"has_prev"`
}

// Success sends a 200 success response.
func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, APIResponse{
		Code:    0,
		Message: "success",
		Data:    data,
	})
}

// SuccessWithMeta sends a 200 success response with pagination.
func SuccessWithMeta(c *gin.Context, data interface{}, meta *Meta) {
	c.JSON(http.StatusOK, APIResponse{
		Code:    0,
		Message: "success",
		Data:    data,
		Meta:    meta,
	})
}

// Created sends a 201 created response.
func Created(c *gin.Context, data interface{}) {
	c.JSON(http.StatusCreated, APIResponse{
		Code:    0,
		Message: "created",
		Data:    data,
	})
}

// Error sends an error response with the given status code.
func Error(c *gin.Context, status int, code string, message string) {
	c.JSON(status, APIResponse{
		Code:    -1,
		Message: message,
	})
}

// BadRequest sends a 400 error response.
func BadRequest(c *gin.Context, message string) {
	Error(c, http.StatusBadRequest, "BAD_REQUEST", message)
}

// Unauthorized sends a 401 error response.
func Unauthorized(c *gin.Context, message string) {
	Error(c, http.StatusUnauthorized, "UNAUTHORIZED", message)
}

// Forbidden sends a 403 error response.
func Forbidden(c *gin.Context, message string) {
	Error(c, http.StatusForbidden, "FORBIDDEN", message)
}

// NotFound sends a 404 error response.
func NotFound(c *gin.Context, message string) {
	Error(c, http.StatusNotFound, "NOT_FOUND", message)
}

// Conflict sends a 409 error response.
func Conflict(c *gin.Context, message string) {
	Error(c, http.StatusConflict, "CONFLICT", message)
}

// TooManyRequests sends a 429 error response.
func TooManyRequests(c *gin.Context, message string) {
	Error(c, http.StatusTooManyRequests, "RATE_LIMIT_EXCEEDED", message)
}

// InternalError sends a 500 error response.
func InternalError(c *gin.Context) {
	Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Internal server error")
}
