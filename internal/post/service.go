package post

import (
	"errors"
	"welog/internal/model"
)

type PostService struct {
	repo *PostRepository
	// UserRepo(토큰 차감), AI Service(비동기 호출) 추가 예정
}

func NewPostService(repo *PostRepository) *PostService {
	return &PostService{repo: repo}
}

func (s *PostService) CreatePost(userID uint, title, description string, postType uint) (*model.Post, error) {
	// 잔여 토큰 확인 후 1회 차감 로직 추가 예정

	post := &model.Post{
		UserID:      userID,
		Title:       title,
		Description: description,
		Type:        postType,
		Count:       0,
	}

	if err := s.repo.Create(post); err != nil {
		return nil, err
	}

	// 비동기로 댓글 생성 시작 로직 추가 예정

	return post, nil
}

func (s *PostService) GetPosts(postType string, page, limit int) ([]model.Post, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	offset := (page - 1) * limit

	var parsedType uint = 0
	switch postType {
	case "PRIVATE":
		parsedType = 1
	case "PUBLIC":
		parsedType = 2
	}

	return s.repo.FindAll(parsedType, offset, limit)
}

func (s *PostService) GetPostDetails(postID uint) (*model.Post, error) {
	return s.repo.FindByID(postID)
}

func (s *PostService) DeletePost(userID, postID uint) error {
	post, err := s.repo.FindByID(postID)
	if err != nil {
		return err
	}

	if post.UserID != userID {
		return errors.New("게시글을 삭제할 권한이 없습니다")
	}

	return s.repo.Delete(postID)
}
