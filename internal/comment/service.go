package comment

import (
	"encoding/json"
	"errors"
	"welog/internal/model"
	"welog/internal/notification"
)

type PostRepository interface {
	FindByID(id uint) (*model.Post, error)
}

type UserRepository interface {
	FindByID(id uint) (*model.User, error)
}

type CommentService struct {
	repo                *CommentRepository
	postRepo            PostRepository
	userRepo            UserRepository
	notificationService *notification.NotificationService
}

type CreateCommentParams struct {
	UserID      uint
	PostID      uint
	Description string
	ParentID    *uint
	IsAI        bool
	AIType      *uint
}

func NewCommentService(repo *CommentRepository, postRepo PostRepository, userRepo UserRepository, notificationService *notification.NotificationService) *CommentService {
	return &CommentService{
		repo:                repo,
		postRepo:            postRepo,
		userRepo:            userRepo,
		notificationService: notificationService,
	}
}

func (s *CommentService) CreateComment(params CreateCommentParams) (*model.Comment, error) {
	comment := &model.Comment{
		UserID:      params.UserID,
		PostID:      params.PostID,
		Description: params.Description,
		ParentID:    params.ParentID,
		IsAI:        params.IsAI,
		AIType:      params.AIType,
	}

	if err := s.repo.Create(comment); err != nil {
		return nil, err
	}

	post, err := s.postRepo.FindByID(params.PostID)
	if err == nil && post != nil {
		notificationType := "COMMENT_ADDED"
		if params.IsAI {
			notificationType = "AI_COMMENT_ADDED"
		}

		payload := map[string]interface{}{
			"type":                notificationType,
			"post_id":             params.PostID,
			"post_title":          post.Title,
			"comment_description": comment.Description,
		}
		jsonBytes, err := json.Marshal(payload)
		if err == nil {
			s.notificationService.Notify(post.UserID, string(jsonBytes))
		}
	}

	return comment, nil
}

func (s *CommentService) DeleteComment(userID, commentID uint) error {
	comment, err := s.repo.FindByID(commentID)
	if err != nil {
		return err
	}

	reqUser, err := s.userRepo.FindByID(userID)
	if err != nil || reqUser == nil {
		return errors.New("유저 정보를 찾을 수 없습니다")
	}

	if comment.UserID != userID && reqUser.Role != "ADMIN" {
		return errors.New("댓글을 삭제할 권한이 없습니다")
	}

	return s.repo.Delete(commentID)
}

func (s *CommentService) LikeComment(userID, commentID uint) error {
	comment, err := s.repo.FindByID(commentID)
	if err != nil {
		return err
	}

	if comment.IsAI && comment.AIType != nil {
		if err := s.repo.UpsertPreference(userID, *comment.AIType, 1); err != nil {
			return err
		}
	}

	return s.repo.IncrementLikeCount(commentID)
}
