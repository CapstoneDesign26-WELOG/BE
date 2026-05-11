package user

import (
	"encoding/json"
	"net/http"
	"welog/internal/auth"
)

type UserHandler struct {
	userService *UserService
}

func NewUserHandler(userService *UserService) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}

// GET /api/users/me
func (h *UserHandler) GetMe(w http.ResponseWriter, r *http.Request) {
	userClaims := auth.GetUserFromContext(r.Context())
	if userClaims == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	user, err := h.userService.GetUser(userClaims.UserID)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"email":       user.Email,
		"token_count": user.TokenCount,
	})
}
