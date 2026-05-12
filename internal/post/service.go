package post

import (
	"encoding/json"
	"errors"
	"log"
	"strings"
	"welog/internal/comment"
	"welog/internal/model"
	"welog/internal/user"
	"welog/pkg/ai"
)

type PostService struct {
	repo           *PostRepository
	userService    *user.UserService
	commentService *comment.CommentService
	clovaClient    *ai.ClovaClient
}

func NewPostService(repo *PostRepository, userService *user.UserService, commentService *comment.CommentService, clovaClient *ai.ClovaClient) *PostService {
	return &PostService{
		repo:           repo,
		userService:    userService,
		commentService: commentService,
		clovaClient:    clovaClient,
	}
}

func (s *PostService) CreatePost(userID uint, title, description string, postType uint) (*model.Post, error) {
	var tokenCost uint = 1
	if err := s.userService.ConsumeToken(userID, tokenCost); err != nil {
		return nil, err
	}

	post := &model.Post{
		UserID:      userID,
		Title:       title,
		Description: description,
		Type:        postType,
		Count:       0,
	}

	if err := s.repo.Create(post); err != nil {
		// 실패시 토큰 되돌려주는 로직 추가 예정
		return nil, err
	}

	go s.handleAIComments(post.ID, 0, post.Title+"\n"+post.Description)

	return post, nil
}

func (s *PostService) handleAIComments(postID, parentID uint, content string) {
	resp, err := s.clovaClient.GetAIComments(postID, content)
	if err != nil {
		log.Printf("AI 댓글 생성 실패 (PostID: %d): %v", postID, err)
		return
	}

	var aiResults []struct {
		ReactionType string `json:"reaction_type"`
		Comment      string `json:"comment"`
	}

	cleaned := strings.TrimSpace(resp)
	cleaned = strings.TrimPrefix(cleaned, "```json")
	cleaned = strings.TrimPrefix(cleaned, "```")
	cleaned = strings.TrimSuffix(cleaned, "```")
	cleaned = strings.TrimSpace(cleaned)

	if err := json.Unmarshal([]byte(cleaned), &aiResults); err != nil {
		log.Printf("AI 응답 파싱 실패: %v", err)
		return
	}

	typeMap := map[string]uint{
		"A1": 1, "A2": 2, "B1": 3, "B2": 4, "C1": 5, "C2": 6,
	}

	for _, res := range aiResults {
		aiType := typeMap[res.ReactionType]
		_, err := s.commentService.CreateComment(comment.CreateCommentParams{
			// 시스템 유저
			UserID:      1,
			PostID:      postID,
			Description: res.Comment,
			ParentID:    &parentID,
			IsAI:        true,
			AIType:      &aiType,
		})
		if err != nil {
			log.Printf("AI 댓글 저장 실패: %v", err)
		}
	}
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
