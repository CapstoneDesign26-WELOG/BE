package post

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"
	"welog/internal/auth"
	"welog/internal/model"
)

type PostHandler struct {
	service *PostService
}

type CommentResponse struct {
	ID          uint      `json:"id"`
	UserID      uint      `json:"user_id"`
	Nickname    string    `json:"nickname"`
	Description string    `json:"description"`
	IsAI        bool      `json:"is_ai"`
	LikeCount   uint      `json:"like_count"`
	IsLiked     bool      `json:"is_liked"`
	CreatedAt   time.Time `json:"created_at"`
	ParentID    *uint     `json:"parent_id"`
}

type PostDetailResponse struct {
	ID           uint              `json:"id"`
	UserID       uint              `json:"user_id"`
	Title        string            `json:"title"`
	Description  string            `json:"description"`
	Type         uint              `json:"type"`
	CommentCount uint              `json:"comment_count"`
	CreatedAt    time.Time         `json:"created_at"`
	Comments     []CommentResponse `json:"comments"`
}

type CreatePostRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Type        string `json:"type"`
}

func NewPostHandler(service *PostService) *PostHandler {
	return &PostHandler{service: service}
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

	postType := model.PostTypePrivate
	if req.Type == "PUBLIC" {
		postType = model.PostTypePublic
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
	userClaims := auth.GetUserFromContext(r.Context())

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

	var likedMap map[uint]bool
	if userClaims != nil {
		commentIDs := make([]uint, len(post.Comments))
		for i, c := range post.Comments {
			commentIDs[i] = c.ID
		}
		likedMap, _ = h.service.commentService.GetLikedMap(userClaims.UserID, commentIDs)
	}

	commentResponses := make([]CommentResponse, len(post.Comments))
	for i, c := range post.Comments {
		isLiked := false
		if likedMap != nil {
			isLiked = likedMap[c.ID]
		}
		commentResponses[i] = CommentResponse{
			ID:          c.ID,
			UserID:      c.UserID,
			Nickname:    c.User.Nickname,
			Description: c.Description,
			IsAI:        c.IsAI,
			LikeCount:   c.LikeCount,
			IsLiked:     isLiked,
			CreatedAt:   c.CreatedAt,
			ParentID:    c.ParentID,
		}
	}

	response := PostDetailResponse{
		ID:           post.ID,
		UserID:       post.UserID,
		Title:        post.Title,
		Description:  post.Description,
		Type:         post.Type,
		CommentCount: post.CommentCount,
		CreatedAt:    post.CreatedAt,
		Comments:     commentResponses,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
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
