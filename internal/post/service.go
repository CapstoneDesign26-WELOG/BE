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
		s.userService.RefundToken(userID, tokenCost)
		return nil, err
	}

	delay := getRandomDelay(1, 3)
	time.AfterFunc(delay, func() {
		s.handleAICommentStep(userID, post.ID, aiCount)
	})

	return post, nil
}

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

			count := 0
			for i := len(post.Comments) - 1; i >= 0; i-- {
				c := post.Comments[i]
				if c.ParentID != nil && *c.ParentID == selectedRoot.ID {
					threadBuilder.WriteString(fmt.Sprintf(" ㄴ %s\n", c.Description))
					count++

					if count >= 4 {
						break
					}
				}
			}

			sb.WriteString(fmt.Sprintf("\n(특별 지시: 위 댓글들 중 다음 스레드에 자연스럽게 이어지는 답글 형태로 작성해줘:\n%s)", threadBuilder.String()))
		}
	}

	targetAIType := s.getWeightedAIType(userID)
	typeCode, _ := ai.GetAITypeInfo(targetAIType)

	var resp string
	var aiRes struct {
		ReactionType string `json:"reaction_type"`
		Comment      string `json:"comment"`
	}

	maxRetries := 3
	success := false

	for i := 0; i < maxRetries; i++ {
		resp, err = s.clovaClient.GetSingleAIComment(sb.String(), targetAIType)
		if err != nil {
			log.Printf("[Attempt %d] AI API 호출 실패: %v", i+1, err)
			continue
		}

		cleaned := strings.TrimSpace(resp)
		start := strings.Index(cleaned, "{")
		end := strings.LastIndex(cleaned, "}")
		if start != -1 && end != -1 && end > start {
			cleaned = cleaned[start : end+1]
		}

		if err := json.Unmarshal([]byte(cleaned), &aiRes); err != nil {
			log.Printf("[Attempt %d] JSON 파싱 실패: %v", i+1, err)
			continue
		}

		if aiRes.Comment == "" {
			log.Printf("[Attempt %d] AI 응답 내용이 비어있음", i+1)
			continue
		}

		if aiRes.ReactionType != typeCode {
			log.Printf("[Attempt %d] 성향 코드 불일치 (Expected: %s, Got: %s)", i+1, typeCode, aiRes.ReactionType)
			continue
		}

		success = true
		break
	}

	if !success {
		log.Printf("AI 댓글 생성 최종 실패 (ID: %d), 다음 스텝으로 시도합니다.", postID)
		time.AfterFunc(getRandomDelay(1, 3), func() {
			s.handleAICommentStep(userID, postID, remaining-1)
		})
		return
	}

	_, err = s.commentService.CreateComment(comment.CreateCommentParams{
		UserID:      1,
		PostID:      postID,
		Description: aiRes.Comment,
		ParentID:    parentID,
		IsAI:        true,
		AIType:      &targetAIType,
	})

	if err != nil {
		log.Printf("AI 댓글 저장 실패: %v", err)
		time.AfterFunc(getRandomDelay(1, 3), func() {
			s.handleAICommentStep(userID, postID, remaining-1)
		})
		return
	}

	nextDelay := getRandomDelay(1, 3)
	time.AfterFunc(nextDelay, func() {
		s.handleAICommentStep(userID, postID, remaining-1)
	})
}

func (s *PostService) GetPublicPosts(page, limit int) ([]model.Post, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	offset := (page - 1) * limit

	return s.repo.FindAllPublic(offset, limit)
}

func (s *PostService) GetPrivatePosts(userID uint, page, limit int) ([]model.Post, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	offset := (page - 1) * limit

	return s.repo.FindAllPrivate(userID, offset, limit)
}

func (s *PostService) GetPostDetails(postID uint) (*model.Post, error) {
	return s.repo.FindByID(postID)
}

func (s *PostService) GetPostsByUserID(userID uint) ([]model.Post, error) {
	return s.repo.FindAllByUserID(userID)
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

func (s *PostService) getWeightedAIType(userID uint) uint {
	weights := map[uint]int{1: 1, 2: 1, 3: 1, 4: 1, 5: 1, 6: 1}
	if userPrefs, err := s.userService.GetUserPreferences(userID); err == nil {
		for aiType, score := range userPrefs {
			if score > 0 {
				weights[aiType] += score
			}
		}
	}

	totalWeight := 0
	for _, w := range weights {
		totalWeight += w
	}

	if u, err := s.userService.GetUser(userID); err == nil && u.AIPreference != 0 {
		boost := int(float64(totalWeight) * 0.5)
		if boost < 1 {
			boost = 1
		}

		switch u.AIPreference {
		case 1: // 현실조언형: A1(1), A2(2), C2(6)
			weights[1] += boost
			weights[2] += boost
			weights[6] += boost
		case 2: // 감정위로형: B1(3), B2(4), C1(5)
			weights[3] += boost
			weights[4] += boost
			weights[5] += boost
		}

		totalWeight = 0
		for _, w := range weights {
			totalWeight += w
		}
	}

	r := rand.Intn(totalWeight)
	cumulative := 0
	for aiType := uint(1); aiType <= 6; aiType++ {
		cumulative += weights[aiType]
		if r < cumulative {
			return aiType
		}
	}
	return 1
}
