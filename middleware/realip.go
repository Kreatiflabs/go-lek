package middleware

import (
	"net/http"
	"strings"
)

// RealIP is a middleware that sets RemoteAddr to the client's real IP
// based on trusted proxy headers (X-Forwarded-For, X-Real-IP).
func RealIP(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
			ips := strings.Split(xff, ",")
			if len(ips) > 0 {
				r.RemoteAddr = strings.TrimSpace(ips[0])
				next.ServeHTTP(w, r)
				return
			}
		}

		if xri := r.Header.Get("X-Real-IP"); xri != "" {
			r.RemoteAddr = strings.TrimSpace(xri)
		}

		next.ServeHTTP(w, r)
	})
}
