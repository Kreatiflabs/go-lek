package golek

import (
	"fmt"
	"net/http"
	"strings"
)

// HTTPError represents a structured HTTP error with status code and details.
// It implements the error interface and can be used with golek.H() handlers.
type HTTPError struct {
	StatusCode int    `json:"-"`
	Code       string `json:"code"`
	Message    string `json:"message"`
	Details    any    `json:"details,omitempty"`
	Internal   error  `json:"-"` // internal error for logging
}

// Error returns the error message string.
func (e *HTTPError) Error() string {
	if e.Internal != nil {
		return fmt.Sprintf("code=%s, message=%s, internal=%v", e.Code, e.Message, e.Internal)
	}
	return fmt.Sprintf("code=%s, message=%s", e.Code, e.Message)
}

// Unwrap returns the internal error for error wrapping.
func (e *HTTPError) Unwrap() error {
	return e.Internal
}

// WithInternal sets the internal error and returns the HTTPError.
func (e *HTTPError) WithInternal(err error) *HTTPError {
	e.Internal = err
	return e
}

// WithDetails sets the details of the error and returns the HTTPError.
func (e *HTTPError) WithDetails(details any) *HTTPError {
	e.Details = details
	return e
}

// WithCode sets the error code and returns the HTTPError.
func (e *HTTPError) WithCode(code string) *HTTPError {
	e.Code = code
	return e
}

// NewHTTPError creates a new HTTPError with the given status code and optional custom message.
func NewHTTPError(statusCode int, message ...string) *HTTPError {
	msg := http.StatusText(statusCode)
	if len(message) > 0 && message[0] != "" {
		msg = message[0]
	}

	codeStr := strings.ToUpper(strings.ReplaceAll(http.StatusText(statusCode), " ", "_"))
	if codeStr == "" {
		codeStr = "UNKNOWN_ERROR"
	}

	return &HTTPError{
		StatusCode: statusCode,
		Code:       codeStr,
		Message:    msg,
	}
}

// BadRequest creates a 400 Bad Request HTTPError.
func BadRequest(message ...string) *HTTPError {
	return NewHTTPError(http.StatusBadRequest, message...)
}

// Unauthorized creates a 401 Unauthorized HTTPError.
func Unauthorized(message ...string) *HTTPError {
	return NewHTTPError(http.StatusUnauthorized, message...)
}

// Forbidden creates a 403 Forbidden HTTPError.
func Forbidden(message ...string) *HTTPError {
	return NewHTTPError(http.StatusForbidden, message...)
}

// NotFound creates a 404 Not Found HTTPError.
func NotFound(message ...string) *HTTPError {
	return NewHTTPError(http.StatusNotFound, message...)
}

// MethodNotAllowedError creates a 405 Method Not Allowed HTTPError.
func MethodNotAllowedError(message ...string) *HTTPError {
	return NewHTTPError(http.StatusMethodNotAllowed, message...)
}

// Conflict creates a 409 Conflict HTTPError.
func Conflict(message ...string) *HTTPError {
	return NewHTTPError(http.StatusConflict, message...)
}

// UnprocessableEntity creates a 422 Unprocessable Entity HTTPError.
func UnprocessableEntity(message ...string) *HTTPError {
	return NewHTTPError(http.StatusUnprocessableEntity, message...)
}

// TooManyRequests creates a 429 Too Many Requests HTTPError.
func TooManyRequests(message ...string) *HTTPError {
	return NewHTTPError(http.StatusTooManyRequests, message...)
}

// InternalServerError creates a 500 Internal Server Error HTTPError.
func InternalServerError(message ...string) *HTTPError {
	return NewHTTPError(http.StatusInternalServerError, message...)
}

// ServiceUnavailable creates a 503 Service Unavailable HTTPError.
func ServiceUnavailable(message ...string) *HTTPError {
	return NewHTTPError(http.StatusServiceUnavailable, message...)
}

// ValidationError represents a field validation error.
type ValidationError struct {
	Field   string `json:"field"`
	Tag     string `json:"tag"`
	Value   any    `json:"value,omitempty"`
	Message string `json:"message"`
}

// ValidationErrors is a collection of validation errors.
type ValidationErrors []ValidationError

// Error returns the string representation of validation errors.
func (ve ValidationErrors) Error() string {
	return fmt.Sprintf("%d validation errors occurred", len(ve))
}

// defaultErrorHandler is the built-in error handler.
// It checks if error is *HTTPError and writes JSON response.
// For ValidationErrors, it wraps them in a 422 HTTPError.
// For unknown errors, it returns 500.
func defaultErrorHandler(c *Context, err error) {
	if c.Writer.Written() {
		return
	}

	var httpErr *HTTPError
	switch e := err.(type) {
	case *HTTPError:
		httpErr = e
	case ValidationErrors:
		httpErr = UnprocessableEntity("request validation failed").
			WithCode("VALIDATION_ERROR").
			WithDetails(e)
	default:
		httpErr = InternalServerError().WithInternal(err)
	}

	_ = c.JSON(httpErr.StatusCode, httpErr)
}
