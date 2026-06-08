package comment

import (
	"encoding/json"
	"errors"
	"math/rand/v2"
	"welog/internal/model"
	"welog/internal/notification"
	"welog/pkg/filter"

	"gorm.io/gorm"
)

var ErrProfanityDetected = errors.New("비속어가 포함된 게시글은 작성하실 수 없습니다")

type PostReactor interface {
	ReplyToUserComment(postID, targetCommentID, commentUserID uint, userComment string)
}

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
	postReactor         PostReactor
}

type CreateCommentParams struct {
	UserID       uint
	PostID       uint
	Description  string
	ParentID     *uint
	IsAI         bool
	AIType       *uint
	SystemPrompt string
	UserPrompt   string
}

func NewCommentService(repo *CommentRepository, postRepo PostRepository, userRepo UserRepository, notificationService *notification.NotificationService) *CommentService {
	return &CommentService{
		repo:                repo,
		postRepo:            postRepo,
		userRepo:            userRepo,
		notificationService: notificationService,
	}
}

func (s *CommentService) SetPostReactor(pr PostReactor) {
	s.postReactor = pr
}

func (s *CommentService) CreateComment(params CreateCommentParams) (*model.Comment, error) {
	if !params.IsAI {
		if !filter.ValidateLength(params.Description, 500) {
			return nil, errors.New("댓글은 최대 500자까지만 작성 가능합니다")
		}
		if filter.ContainsProfanity(params.Description) {
			return nil, ErrProfanityDetected
		}
	}

	comment := &model.Comment{
		UserID:       params.UserID,
		PostID:       params.PostID,
		Description:  params.Description,
		ParentID:     params.ParentID,
		IsAI:         params.IsAI,
		AIType:       params.AIType,
		SystemPrompt: params.SystemPrompt,
		UserPrompt:   params.UserPrompt,
	}

	if err := s.repo.Create(comment); err != nil {
		return nil, err
	}

	if !params.IsAI && params.ParentID == nil && s.postReactor != nil {
		if rand.Float32() < 0.8 {
			go s.postReactor.ReplyToUserComment(params.PostID, comment.ID, params.UserID, params.Description)
		}
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
		if err == nil && post.UserID != params.UserID {
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
	return s.repo.Transaction(func(txRepo *CommentRepository) error {
		comment, err := txRepo.FindByID(commentID)
		if err != nil {
			return err
		}

		like, err := txRepo.GetLike(userID, commentID)
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}

		if like != nil {
			return errors.New("이미 좋아요를 누른 댓글입니다")
		}

		newLike := &model.CommentLike{
			UserID:    userID,
			CommentID: commentID,
		}
		if err := txRepo.CreateLike(newLike); err != nil {
			return err
		}

		if comment.IsAI && comment.AIType != nil {
			if err := txRepo.UpsertPreference(userID, *comment.AIType, 1); err != nil {
				return err
			}
		}

		return txRepo.UpdateLikeCount(commentID, 1)
	})
}

func (s *CommentService) UnlikeComment(userID, commentID uint) error {
	return s.repo.Transaction(func(txRepo *CommentRepository) error {
		comment, err := txRepo.FindByID(commentID)
		if err != nil {
			return err
		}

		_, err = txRepo.GetLike(userID, commentID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errors.New("좋아요를 누르지 않은 댓글입니다")
			}
			return err
		}

		if comment.IsAI && comment.AIType != nil {
			if err := txRepo.UpsertPreference(userID, *comment.AIType, -1); err != nil {
				return err
			}
		}

		if err := txRepo.DeleteLike(userID, commentID); err != nil {
			return err
		}

		return txRepo.UpdateLikeCount(commentID, -1)
	})
}

func (s *CommentService) GetLikedMap(userID uint, commentsIDs []uint) (map[uint]bool, error) {
	return s.repo.GetLikedCommentsIDs(userID, commentsIDs)
}
