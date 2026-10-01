package middleware

import (
	"context"
	"crypto/rand"
	"fmt"
	"net/http"
)

// RequestIDHeader is the header name for request ID
const RequestIDHeader = "X-Request-ID"

// RequestIDKey is the context key for the request ID
var RequestIDKey = &struct{ name string }{"RequestID"}

func generateUUID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	b[6] = (b[6] & 0x0f) | 0x40 // version 4
	b[8] = (b[8] & 0x3f) | 0x80 // variant
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}

// RequestID is a middleware that injects a unique request ID into the
// request context and X-Request-ID response header.
// If the request already has an X-Request-ID header, it is reused.
func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqID := r.Header.Get(RequestIDHeader)
		if reqID == "" {
			reqID = generateUUID()
		}

		ctx := context.WithValue(r.Context(), RequestIDKey, reqID)
		w.Header().Set(RequestIDHeader, reqID)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetRequestID returns the request ID from context.
func GetRequestID(r *http.Request) string {
	if val, ok := r.Context().Value(RequestIDKey).(string); ok {
		return val
	}
	return ""
}
