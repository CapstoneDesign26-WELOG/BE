package ai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

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

type AIContextData struct {
	PostTitle     string
	PostDesc      string
	Comments      []string
	ThreadContext string
	UserComment   string
	TargetAIType  uint
}

func NewClovaClient() *ClovaClient {
	return &ClovaClient{
		APIKey: os.Getenv("CLOVA_STUDIO_API_KEY"),
		URL:    "https://clovastudio.stream.ntruss.com/v1/chat-completions/HCX-003",
	}
}

func (c *ClovaClient) GenerateComment(ctxData AIContextData) (string, error) {
	typeCode, typeDesc := GetAITypeInfo(ctxData.TargetAIType)
	examples := GetRandomExamples(ctxData.TargetAIType, 2)

	systemPrompt := fmt.Sprintf(`# Role
너는 익명 온라인 커뮤니티에서 활동하는 유저다. 
사용자의 고민글에 대해 논문에서 정의된 6가지 공감 유형(Persona Base) 중 현재 너에게 지정된 [%s] 유형의 방식으로 댓글을 작성하라.

# 핵심 원칙: 익명 커뮤니티 리얼리티 극대화
1. 말투: 주로 반말을 사용하며, "~임", "~셈", "~함" 등의 음슴체와 "ㄹㅇ", "ㅇㅈ", "ㅋㅋ", "ㅠㅠ" 등 신조어 및 이모티콘도 일부 활용
2. 비완결성: 마침표(.) 사용을 엄격히 금지하며, 문장을 중간에 끊거나 오타, 띄어쓰기 오류를 허용
3. 다양성: 상담사 같은 정중하고 교과서적 태도는 절대 금지.
4. 긍정적 기대 불일치: AI 특유의 무조건적인 긍정이 아닌, 때로는 단호하고 분석적인 '팩트 폭격'을 섞어 기계적 고정관념을 탈피

# 부여된 공감 모델 (Persona)
- 코드: %s
- 성향: %s

# 리얼리티 규칙 (상황에 맞춰 아래 규칙 중 하나 이상을 유연하게 적용할 것)
- 3단어 이하의 '초단문' 스타일 활용 (예: "힘내라", "ㄹㅇ팩트")
- "아니", "근데", "헐", "엥" 등으로 문장을 시작할 것
- 고민글의 지엽적인 부분에 반응하거나 예상치 못한 각도로 접근할 것

# 제약 사항
- [기존 댓글 흐름]이나 [말투 예시]에 등장한 표현, 단어, 문장 구조를 그대로 복사하거나 반복하여 중복되지 않도록 하라.

# Output Format (JSON Only)
결과는 반드시 아래 JSON 형식으로만 출력하라. 마크다운 기호나 추가 텍스트는 절대 포함하지 말 것.
{"reaction_type": "%s", "comment": "실제 커뮤니티 댓글 내용"}`, typeCode, typeCode, typeDesc, typeCode)

	var sb strings.Builder
	sb.WriteString("# 고민글 (Input Content)\n{\n")
	sb.WriteString(fmt.Sprintf("\t\"post_title\": %q,\n", ctxData.PostTitle))
	sb.WriteString(fmt.Sprintf("\t\"post_content\": %q\n", ctxData.PostDesc))
	sb.WriteString("}\n\n")

	if len(ctxData.Comments) > 0 {
		sb.WriteString("# 맥락 정보: [기존 댓글 흐름]\n")
		for i, cDesc := range ctxData.Comments {
			sb.WriteString(fmt.Sprintf("%d. %s\n", i+1, cDesc))
		}
		sb.WriteString("\n")
	}

	if ctxData.ThreadContext != "" {
		sb.WriteString("# 특별 지시: [대댓글 작성 스레드]\n")
		sb.WriteString("아래의 댓글 스레드 흐름에 자연스럽게 연결되는 답글(대댓글)을 작성하라:\n")
		sb.WriteString(ctxData.ThreadContext)
		sb.WriteString("\n")
	} else if ctxData.UserComment != "" {
		sb.WriteString("# 특별 지시: [사용자 댓글에 대한 답글 작성]\n")
		sb.WriteString(fmt.Sprintf("현재 커뮤니티 유저가 다음과 같은 댓글을 남겼다: %q\n이 유저 댓글에 자연스럽고 위트 있게 반응하는 대댓글을 작성하라.\n\n", ctxData.UserComment))
	}

	if len(examples) > 0 {
		sb.WriteString("# 말투 예시 (참고용 데이터)\n")
		for i, ex := range examples {
			sb.WriteString(fmt.Sprintf("- 예시%d: %s\n", i+1, ex))
		}
	}

	maxRetries := 3
	var aiRes struct {
		ReactionType string `json:"reaction_type"`
		Comment      string `json:"comment"`
	}

	for i := 0; i < maxRetries; i++ {
		resp, err := c.callAPI(systemPrompt, sb.String())
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

		if aiRes.Comment == "" || aiRes.ReactionType != typeCode {
			log.Printf("[Attempt %d] 응답 공백 혹은 성향 코드 불일치 (Expected: %s, Got: %s)", i+1, typeCode, aiRes.ReactionType)
			continue
		}

		return aiRes.Comment, nil
	}

	return "", fmt.Errorf("AI 댓글 생성 최대 재시도 횟수 초과")
}

func (c *ClovaClient) callAPI(systemPrompt, userPrompt string) (string, error) {
	reqBody := ClovaRequest{
		Messages: []Message{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: userPrompt},
		},
		MaxTokens:        300,
		Temperature:      0.92,
		TopP:             0.85,
		TopK:             0,
		RepeatPenalty:    1.3,
		StopBefore:       []string{},
		IncludeAiFilters: false,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", c.URL, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}

	req.Header.Set("Authorization", "Bearer "+c.APIKey)
	req.Header.Set("X-NCP-CLOVASTUDIO-REQUEST-ID", uuid.New().String())
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("상태코드 %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var clovaResp ClovaResponse
	if err := json.Unmarshal(bodyBytes, &clovaResp); err != nil {
		return "", err
	}

	return clovaResp.Result.Message.Content, nil
}
