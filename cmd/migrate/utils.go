package migrate

import (
	"strings"
)

// 清理Unicode转义序列
func sanitizeUnicodeString(s string) string {
	// if !utf8.ValidString(s) {
	// 	s = strings.ToValidUTF8(s, "�")
	// }

	// s = strings.ReplaceAll(s, "\x00", "")

	// controlCharRegex := regexp.MustCompile(`[\x00-\x08\x0B\x0C\x0E-\x1F\x7F]`)
	// s = controlCharRegex.ReplaceAllString(s, "")

	// unicodeEscapeRegex := regexp.MustCompile(`\\u[0-9a-fA-F]{4}`)
	// s = unicodeEscapeRegex.ReplaceAllStringFunc(s, func(match string) string {
	// 	// 可以选择保留或转换这些序列
	// 	// 这里选择保留，但确保格式正确
	// 	return match
	// })

	// return s
	return strings.ReplaceAll(s, "\u0000", "")
}

func sanitizeArtworkData(data *CachedArtworkData) *CachedArtworkData {
	if data == nil {
		return nil
	}
	data.Title = sanitizeUnicodeString(data.Title)
	data.Description = sanitizeUnicodeString(data.Description)
	return data
}
