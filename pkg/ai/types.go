package ai

type AITypeInfo struct {
	Code        string
	Description string
}

var AITypeMap = map[uint]AITypeInfo{
	1: {"A1", "인지-맥락: 사건의 인과관계를 논리적으로 분석하고 문제의 핵심 원인을 지적"},
	2: {"A2", "인지-관점: 타인의 입장이나 환경적 요인을 객관적으로 설명하여 문제를 재설정"},
	3: {"B1", "정서-수용: 감정이 틀리지 않았음을 인정하고 따뜻한 심리적 울타리 제공"},
	4: {"B2", "정서-촉진: 미묘한 감정을 구체적인 언어로 짚어주며 정서적 안정감과 동기 부여"},
	5: {"C1", "공유-동질: 자신의 유사한 실패 경험을 공유하여 고립감 해소"},
	6: {"C2", "공유-통찰: 유사 상황 극복 과정과 구체적인 해결 전략 전수"},
}

func GetAITypeInfo(reactionType uint) (string, string) {
	info, ok := AITypeMap[reactionType]
	if !ok {
		info = AITypeMap[1]
	}
	return info.Code, info.Description
}
