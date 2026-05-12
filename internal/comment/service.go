package comment

import (
	"errors"
	"fmt"
	"welog/internal/model"
	"welog/internal/notification"
)

type PostRepository interface {
	FindByID(id uint) (*model.Post, error)
}

type CommentService struct {
	repo                *CommentRepository
	postRepo            PostRepository
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

func NewCommentService(repo *CommentRepository, postRepo PostRepository, notificationService *notification.NotificationService) *CommentService {
	return &CommentService{
		repo:                repo,
		postRepo:            postRepo,
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
		s.notificationService.Notify(post.UserID, fmt.Sprintf(`{"type": "%s", "post_id": %d}`, notificationType, params.PostID))
	}

	return comment, nil
}

func (s *CommentService) DeleteComment(userID, commentID uint) error {
	comment, err := s.repo.FindByID(commentID)
	if err != nil {
		return err
	}

	if comment.UserID != userID && comment.User.Role != "ADMIN" {
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
		return s.repo.UpsertPreference(userID, *comment.AIType, 1)
	}
	return nil
}
