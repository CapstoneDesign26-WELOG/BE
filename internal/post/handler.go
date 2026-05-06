package post

import (
	"encoding/json"
	"net/http"
	"strconv"
	"welog/internal/auth"
)

type PostHandler struct {
	service *PostService
}

func NewPostHandler(service *PostService) *PostHandler {
	return &PostHandler{service: service}
}

type CreatePostRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Type        string `json:"type"`
}

// POST /api/posts
func (h *PostHandler) CreatePost(w http.ResponseWriter, r *http.Request) {
	userClaims := auth.GetUserFromContext(r.Context())
	if userClaims == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req CreatePostRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	var postType uint = 1
	if req.Type == "PUBLIC" {
		postType = 2
	}

	post, err := h.service.CreatePost(userClaims.UserID, req.Title, req.Description, postType)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"post_id": post.ID,
		"status":  "AI processing started",
	})
}

// GET /api/posts?type=PUBLIC&page=1&limit=20
func (h *PostHandler) GetPosts(w http.ResponseWriter, r *http.Request) {
	postType := r.URL.Query().Get("type")
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))

	posts, err := h.service.GetPosts(postType, page, limit)
	if err != nil {
		http.Error(w, "Failed to fetch posts", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(posts)
}

// GET /api/posts/{postId}
func (h *PostHandler) GetPost(w http.ResponseWriter, r *http.Request) {
	postIDStr := r.PathValue("postId")
	postID, err := strconv.ParseUint(postIDStr, 10, 32)
	if err != nil {
		http.Error(w, "Invalid post ID", http.StatusBadRequest)
		return
	}

	post, err := h.service.GetPostDetails(uint(postID))
	if err != nil {
		http.Error(w, "Post not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"post":     post,
		"comments": post.Comments,
	})
}

// DELETE /api/posts/{postId}
func (h *PostHandler) DeletePost(w http.ResponseWriter, r *http.Request) {
	userClaims := auth.GetUserFromContext(r.Context())
	if userClaims == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	postIDStr := r.PathValue("postId")
	postID, err := strconv.ParseUint(postIDStr, 10, 32)
	if err != nil {
		http.Error(w, "Invalid post ID", http.StatusBadRequest)
		return
	}

	err = h.service.DeletePost(userClaims.UserID, uint(postID))
	if err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}

	w.WriteHeader(http.StatusOK)
}
