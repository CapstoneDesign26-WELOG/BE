package post

import (
	"net/http"
	"welog/internal/auth"
	"welog/pkg/middleware"
)

func (h *PostHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.Handle("POST /api/posts", middleware.Chain(http.HandlerFunc(h.CreatePost), auth.JWTAuthMiddleware))
	mux.HandleFunc("GET /api/posts", h.GetPosts)
	mux.HandleFunc("GET /api/posts/{postId}", h.GetPost)
	mux.Handle("DELETE /api/posts/{postId}", middleware.Chain(http.HandlerFunc(h.DeletePost), auth.JWTAuthMiddleware))
}
