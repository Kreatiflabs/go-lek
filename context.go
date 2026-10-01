package golek

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Params is a slice of URL parameters
type Params []Param

// Param represents a single URL parameter
type Param struct {
	Key   string
	Value string
}

// Get returns the value of a parameter by key
func (ps Params) Get(key string) string {
	for _, p := range ps {
		if p.Key == key {
			return p.Value
		}
	}
	return ""
}

// Context carries the request and response through the handler chain.
// It provides helper methods for reading request data and writing responses.
type Context struct {
	// Writer is the response writer (wrapped for status tracking)
	Writer ResponseWriter

	// Request is the HTTP request
	Request *http.Request

	// engine reference for binder/validator/error handler
	engine *Engine

	// params holds URL parameters extracted from the route
	params Params

	// query caches parsed query parameters
	query url.Values

	// store holds request-scoped key-value pairs
	store map[string]any
}

// Param returns the value of the URL parameter with the given name.
func (c *Context) Param(name string) string {
	return c.params.Get(name)
}

// Query returns the value of the URL query parameter with the given name.
func (c *Context) Query(name string) string {
	if c.query == nil {
		c.query = c.Request.URL.Query()
	}
	return c.query.Get(name)
}

// DefaultQuery returns the value of the URL query parameter with the given name.
// If the parameter is not present, it returns the default value.
func (c *Context) DefaultQuery(name, defaultValue string) string {
	if value := c.Query(name); value != "" {
		return value
	}
	return defaultValue
}

// QueryArray returns a slice of strings for a given query parameter.
func (c *Context) QueryArray(name string) []string {
	if c.query == nil {
		c.query = c.Request.URL.Query()
	}
	return c.query[name]
}

// Header returns the value of the request header with the given name.
func (c *Context) Header(name string) string {
	return c.Request.Header.Get(name)
}

// SetHeader sets the value of the response header with the given name.
func (c *Context) SetHeader(name, value string) {
	c.Writer.Header().Set(name, value)
}

// Cookie returns the named cookie provided in the request.
func (c *Context) Cookie(name string) (*http.Cookie, error) {
	return c.Request.Cookie(name)
}

// SetCookie adds a Set-Cookie header to the response.
func (c *Context) SetCookie(cookie *http.Cookie) {
	http.SetCookie(c.Writer, cookie)
}

// ClientIP attempts to parse and return the client's real IP address.
func (c *Context) ClientIP() string {
	if ip := c.Request.Header.Get("X-Forwarded-For"); ip != "" {
		ips := strings.Split(ip, ",")
		return strings.TrimSpace(ips[0])
	}
	if ip := c.Request.Header.Get("X-Real-IP"); ip != "" {
		return ip
	}
	// Fallback to RemoteAddr
	ip := c.Request.RemoteAddr
	if colon := strings.LastIndex(ip, ":"); colon != -1 {
		ip = ip[:colon]
	}
	return ip
}

// Bind delegates binding the request body to the engine's Binder.
func (c *Context) Bind(v any) error {
	if c.engine != nil && c.engine.Binder != nil {
		return c.engine.Binder.Bind(c, v)
	}
	return nil
}

// Validate delegates struct validation to the engine's Validator.
func (c *Context) Validate(v any) error {
	if c.engine != nil && c.engine.Validator != nil {
		return c.engine.Validator.Validate(v)
	}
	return nil
}

// Status sets the HTTP response status code and returns the Context to allow chaining.
func (c *Context) Status(code int) *Context {
	c.Writer.WriteHeader(code)
	return c
}

// JSON sends a JSON response with the given status code.
func (c *Context) JSON(code int, v any) error {
	c.SetHeader("Content-Type", "application/json")
	c.Status(code)
	encoder := json.NewEncoder(c.Writer)
	return encoder.Encode(v)
}

// XML sends an XML response with the given status code.
func (c *Context) XML(code int, v any) error {
	c.SetHeader("Content-Type", "application/xml")
	c.Status(code)
	encoder := xml.NewEncoder(c.Writer)
	return encoder.Encode(v)
}

// Text sends a formatted text response with the given status code.
func (c *Context) Text(code int, format string, values ...any) error {
	c.SetHeader("Content-Type", "text/plain")
	c.Status(code)
	_, err := fmt.Fprintf(c.Writer, format, values...)
	return err
}

// HTML sends an HTML response with the given status code.
func (c *Context) HTML(code int, html string) error {
	c.SetHeader("Content-Type", "text/html")
	c.Status(code)
	_, err := c.Writer.Write([]byte(html))
	return err
}

// Blob sends a byte slice response with the given status code and content type.
func (c *Context) Blob(code int, contentType string, data []byte) error {
	c.SetHeader("Content-Type", contentType)
	c.Status(code)
	_, err := c.Writer.Write(data)
	return err
}

// Stream sends a streaming response with the given status code and content type.
func (c *Context) Stream(code int, contentType string, reader io.Reader) error {
	c.SetHeader("Content-Type", contentType)
	c.Status(code)
	_, err := io.Copy(c.Writer, reader)
	return err
}

// SSEvent sends a Server-Sent Event to the client.
func (c *Context) SSEvent(event string, data any) error {
	c.SetHeader("Content-Type", "text/event-stream")
	c.SetHeader("Cache-Control", "no-cache")
	c.SetHeader("Connection", "keep-alive")

	var dataStr string
	switch v := data.(type) {
	case string:
		dataStr = v
	case []byte:
		dataStr = string(v)
	default:
		b, err := json.Marshal(v)
		if err != nil {
			return err
		}
		dataStr = string(b)
	}

	_, err := fmt.Fprintf(c.Writer, "event: %s\ndata: %s\n\n", event, dataStr)
	if err == nil {
		if flusher, ok := c.Writer.(http.Flusher); ok {
			flusher.Flush()
		}
	}
	return err
}

// NoContent sends an empty response with status 204 No Content.
func (c *Context) NoContent() error {
	c.Status(http.StatusNoContent)
	return nil
}

// Redirect redirects the request to a new URL with the given status code.
func (c *Context) Redirect(code int, url string) error {
	http.Redirect(c.Writer, c.Request, url, code)
	return nil
}

// File serves a static file from the given path.
func (c *Context) File(filepath string) {
	http.ServeFile(c.Writer, c.Request, filepath)
}

// Set stores a new key/value pair exclusively for this context.
func (c *Context) Set(key string, value any) {
	if c.store == nil {
		c.store = make(map[string]any)
	}
	c.store[key] = value
}

// Get returns the value for the given key, if it exists.
func (c *Context) Get(key string) (any, bool) {
	if c.store == nil {
		return nil, false
	}
	value, ok := c.store[key]
	return value, ok
}

// MustGet returns the value for the given key, and panics if it doesn't exist.
func (c *Context) MustGet(key string) any {
	if value, ok := c.Get(key); ok {
		return value
	}
	panic(fmt.Sprintf("Key %q does not exist", key))
}

// reset prepares the Context for the next request in the pool.
func (c *Context) reset(w http.ResponseWriter, r *http.Request) {
	c.Writer = NewResponseWriter(w)
	c.Request = r
	c.params = nil
	c.query = nil
	for k := range c.store {
		delete(c.store, k)
	}
}

// SetParams sets the URL parameters for this context (typically called by router).
func (c *Context) SetParams(params Params) {
	c.params = params
}

// Context returns the request's context.
func (c *Context) Context() context.Context {
	return c.Request.Context()
}

// SetContext sets a new context on the underlying request.
func (c *Context) SetContext(ctx context.Context) {
	c.Request = c.Request.WithContext(ctx)
}

// Deadline implements the context.Context interface.
func (c *Context) Deadline() (time.Time, bool) {
	return c.Request.Context().Deadline()
}

// Done implements the context.Context interface.
func (c *Context) Done() <-chan struct{} {
	return c.Request.Context().Done()
}

// Err implements the context.Context interface.
func (c *Context) Err() error {
	return c.Request.Context().Err()
}

// Value implements the context.Context interface.
func (c *Context) Value(key any) any {
	if keyAsString, ok := key.(string); ok {
		if val, exists := c.Get(keyAsString); exists {
			return val
		}
	}
	return c.Request.Context().Value(key)
}
