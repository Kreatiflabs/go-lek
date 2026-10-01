package middleware

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime"
)

// RecoveryConfig defines configuration for the Recovery middleware.
type RecoveryConfig struct {
	// StackSize is the max stack trace size to print (default: 4096)
	StackSize int
	// DisableStackAll disables capturing all goroutines' stack (default: false)
	DisableStackAll bool
	// Output is the writer for stack traces (default: os.Stderr)
	Output io.Writer
}

// Recovery is a middleware that recovers from panics, logs the stack trace,
// and returns a 500 Internal Server Error response.
func Recovery(next http.Handler) http.Handler {
	return RecoveryWithConfig(RecoveryConfig{})(next)
}

// RecoveryWithConfig returns a Recovery middleware with custom config.
func RecoveryWithConfig(config RecoveryConfig) func(http.Handler) http.Handler {
	if config.StackSize <= 0 {
		config.StackSize = 4096
	}
	if config.Output == nil {
		config.Output = os.Stderr
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					stack := make([]byte, config.StackSize)
					length := runtime.Stack(stack, !config.DisableStackAll)
					fmt.Fprintf(config.Output, "[PANIC RECOVER] %v\n%s\n", err, stack[:length])
					http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}
