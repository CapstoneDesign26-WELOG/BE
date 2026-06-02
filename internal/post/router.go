package post

import (
	"net/http"
	"welog/internal/auth"
	"welog/pkg/middleware"
)

func (h *PostHandler) RegisterRoutes(mux *http.ServeMux, secretKey []byte) {
	mux.Handle("POST /api/posts", middleware.Chain(http.HandlerFunc(h.CreatePost), auth.JWTAuthMiddleware(secretKey)))
	mux.Handle("GET /api/posts/public", middleware.Chain(http.HandlerFunc(h.GetPublicPosts), auth.JWTAuthMiddleware(secretKey)))
	mux.Handle("GET /api/posts/private", middleware.Chain(http.HandlerFunc(h.GetPrivatePosts), auth.JWTAuthMiddleware(secretKey)))
	mux.Handle("GET /api/posts/{postId}", middleware.Chain(http.HandlerFunc(h.GetPost), auth.JWTAuthMiddleware(secretKey)))
	mux.Handle("DELETE /api/posts/{postId}", middleware.Chain(http.HandlerFunc(h.DeletePost), auth.JWTAuthMiddleware(secretKey)))
}
