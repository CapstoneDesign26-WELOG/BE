package filter

import "github.com/cloudflare/ahocorasick"

var badWords = []string{
	"씨발", "개새끼", "지랄", "병신",
}

var matcher *ahocorasick.Matcher

func init() {
	matcher = ahocorasick.NewStringMatcher(badWords)
}

func ContainsProfanity(text string) bool {
	hits := matcher.Match([]byte(text))
	return len(hits) > 0
}
