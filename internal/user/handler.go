package user

import (
	"encoding/json"
	"net/http"
	"welog/internal/auth"
)

type UserHandler struct {
	userService *UserService
}

type UpdatePreferenceRequest struct {
	AIPreference uint `json:"ai_preference"`
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
	if err != nil || user == nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"user_id":       user.ID,
		"email":         user.Email,
		"nickname":      user.Nickname,
		"role":          user.Role,
		"token_count":   user.TokenCount,
		"ai_preference": user.AIPreference,
	})
}

func (h *UserHandler) GetMyPage(w http.ResponseWriter, r *http.Request) {
	userClaims := auth.GetUserFromContext(r.Context())
	if userClaims == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	resp, err := h.userService.GetMyPage(userClaims.UserID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// PATCH /api/users/preference
func (h *UserHandler) UpdateAIPreference(w http.ResponseWriter, r *http.Request) {
	userClaims := auth.GetUserFromContext(r.Context())
	if userClaims == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req UpdatePreferenceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.AIPreference > 2 {
		http.Error(w, "Invalid AI preference value", http.StatusBadRequest)
		return
	}

	if err := h.userService.UpdateAIPreference(userClaims.UserID, req.AIPreference); err != nil {
		http.Error(w, "Failed to update preference", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Preference updated successfully"})
}
