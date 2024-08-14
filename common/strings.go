package common

import (
	"math/rand"
	"regexp"
	"strings"
)

func ReplaceFileNameInvalidChar(fileName string) string {
	return strings.NewReplacer(
		" ", "_",
		"/", "_",
		"\\", "_",
		":", "_",
		"*", "_",
		"?", "_",
		"\"", "_",
		"<", "_",
		">", "_",
		"|", "_",
		"%", "_",
		"#", "_",
		"+", "_",
	).Replace(fileName)
}

func EscapeMarkdown(text string) string {
	escapeChars := `\_*[]()~` + "`" + `>#+-=|{}.!`
	re := regexp.MustCompile("([" + regexp.QuoteMeta(escapeChars) + "])")
	return re.ReplaceAllString(text, "\\$1")
}

func EscapeHTML(text string) string {
	return strings.NewReplacer(
		"<", "&lt;",
		">", "&gt;",
		"&", "&amp;",
	).Replace(text)
}

// 解析字符串为二维数组, 如果以字符串以引号包裹, 则无视分隔符
//
// ParseStringTo2DArray("1,2,3;4,5,6", ",", ";") => [][]string{{"1", "2", "3"}, {"4", "5", "6"}}
//
// ParseStringTo2DArray("1,2,3;\"4,5,6\"", ",", ";") => [][]string{{"1", "2", "3"}, {"4,5,6"}}
func ParseStringTo2DArray(str, sep, sep2 string) [][]string {
	var result [][]string
	if str == "" {
		return result
	}

	var row []string
	var inQuote bool
	var builder strings.Builder

	for _, c := range str {
		if inQuote {
			if c == '"' || c == '\'' {
				inQuote = false
			} else {
				builder.WriteRune(c)
			}
		} else {
			if c == '"' || c == '\'' {
				inQuote = true
			} else if string(c) == sep {
				row = append(row, builder.String())
				builder.Reset()
			} else if string(c) == sep2 {
				row = append(row, builder.String())
				result = append(result, row)
				row = nil
				builder.Reset()
			} else {
				builder.WriteRune(c)
			}
		}
	}

	if builder.Len() > 0 {
		row = append(row, builder.String())
	}
	if len(row) > 0 {
		result = append(result, row)
	}

	return result
}

// 去除字符串切片中的重复元素
func RemoveDuplicateStringSlice(s []string) []string {
	encountered := map[string]bool{}
	result := []string{}
	for _, str := range s {
		if !encountered[str] {
			encountered[str] = true
			result = append(result, str)
		}
	}
	return result
}

const defaultCharset = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

func GenerateRandomString(length int, charset ...string) string {
	var letters string
	if len(charset) > 0 {
		letters = strings.Join(charset, "")
	} else {
		letters = defaultCharset
	}
	b := make([]byte, length)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}
