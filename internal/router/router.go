package router

import (
	"encoding/json"
	"net/http"
	"welog/internal/auth"
	"welog/internal/user"
	"welog/pkg/middleware"

	"gorm.io/gorm"
)

func NewRouter(db *gorm.DB, jwtSecret string, googleClientID string) http.Handler {
	mux := http.NewServeMux()

	userRepo := user.NewUserRepository(db)
	authService := auth.NewAuthService(userRepo, jwtSecret, googleClientID)
	authHandler := auth.NewAuthHandler(authService)

	mux.HandleFunc("GET /api/ping", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "success",
			"message": "pong",
		})
	})
	mux.HandleFunc("POST /api/auth/google", authHandler.GoogleLogin)

	return middleware.Chain(mux, middleware.CorsMiddleware)
}
