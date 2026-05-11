package auth

import "net/http"

func (h *AuthHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/auth/google", h.GoogleLogin)

}
