package middleware

import (
	"net/http"
	"strings"
)

func CorsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")

		allowedOrigins := map[string]bool{
			"https://welog-fe.pages.dev": true,
			"http://localhost:5173":      true,
			"https://www.welog.site":     true,
			"https://welog.site":         true,
		}

		isAllowed := allowedOrigins[origin] || strings.HasSuffix(origin, ".welog-fe.pages.dev")

		if isAllowed {
			w.Header().Set("Access-Control-Allow-Origin", origin)
		}
		w.Header().Add("Vary", "Origin")

		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization, Last-Event-ID")
		w.Header().Set("Access-Control-Max-Age", "43200")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}
