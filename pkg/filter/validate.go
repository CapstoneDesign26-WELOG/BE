package filter

import "unicode/utf8"

func ValidateLength(text string, maxLen int) bool {
	return utf8.RuneCountInString(text) <= maxLen
}

func TrimToLength(text string, maxLen int) string {
	runes := []rune(text)
	if len(runes) > maxLen {
		return string(runes[:maxLen])
	}
	return text
}
