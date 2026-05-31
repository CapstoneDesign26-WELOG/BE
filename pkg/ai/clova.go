package ai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/google/uuid"
)

type ClovaClient struct {
	APIKey string
	URL    string
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ClovaRequest struct {
	Messages         []Message `json:"messages"`
	TopP             float64   `json:"topP"`
	TopK             int       `json:"topK"`
	MaxTokens        int       `json:"maxTokens"`
	Temperature      float64   `json:"temperature"`
	RepeatPenalty    float64   `json:"repeatPenalty"`
	StopBefore       []string  `json:"stopBefore"`
	IncludeAiFilters bool      `json:"includeAiFilters"`
}

type ClovaResponse struct {
	Result struct {
		Message Message `json:"message"`
	} `json:"result"`
}

func NewClovaClient() *ClovaClient {
	return &ClovaClient{
		APIKey: os.Getenv("CLOVA_STUDIO_API_KEY"),
		URL:    "https://clovastudio.stream.ntruss.com/v1/chat-completions/HCX-003",
	}
}

func (c *ClovaClient) GetSingleAIComment(content string) (string, error) {
	systemPrompt := `Role: 익명 커뮤니티 유저. 문맥에 맞는 리얼한 댓글 1개 작성.
Rules:
- 분량: 50자~200자 사이로 다양하게 작성 (때로는 짧게, 때로는 썰을 풀듯이 길게).
- 말투: 반말, 음슴체(~임, ~함), 신조어(ㄹㅇ, ㅋㅋ) 권장. 마침표 금지. 오타/띄어쓰기 파괴 권장.
- 태도: 상담사 말투 절대 금지. 날것의 커뮤니티 감성으로 팩폭/공감/분석/비완결성 유지.
- 공감모델: 아래 6개 중 무작위 1개 선택

(A1) 인지-맥락: 사건의 인과관계를 논리적으로 분석하고 문제의 핵심 원인을 지적
(A2) 인지-관점: 타인의 입장이나 환경적 요인을 객관적으로 설명하여 문제를 재설정
(B1) 정서-수용: 감정이 틀리지 않았음을 인정하고 따뜻한 심리적 울타리 제공
(B2) 정서-촉진: 미묘한 감정을 구체적인 언어로 짚어주며 정서적 안정감과 동기 부여
(C1) 공유-동질: 에이전트 자신의 유사한 실패/고난 경험을 공유하여 고립감 해소
(C2) 공유-통찰: 유사 상황 극복 과정과 구체적인 해결 전략 전수

Output: JSON Only. (반드시 JSON 형식을 끝까지 닫아서 완성할 것)
{"reaction_type":"A1~C2 코드","comment":"내용"}`

	userInput := fmt.Sprintf("게시글 및 댓글 문맥:\n%s", content)
	reqBody := ClovaRequest{
		Messages: []Message{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: userInput},
		},
		MaxTokens:        500,
		Temperature:      0.8,
		TopP:             0.8,
		TopK:             0,
		RepeatPenalty:    1.2,
		StopBefore:       []string{},
		IncludeAiFilters: false,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("JSON 마샬링 실패: %v", err)
	}

	req, err := http.NewRequest("POST", c.URL, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("요청 생성 실패: %v", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.APIKey)
	req.Header.Set("X-NCP-CLOVASTUDIO-REQUEST-ID", uuid.New().String())
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("API 호출 실패: %v", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("응답 본문 읽기 실패: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API 응답 에러 (상태코드: %d): %s", resp.StatusCode, string(bodyBytes))
	}

	var clovaResp ClovaResponse
	if err := json.Unmarshal(bodyBytes, &clovaResp); err != nil {
		return "", fmt.Errorf("응답 파싱 실패: %v\n원본 응답: %s", err, string(bodyBytes))
	}

	return clovaResp.Result.Message.Content, nil
}
