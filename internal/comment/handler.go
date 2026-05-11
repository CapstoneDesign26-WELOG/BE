package comment

import (
	"encoding/json"
	"net/http"
	"strconv"
	"welog/internal/auth"
)

type CommentHandler struct {
	service *CommentService
}

func NewCommentHandler(service *CommentService) *CommentHandler {
	return &CommentHandler{
		service: service,
	}
}

type CreateCommentRequest struct {
	Descrtipion string `json:"description"`
	ParentID    *uint  `json:"parent_id"`
}

// POST /api/posts/{postId}/comments
func (h *CommentHandler) CreateComment(w http.ResponseWriter, r *http.Request) {
	userClaims := auth.GetUserFromContext(r.Context())
	if userClaims == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	postID, _ := strconv.ParseUint(r.PathValue("postId"), 10, 32)

	var req CreateCommentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	params := CreateCommentParams{
		UserID:      userClaims.UserID,
		PostID:      uint(postID),
		Description: req.Descrtipion,
		ParentID:    req.ParentID,
		IsAI:        false,
		AIType:      nil,
	}

	comment, err := h.service.CreateComment(params)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"comment_id": comment.ID,
		"created_at": comment.CreatedAt,
	})
}

// DELETE /api/comments/{commentId}
func (h *CommentHandler) DeleteComment(w http.ResponseWriter, r *http.Request) {
	userClaims := auth.GetUserFromContext(r.Context())
	commentID, _ := strconv.ParseUint(r.PathValue("commentId"), 10, 32)

	if err := h.service.DeleteComment(userClaims.UserID, uint(commentID)); err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (h *CommentHandler) LikeComment(w http.ResponseWriter, r *http.Request) {
	userClaims := auth.GetUserFromContext(r.Context())
	commentID, _ := strconv.ParseUint(r.PathValue("commentId"), 10, 32)

	if err := h.service.LikeComment(userClaims.UserID, uint(commentID)); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}
