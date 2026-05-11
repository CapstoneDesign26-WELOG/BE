package user

import (
	"net/http"
	"welog/internal/auth"
	"welog/pkg/middleware"
)

func (h *UserHandler) RegisterRoutes(mux *http.ServeMux, secretKey []byte) {
	mux.Handle("GET /api/users/me", middleware.Chain(http.HandlerFunc(h.GetMe), auth.JWTAuthMiddleware(secretKey)))
}
