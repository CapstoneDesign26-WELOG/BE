package post

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"strings"
	"time"
	"welog/internal/comment"
	"welog/internal/model"
	"welog/internal/notification"
	"welog/internal/user"
	"welog/pkg/ai"
)

type PostService struct {
	repo                *PostRepository
	userService         *user.UserService
	commentService      *comment.CommentService
	clovaClient         *ai.ClovaClient
	notificationService *notification.NotificationService
}

func NewPostService(repo *PostRepository, userService *user.UserService, commentService *comment.CommentService, clovaClient *ai.ClovaClient, notificationService *notification.NotificationService) *PostService {
	return &PostService{
		repo:                repo,
		userService:         userService,
		commentService:      commentService,
		clovaClient:         clovaClient,
		notificationService: notificationService,
	}
}

func (s *PostService) CreatePost(userID uint, title, description string, postType uint) (*model.Post, error) {
	var tokenCost uint = 1
	if err := s.userService.ConsumeToken(userID, tokenCost); err != nil {
		return nil, err
	}

	aiCount := uint(rand.Intn(3) + 3)
	post := &model.Post{
		UserID:      userID,
		Title:       title,
		Description: description,
		Type:        postType,
		Count:       aiCount,
	}

	if err := s.repo.Create(post); err != nil {
		// 실패시 토큰 되돌려주는 로직 추가 예정
		return nil, err
	}

	delay := getRandomDelay(1, 3)
	time.AfterFunc(delay, func() {
		s.handleAICommentStep(userID, post.ID, aiCount)
	})

	return post, nil
}

/*
func (s *PostService) handleAIComments(userID, postID, parentID uint, content string) {
	resp, err := s.clovaClient.GetSingleAIComment(postID, content)
	if err != nil {
		log.Printf("AI 댓글 생성 실패 (PostID: %d): %v", postID, err)
		s.notificationService.Notify(userID, fmt.Sprintf(`{"type": "AI_COMMENT_FAILED", "post_id": %d}`, postID))
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
		s.notificationService.Notify(userID, fmt.Sprintf(`{"type": "AI_COMMENT_FAILED", "post_id": %d}`, postID))
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

	s.notificationService.Notify(userID, fmt.Sprintf(`{"type": "AI_COMMENT_COMPLETE", "post_id": %d}`, postID))
}
*/

func getRandomDelay(st, en int) time.Duration {
	return time.Duration(rand.Intn(en-st+1)+st) * time.Minute
}

func (s *PostService) handleAICommentStep(userID, postID, remaining uint) {
	if remaining == 0 {
		s.notificationService.Notify(userID, fmt.Sprintf(`{"type": "AI_COMMENT_COMPLETE", "post_id": %d}`, postID))
		return
	}

	post, err := s.repo.FindByID(postID)
	if err != nil {
		log.Printf("AI 댓글 작업 중 게시글 조회 실패: %v", err)
		return
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("[글 제목]: %s\n[글 내용]: %s\n\n[기존 댓글 흐름]\n", post.Title, post.Description))
	for _, c := range post.Comments {
		sb.WriteString(fmt.Sprintf("[닉네임: %s] - %s\n", c.User.Nickname, c.Description))
	}

	var rootComments []model.Comment
	var parentID *uint
	if len(post.Comments) > 0 && rand.Float32() < 0.3 {
		for _, c := range post.Comments {
			if c.ParentID == nil {
				rootComments = append(rootComments, c)
			}
		}

		if len(rootComments) > 0 {
			selectedRoot := rootComments[rand.Intn(len(rootComments))]
			parentID = &selectedRoot.ID

			var threadBuilder strings.Builder
			threadBuilder.WriteString(fmt.Sprintf("-%s (원문)\n", selectedRoot.Description))

			for _, c := range post.Comments {
				if c.ParentID != nil && *c.ParentID == selectedRoot.ID {
					threadBuilder.WriteString(fmt.Sprintf(" ㄴ %s\n", c.Description))
				}
			}

			sb.WriteString(fmt.Sprintf("\n(특별 지시: 위 댓글들 중 다음 스레드에 자연스럽게 이어지는 답글 형태로 작성해줘:\n%s)", threadBuilder.String()))
		}
	}

	resp, err := s.clovaClient.GetSingleAIComment(sb.String())
	if err != nil {
		log.Printf("AI API 호출 실패: %v", err)
		return
	}

	var aiRes struct {
		ReactionType string `json:"reaction_type"`
		Comment      string `json:"comment"`
	}

	cleaned := strings.TrimSpace(resp)
	cleaned = strings.TrimPrefix(cleaned, "```json")
	cleaned = strings.TrimPrefix(cleaned, "```")
	cleaned = strings.TrimSuffix(cleaned, "```")
	cleaned = strings.TrimSpace(cleaned)

	if err := json.Unmarshal([]byte(cleaned), &aiRes); err != nil {
		log.Printf("AI 응답 파싱 실패: %v", err)
		return
	}

	typeMap := map[string]uint{
		"A1": 1, "A2": 2, "B1": 3, "B2": 4, "C1": 5, "C2": 6,
	}
	aiType := typeMap[aiRes.ReactionType]
	if aiType == 0 {
		aiType = uint(rand.Intn(6) + 1)
	}

	_, err = s.commentService.CreateComment(comment.CreateCommentParams{
		UserID:      1,
		PostID:      postID,
		Description: aiRes.Comment,
		ParentID:    parentID,
		IsAI:        true,
		AIType:      &aiType,
	})

	if err != nil {
		log.Printf("AI 댓글 저장 실패: %v", err)
		return
	}

	nextDelay := getRandomDelay(1, 3)
	time.AfterFunc(nextDelay, func() {
		s.handleAICommentStep(userID, postID, remaining-1)
	})
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

	reqUser, err := s.userService.GetUser(userID)
	if err != nil {
		return err
	}

	if post.UserID != userID && reqUser.Role != "ADMIN" {
		return errors.New("게시글을 삭제할 권한이 없습니다")
	}

	return s.repo.Delete(postID)
}
