package router

import (
	"encoding/json"
	"net/http"
	"welog/middleware"
)

func NewRouter() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/ping", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "success",
			"message": "pong",
		})
	})
	return middleware.Chain(mux, middleware.CorsMiddleware)
}
