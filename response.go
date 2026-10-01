package golek

import (
	"bufio"
	"net"
	"net/http"
)

// ResponseWriter wraps http.ResponseWriter to track status code and bytes written.
type ResponseWriter interface {
	http.ResponseWriter

	// Status returns the HTTP status code written.
	Status() int

	// Written returns whether the response has been written to.
	Written() bool

	// Size returns the number of bytes written.
	Size() int

	// Unwrap returns the underlying http.ResponseWriter.
	Unwrap() http.ResponseWriter
}

type responseWriter struct {
	http.ResponseWriter
	status  int
	size    int
	written bool
}

// NewResponseWriter creates a new wrapped ResponseWriter.
func NewResponseWriter(w http.ResponseWriter) ResponseWriter {
	return &responseWriter{
		ResponseWriter: w,
		status:         http.StatusOK, // default to 200 OK
	}
}

// WriteHeader sends an HTTP response header with the provided status code.
func (w *responseWriter) WriteHeader(statusCode int) {
	if w.written {
		return
	}
	w.status = statusCode
	w.written = true
	w.ResponseWriter.WriteHeader(statusCode)
}

// Write writes the data to the connection as part of an HTTP reply.
func (w *responseWriter) Write(b []byte) (int, error) {
	if !w.written {
		w.WriteHeader(http.StatusOK)
	}
	size, err := w.ResponseWriter.Write(b)
	w.size += size
	return size, err
}

// Status returns the HTTP status code written.
func (w *responseWriter) Status() int {
	return w.status
}

// Written returns whether the response has been written to.
func (w *responseWriter) Written() bool {
	return w.written
}

// Size returns the number of bytes written.
func (w *responseWriter) Size() int {
	return w.size
}

// Unwrap returns the underlying http.ResponseWriter.
func (w *responseWriter) Unwrap() http.ResponseWriter {
	return w.ResponseWriter
}

// Flush implements the http.Flusher interface.
func (w *responseWriter) Flush() {
	if flusher, ok := w.ResponseWriter.(http.Flusher); ok {
		if !w.written {
			w.WriteHeader(http.StatusOK)
		}
		flusher.Flush()
	}
}

// Hijack implements the http.Hijacker interface.
func (w *responseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if hijacker, ok := w.ResponseWriter.(http.Hijacker); ok {
		return hijacker.Hijack()
	}
	return nil, nil, http.ErrNotSupported
}

// Push implements the http.Pusher interface.
func (w *responseWriter) Push(target string, opts *http.PushOptions) error {
	if pusher, ok := w.ResponseWriter.(http.Pusher); ok {
		return pusher.Push(target, opts)
	}
	return http.ErrNotSupported
}
