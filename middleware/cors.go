package middleware

import (
	"net/http"
	"strconv"
	"strings"
)

// CORSConfig defines CORS policy configuration.
type CORSConfig struct {
	AllowOrigins     []string // default: ["*"]
	AllowMethods     []string // default: ["GET","POST","PUT","DELETE","PATCH","HEAD","OPTIONS"]
	AllowHeaders     []string // default: []
	ExposeHeaders    []string // default: []
	AllowCredentials bool     // default: false
	MaxAge           int      // preflight cache duration in seconds, default: 86400
}

// CORS returns a CORS middleware with default configuration (allow all origins).
func CORS() func(http.Handler) http.Handler {
	return CORSWithConfig(CORSConfig{})
}

// CORSWithConfig returns a CORS middleware with custom configuration.
func CORSWithConfig(config CORSConfig) func(http.Handler) http.Handler {
	if len(config.AllowOrigins) == 0 {
		config.AllowOrigins = []string{"*"}
	}
	if len(config.AllowMethods) == 0 {
		config.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "PATCH", "HEAD", "OPTIONS"}
	}
	if config.MaxAge == 0 {
		config.MaxAge = 86400
	}

	allowOriginsMap := make(map[string]bool)
	allowAllOrigins := false
	for _, o := range config.AllowOrigins {
		if o == "*" {
			allowAllOrigins = true
		}
		allowOriginsMap[o] = true
	}

	allowMethodsStr := strings.Join(config.AllowMethods, ",")
	allowHeadersStr := strings.Join(config.AllowHeaders, ",")
	exposeHeadersStr := strings.Join(config.ExposeHeaders, ",")

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			if origin == "" {
				next.ServeHTTP(w, r)
				return
			}

			allowedOrigin := ""
			if allowAllOrigins {
				if config.AllowCredentials {
					allowedOrigin = origin
				} else {
					allowedOrigin = "*"
				}
			} else if allowOriginsMap[origin] {
				allowedOrigin = origin
			}

			if allowedOrigin != "" {
				w.Header().Set("Access-Control-Allow-Origin", allowedOrigin)
				if config.AllowCredentials {
					w.Header().Set("Access-Control-Allow-Credentials", "true")
				}
			}

			if r.Method == "OPTIONS" && r.Header.Get("Access-Control-Request-Method") != "" {
				w.Header().Set("Access-Control-Allow-Methods", allowMethodsStr)
				if allowHeadersStr != "" {
					w.Header().Set("Access-Control-Allow-Headers", allowHeadersStr)
				} else {
					reqHeaders := r.Header.Get("Access-Control-Request-Headers")
					if reqHeaders != "" {
						w.Header().Set("Access-Control-Allow-Headers", reqHeaders)
					}
				}
				w.Header().Set("Access-Control-Max-Age", strconv.Itoa(config.MaxAge))
				w.WriteHeader(http.StatusNoContent)
				return
			}

			if exposeHeadersStr != "" {
				w.Header().Set("Access-Control-Expose-Headers", exposeHeadersStr)
			}

			next.ServeHTTP(w, r)
		})
	}
}
