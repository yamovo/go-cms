package errs

import (
	"fmt"
	"net/http"
)

// AppError represents a structured application error with an error code.
type AppError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Status  int    `json:"-"`
	Err     error  `json:"-"`
}

// Error implements the error interface.
func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Code, e.Err)
	}
	return e.Code
}

// Unwrap supports errors.Is/As.
func (e *AppError) Unwrap() error {
	return e.Err
}

// StatusCode returns the HTTP status code for this error.
func (e *AppError) StatusCode() int {
	return e.Status
}

// Wrap wraps an underlying error into an AppError.
func (e *AppError) Wrap(err error) *AppError {
	return &AppError{
		Code:    e.Code,
		Message: e.Message,
		Status:  e.Status,
		Err:     err,
	}
}

// WithMessage returns a copy with a custom message.
func (e *AppError) WithMessage(msg string) *AppError {
	return &AppError{
		Code:    e.Code,
		Message: msg,
		Status:  e.Status,
		Err:     e.Err,
	}
}

// ──────────────────────────────────────────────────────────────────────────────
// Pre-defined errors
// ──────────────────────────────────────────────────────────────────────────────

var (
	// 400 Bad Request
	ErrBadRequest = &AppError{Code: "BAD_REQUEST", Message: "Bad request", Status: http.StatusBadRequest}
	ErrValidation = &AppError{Code: "VALIDATION_ERROR", Message: "Validation failed", Status: http.StatusUnprocessableEntity}

	// 401 Unauthorized
	ErrUnauthorized  = &AppError{Code: "UNAUTHORIZED", Message: "Authentication required", Status: http.StatusUnauthorized}
	ErrInvalidCreds  = &AppError{Code: "INVALID_CREDENTIALS", Message: "Invalid username or password", Status: http.StatusUnauthorized}
	ErrTokenExpired  = &AppError{Code: "TOKEN_EXPIRED", Message: "Token has expired", Status: http.StatusUnauthorized}
	ErrTokenRevoked  = &AppError{Code: "TOKEN_REVOKED", Message: "Token has been revoked", Status: http.StatusUnauthorized}

	// 403 Forbidden
	ErrForbidden     = &AppError{Code: "FORBIDDEN", Message: "Insufficient permissions", Status: http.StatusForbidden}
	ErrAccountLocked = &AppError{Code: "ACCOUNT_LOCKED", Message: "Account temporarily locked due to too many failed attempts", Status: http.StatusForbidden}
	ErrAccountDisabled = &AppError{Code: "ACCOUNT_DISABLED", Message: "Account is disabled", Status: http.StatusForbidden}

	// 404 Not Found
	ErrNotFound = &AppError{Code: "NOT_FOUND", Message: "Resource not found", Status: http.StatusNotFound}

	// 409 Conflict
	ErrConflict      = &AppError{Code: "CONFLICT", Message: "Resource already exists", Status: http.StatusConflict}
	ErrDuplicateUser = &AppError{Code: "DUPLICATE_USER", Message: "Username or email already exists", Status: http.StatusConflict}

	// 429 Too Many Requests
	ErrRateLimitExceeded = &AppError{Code: "RATE_LIMIT_EXCEEDED", Message: "Too many requests", Status: http.StatusTooManyRequests}

	// 500 Internal Server Error
	ErrInternal = &AppError{Code: "INTERNAL_ERROR", Message: "Internal server error", Status: http.StatusInternalServerError}

	// 503 Service Unavailable
	ErrServiceUnavailable = &AppError{Code: "SERVICE_UNAVAILABLE", Message: "Service temporarily unavailable", Status: http.StatusServiceUnavailable}
)

// New creates a new AppError with the given code and message.
func New(code string, message string, status int) *AppError {
	return &AppError{Code: code, Message: message, Status: status}
}

// Is checks if err is an *AppError and assigns it to target.
// Usage: if ok := errs.Is(err, &appErr); ok { ... }
func Is(err error, target **AppError) bool {
	if err == nil {
		return false
	}
	if e, ok := err.(*AppError); ok {
		*target = e
		return true
	}
	// Check wrapped error.
	if e, ok := err.(interface{ Unwrap() error }); ok {
		return Is(e.Unwrap(), target)
	}
	return false
}
