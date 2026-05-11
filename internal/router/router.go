package router

import (
	"encoding/json"
	"net/http"
	"welog/internal/auth"
	"welog/internal/post"
	"welog/internal/scheduler"
	"welog/internal/user"

	"gorm.io/gorm"
)

func NewRouter(db *gorm.DB, jwtSecret, googleClientID string) (http.Handler, func()) {
	mux := http.NewServeMux()

	userRepo := user.NewUserRepository(db)
	userService := user.NewUserService(userRepo)
	userHandler := user.NewUserHandler(userService)

	postRepo := post.NewPostRepository(db)
	postService := post.NewPostService(postRepo, userService)
	postHandler := post.NewPostHandler(postService)

	authService := auth.NewAuthService(userRepo, jwtSecret, googleClientID)
	authHandler := auth.NewAuthHandler(authService)

	appScheduler := scheduler.NewScheduler(userService)
	appScheduler.Start()

	mux.HandleFunc("GET /api/ping", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "success",
			"message": "pong",
		})
	})

	// 각 도메인에게 라우팅 전권 위임
	authHandler.RegisterRoutes(mux)
	userHandler.RegisterRoutes(mux)
	postHandler.RegisterRoutes(mux)

	return mux, appScheduler.Stop
}
