package middleware

import (
	"net/http"
	"strings"
)

// BasicAuth returns a middleware that enforces HTTP Basic Authentication.
// The credentials map contains username -> password pairs.
func BasicAuth(realm string, credentials map[string]string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user, pass, ok := r.BasicAuth()
			if !ok {
				unauthorizedBasic(w, realm)
				return
			}

			expectedPass, exists := credentials[user]
			if !exists || expectedPass != pass {
				unauthorizedBasic(w, realm)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func unauthorizedBasic(w http.ResponseWriter, realm string) {
	w.Header().Set("WWW-Authenticate", `Basic realm="`+realm+`"`)
	http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
}

// BearerAuth returns a middleware that validates Bearer tokens.
// The validateToken function should return true if the token is valid.
func BearerAuth(validateToken func(token string) bool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
				return
			}

			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
				http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
				return
			}

			if !validateToken(parts[1]) {
				http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
