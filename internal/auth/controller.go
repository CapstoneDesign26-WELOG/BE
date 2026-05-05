package auth

import (
	"encoding/json"
	"net/http"
)

type AuthHandler struct {
	AuthService *AuthService
}

func NewAuthHandler(authService *AuthService) *AuthHandler {
	return &AuthHandler{AuthService: authService}
}

type GoogleLoginRequest struct {
	GoogleToken string `json: "google_token"`
}

func (h *AuthHandler) GoogleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req GoogleLoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	accessToken, userObj, err := h.AuthService.ProcessGoogleLogin(r.Context(), req.GoogleToken)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"access_token": accessToken,
		"user": map[string]interface{}{
			"id":    userObj.ID,
			"email": userObj.Email,
			"role":  userObj.Role,
		},
	})
}
