package user

import (
	"time"
	"welog/internal/model"
)

type PostRepository interface {
	FindAllByUserID(userID uint) ([]model.Post, error)
}

type CommentRepository interface {
	FindAllByUserIDWithPost(userID uint) ([]model.Comment, error)
}

type UserService struct {
	repo        *UserRepository
	postRepo    PostRepository
	commentRepo CommentRepository
}

type MyPageComment struct {
	ID               uint      `json:"id"`
	Description      string    `json:"description"`
	CreatedAt        time.Time `json:"created_at"`
	PostTitle        string    `json:"post_title"`
	PostCreatedAt    time.Time `json:"post_created_at"`
	PostCommentCount int       `json:"post_comment_count"`
	LikeCount        uint      `json:"like_count"`
}

type MyPageResponse struct {
	User     *model.User     `json:"user"`
	Posts    []model.Post    `json:"posts"`
	Comments []MyPageComment `json:"comments"`
}

func NewUserService(repo *UserRepository, postRepo PostRepository, commentRepo CommentRepository) *UserService {
	return &UserService{
		repo:        repo,
		postRepo:    postRepo,
		commentRepo: commentRepo,
	}
}

func (s *UserService) GetUser(userID uint) (*model.User, error) {
	return s.repo.FindByID(userID)
}

func (s *UserService) GetMyPage(userID uint) (*MyPageResponse, error) {
	user, err := s.repo.FindByID(userID)
	if err != nil {
		return nil, err
	}

	posts, err := s.postRepo.FindAllByUserID(userID)
	if err != nil {
		return nil, err
	}
	comments, err := s.commentRepo.FindAllByUserIDWithPost(userID)
	if err != nil {
		return nil, err
	}

	myPageComments := make([]MyPageComment, 0, len(comments))
	for _, c := range comments {
		myPageComments = append(myPageComments, MyPageComment{
			ID:               c.ID,
			Description:      c.Description,
			CreatedAt:        c.CreatedAt,
			PostTitle:        c.Post.Title,
			PostCreatedAt:    c.Post.CreatedAt,
			PostCommentCount: len(c.Post.Comments),
			LikeCount:        c.LikeCount,
		})
	}

	return &MyPageResponse{
		User:     user,
		Posts:    posts,
		Comments: myPageComments,
	}, nil
}

func (s *UserService) ConsumeToken(userID, amount uint) error {
	return s.repo.ConsumeToken(userID, amount)
}

func (s *UserService) RefundToken(userID, amount uint) error {
	return s.repo.RefundToken(userID, amount)
}

func (s *UserService) RefillDailyTokens() error {
	const threshold uint = 3
	const targetCount uint = 3
	return s.repo.UpdateTokenBelowThreshold(threshold, targetCount)
}
