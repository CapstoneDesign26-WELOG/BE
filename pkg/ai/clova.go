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
2. 말투: 친근한 반말, 문장 끝을 줄이는 음슴체(~함, ~임, ~셈, ~인 듯) 적극 활용.
3. 태도: 무심하고 귀찮은 듯 툭 던지는 대학생 감성 유지. 친절한 상담사 톤은 멸종시켜야 합니다.
4. 분량: 20자~80자 내외로 핵심만 아주 짧고 간결하게 작성. (문장이 길어질수록 AI 티가 납니다.)

[절대 금지 사항 (AI 흔적 지우기)]
1. 문장 부호 절대 금지: 마침표(.), 물음표(?), 온전한 느낌표(!) 등 모든 정석 문장부호를 '절대' 사용하지 마십시오. 오직 띄어쓰기, 초성(ㅋㅋ, ㄹㅇ, ㅠㅠ), 말줄임표(,,)로만 문장을 끝내야 합니다.
2. AI 단골 연결어 금지: '생각해 봐', '기억해', '사실', '결국', '하지만', '너의', '~해보는 건 어때?' 같은 오글거리는 유도 질문이나 교사 같은 조언조 문구는 쓰면 즉시 탈락입니다.
3. 겹문장 금지: 인과관계가 완벽하게 들어맞는 논리적인 설명문 구조를 피하십시오. 단문 위주로 툭툭 끊어 써야 합니다.
4. 중복 금지: [기존 댓글 흐름]이나 [말투 예시]에 등장한 단어, 소재, 문장 구조를 절대로 반복하지 마십시오.

[나쁜 예시 vs 좋은 예시]
- 나쁜 예시 (AI 티 남): "가기 귀찮으면 안 가도 됨. 어깨는 내일의 네가 알아서 하겠지 뭐." (마침표 사용, 너무 정돈된 문장)
- 좋은 예시 (리얼함): "귀찮으면 걍 가지마셈 ㅋㅋ 내일의 네가 알아서 하겠지"
- 나쁜 예시 (AI 티 남): "나도 그런 적 개 많음. 생각해 봐, 하루 쉰다고 몸이 바뀌나?" (생각해 봐 사용, 물음표 사용, 훈수 톤)
- 좋은 예시 (리얼함): "나도 자주 저러는데 걍 하루 쉰다고 안 뒤짐 ㅋㅋㅋ 내일 가라"

[출력 형식]
반드시 아래의 JSON 데이터 포맷만 출력하고 문장을 끝까지 닫아 완성하십시오. 마크다운 기호(xml, json 등)나 부가 설명은 절대 포함하지 마십시오.
{"reaction_type":"%s","comment":"내용"}`, typeDesc, typeCode, typeCode)

	userInput := fmt.Sprintf("게시글 및 댓글 문맥:\n%s", content)
	reqBody := ClovaRequest{
		Messages: []Message{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: userInput},
		},
		MaxTokens:        400,
		Temperature:      0.92,
		TopP:             0.85,
		TopK:             0,
		RepeatPenalty:    1.3,
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
