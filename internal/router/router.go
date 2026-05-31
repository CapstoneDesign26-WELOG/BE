package router

import (
	"encoding/json"
	"net/http"
	"welog/internal/auth"
	"welog/internal/comment"
	"welog/internal/notification"
	"welog/internal/post"
	"welog/internal/scheduler"
	"welog/internal/user"
	"welog/pkg/ai"

	"gorm.io/gorm"
)

func NewRouter(db *gorm.DB, jwtSecret, googleClientID string) (http.Handler, func()) {
	mux := http.NewServeMux()

	userRepo := user.NewUserRepository(db)
	postRepo := post.NewPostRepository(db)
	commentRepo := comment.NewCommentRepository(db)

	clovaClient := ai.NewClovaClient()

	notificationService := notification.NewNotificationService()
	userService := user.NewUserService(userRepo, postRepo, commentRepo)
	authService := auth.NewAuthService(userRepo, jwtSecret, googleClientID)
	commentService := comment.NewCommentService(commentRepo, postRepo, userRepo, notificationService)
	postService := post.NewPostService(postRepo, userService, commentService, clovaClient, notificationService)

	userHandler := user.NewUserHandler(userService)
	postHandler := post.NewPostHandler(postService)
	commentHandler := comment.NewCommentHandler(commentService)
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

	authHandler.RegisterRoutes(mux)
	userHandler.RegisterRoutes(mux, []byte(jwtSecret))
	postHandler.RegisterRoutes(mux, []byte(jwtSecret))
	commentHandler.RegisterRoutes(mux, []byte(jwtSecret))

	mux.Handle("GET /api/notifications/stream", auth.JWTAuthMiddleware([]byte(jwtSecret))(http.HandlerFunc(notificationService.Subscribe)))

	return mux, appScheduler.Stop
}
