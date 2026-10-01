package middleware

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

// logResponseWriter wraps http.ResponseWriter to track status and size.
type logResponseWriter struct {
	http.ResponseWriter
	statusCode int
	size       int
}

func (w *logResponseWriter) WriteHeader(code int) {
	w.statusCode = code
	w.ResponseWriter.WriteHeader(code)
}

func (w *logResponseWriter) Write(b []byte) (int, error) {
	if w.statusCode == 0 {
		w.statusCode = http.StatusOK
	}
	size, err := w.ResponseWriter.Write(b)
	w.size += size
	return size, err
}

// LoggerConfig defines configuration for the Logger middleware.
type LoggerConfig struct {
	// Output is the writer to log to (default: os.Stdout)
	Output io.Writer
	// SkipPaths are paths to skip logging (e.g., /health)
	SkipPaths []string
}

// Logger is a middleware that logs each request's method, path,
// status code, response size, and duration.
func Logger(next http.Handler) http.Handler {
	return LoggerWithConfig(LoggerConfig{
		Output: os.Stdout,
	})(next)
}

// LoggerWithConfig returns a Logger middleware with custom configuration.
func LoggerWithConfig(config LoggerConfig) func(http.Handler) http.Handler {
	if config.Output == nil {
		config.Output = os.Stdout
	}
	skipPaths := make(map[string]bool)
	for _, p := range config.SkipPaths {
		skipPaths[p] = true
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if skipPaths[r.URL.Path] {
				next.ServeHTTP(w, r)
				return
			}

			start := time.Now()
			lw := &logResponseWriter{ResponseWriter: w, statusCode: http.StatusOK}

			next.ServeHTTP(lw, r)

			duration := time.Since(start)
			color := "\033[0m"
			switch {
			case lw.statusCode >= 200 && lw.statusCode < 300:
				color = "\033[32m" // Green
			case lw.statusCode >= 300 && lw.statusCode < 400:
				color = "\033[36m" // Cyan
			case lw.statusCode >= 400 && lw.statusCode < 500:
				color = "\033[33m" // Yellow
			case lw.statusCode >= 500:
				color = "\033[31m" // Red
			}

			fmt.Fprintf(config.Output, "%s %s%s %s%s %d %dB %v\n",
				start.Format("2006/01/02 15:04:05"),
				color, r.Method, r.URL.Path, "\033[0m",
				lw.statusCode, lw.size, duration,
			)
		})
	}
}
