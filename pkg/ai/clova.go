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

func (c *ClovaClient) GetSingleAIComment(content string, reactionType uint) (string, error) {
	typeCode, typeDesc := GetAITypeInfo(reactionType)

	systemPrompt := fmt.Sprintf(`[역할]
대학생 익명 커뮤니티(에브리타임)의 실제 유저로 빙의하여 문맥에 맞는 리얼한 댓글 1개를 작성하십시오.

[성향 및 규칙]
1. 성향: %s (%s)
2. 말투: 친근한 반말, 문장 끝을 줄이는 음슴체(~함, ~임), 자음 초성(ㅋㅋ, ㄹㅇ) 적극 활용. 마침표(.) 사용 금지.
3. 태도: AI 상담사 같은 친절함이나 형식적인 위로 조는 절대 금지. 날것의 대학생 감성 유지.
4. 분량: 20자~150자 내외로 핵심만 짧고 간결하게 작성.

[제약사항]
1. 중복 금지: [기존 댓글 흐름]이나 [말투 예시]에 등장한 단어, 소재, 문장 구조를 절대로 반복하지 마십시오. 완전히 새로운 시각으로 반응해야 합니다.
2. 예시 활용: [말투 예시] 항목은 오직 '말투와 분위기'만 모방하고, 예시의 '내용(토픽)'은 절대 베끼지 마십시오.
3. 흐름 유지: [특별 지시: 대댓글 작성]이 제공된 경우, 해당 스레드에 자연스럽게 이어지도록 대댓글 형태로 작성하십시오.

[출력 형식]
반드시 아래의 JSON 데이터 포맷만 출력하고 문장을 끝까지 닫아 완성하십시오. 마크다운 기호(xml, json 등)나 부가 설명은 절대 포함하지 마십시오.
{"reaction_type":"%s","comment":"내용"}`, typeDesc, typeCode, typeCode)

	userInput := fmt.Sprintf("게시글 및 댓글 문맥:\n%s", content)
	reqBody := ClovaRequest{
		Messages: []Message{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: userInput},
		},
		MaxTokens:        500,
		Temperature:      0.82,
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
