package comment

import (
	"errors"
	"welog/internal/model"
)

type CommentService struct {
	repo *CommentRepository
}

type CreateCommentParams struct {
	UserID      uint
	PostID      uint
	Description string
	ParentID    *uint
	IsAI        bool
	AIType      *uint
}

func NewCommentService(repo *CommentRepository) *CommentService {
	return &CommentService{
		repo: repo,
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
