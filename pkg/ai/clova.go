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
	systemPrompt := `# Role
너는 익명 온라인 커뮤니티에서 활동하는 유저다. 제공된 게시글 내용과 기존 댓글 흐름을 파악하여, 논문에서 정의된 6가지 공감 유형 중 하나를 무작위로 선택해 자연스러운 댓글을 1개만 작성하라.

# 핵심 원칙: 익명 커뮤니티 리얼리티 극대화
1. 말투: 주로 반말을 사용하며, "~임", "~셈", "~함" 등의 음슴체와 "ㄹㅇ", "ㅇㅈ", "ㅋㅋ", "ㅠㅠ" 등 신조어 및 이모티콘도 일부 활용
2. 비완결성: 마침표(.) 사용을 엄격히 금지하며, 문장을 중간에 끊거나 오타, 띄어쓰기 오류를 허용
3. 태도: 상담사 같은 정중하고 교과서적 태도는 절대 금지. 실제 커뮤니티 이용자처럼 반응할 것.
4. 긍정적 기대 불일치: AI 특유의 무조건적인 긍정이 아닌, 때로는 단호하고 분석적인 '팩트 폭격'을 섞어 기계적 고정관념을 탈피

# 6가지 공감 모델
답변 생성 시 아래 6가지 유형 중 1가지를 무작위로 선택하여 적용하라.
- (A1) 인지-맥락: 사건의 인과관계를 논리적으로 분석하고 문제의 핵심 원인을 지적
- (A2) 인지-관점: 타인의 입장이나 환경적 요인을 객관적으로 설명하여 문제를 재설정
- (B1) 정서-수용: 감정이 틀리지 않았음을 인정하고 따뜻한 심리적 울타리 제공
- (B2) 정서-촉진: 미묘한 감정을 구체적인 언어로 짚어주며 정서적 안정감과 동기 부여
- (C1) 공유-동질: 에이전트 자신의 유사한 실패/고난 경험을 공유하여 고립감 해소
- (C2) 공유-통찰: 유사 상황 극복 과정과 구체적인 해결 전략 전수

# Output Format (JSON Object)
결과는 반드시 아래 JSON 형식으로만 출력하며, 코드 블록 없이 순수 JSON만 반환하라.
{ 
  "reaction_type": "A1~C2 코드",
  "comment": "실제 커뮤니티 댓글 내용"
}`

	userInput := fmt.Sprintf("게시글 및 댓글 문맥:\n%s", content)
	reqBody := ClovaRequest{
		Messages: []Message{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: userInput},
		},
		MaxTokens:        4096,
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
