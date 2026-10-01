package middleware

import (
	"bytes"
	"context"
	"net/http"
	"sync"
	"time"
)

type timeoutWriter struct {
	http.ResponseWriter
	body        *bytes.Buffer
	wroteHeader bool
	statusCode  int
	mu          sync.Mutex
	timedOut    bool
}

func (w *timeoutWriter) WriteHeader(code int) {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.timedOut || w.wroteHeader {
		return
	}
	w.wroteHeader = true
	w.statusCode = code
}

func (w *timeoutWriter) Write(b []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.timedOut {
		return 0, http.ErrHandlerTimeout
	}
	if !w.wroteHeader {
		w.statusCode = http.StatusOK
		w.wroteHeader = true
	}
	return w.body.Write(b)
}

// Timeout returns a middleware that cancels the request context after the
// specified duration and returns 504 Gateway Timeout if the handler doesn't
// complete in time.
func Timeout(duration time.Duration) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx, cancel := context.WithTimeout(r.Context(), duration)
			defer cancel()

			r = r.WithContext(ctx)
			done := make(chan struct{})
			tw := &timeoutWriter{
				ResponseWriter: w,
				body:           &bytes.Buffer{},
				statusCode:     http.StatusOK,
			}

			go func() {
				defer close(done)
				next.ServeHTTP(tw, r)
			}()

			select {
			case <-ctx.Done():
				tw.mu.Lock()
				tw.timedOut = true
				tw.mu.Unlock()
				http.Error(w, http.StatusText(http.StatusGatewayTimeout), http.StatusGatewayTimeout)
			case <-done:
				tw.mu.Lock()
				defer tw.mu.Unlock()
				if tw.wroteHeader {
					w.WriteHeader(tw.statusCode)
				}
				w.Write(tw.body.Bytes())
			}
		})
	}
}
