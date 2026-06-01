package comment

import (
	"net/http"
	"welog/internal/auth"
	"welog/pkg/middleware"
)

func (h *CommentHandler) RegisterRoutes(mux *http.ServeMux, secretKey []byte) {
	mux.Handle("POST /api/posts/{postId}/comments", middleware.Chain(http.HandlerFunc(h.CreateComment), auth.JWTAuthMiddleware(secretKey)))
	mux.Handle("DELETE /api/comments/{commentId}", middleware.Chain(http.HandlerFunc(h.DeleteComment), auth.JWTAuthMiddleware(secretKey)))
	mux.Handle("POST /api/comments/{commentId}/like", middleware.Chain(http.HandlerFunc(h.LikeComment), auth.JWTAuthMiddleware(secretKey)))
	mux.Handle("DELETE /api/comments/{commentId}/unlike", middleware.Chain(http.HandlerFunc(h.UnlikeComment), auth.JWTAuthMiddleware(secretKey)))
}
