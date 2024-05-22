package common

import (
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

var (
	escapeChars = []string{
		"_", "*", "[", "]", "(", ")", "~", "`", ">", "#", "+", "-", "=", "|", "{", "}", ".", "!",
	}
)

func EscapeMarkdown(text string) string {
	text = strings.ReplaceAll(text, "\\", "\\\\")
	for _, char := range escapeChars {
		text = strings.ReplaceAll(text, char, "\\"+char)
	}
	return text
	// return regexp.MustCompile(`([_\*[\]\(\)~`+"`"+`>#\+\-=|{}\.!])`).ReplaceAllString(text, `\$1`)
}

// 解析字符串为二维数组, 如果以字符串以引号包裹, 则无视分隔符
//
// ParseStringTo2DArray("1,2,3;4,5,6", ",", ";") => [][]string{{"1", "2", "3"}, {"4", "5", "6"}}
//
// ParseStringTo2DArray("1,2,3;\"4,5,6\"", ",", ";") => [][]string{{"1", "2", "3"}, {"4,5,6"}}
func ParseStringTo2DArray(str, sep, sep2 string) ([][]string, error) {
	var result [][]string
	if str == "" {
		return result, nil
	}

	var row []string
	var inQuote bool
	var builder strings.Builder

	for _, c := range str {
		if inQuote {
			if c == '"' || c == '\'' {
				inQuote = false
			} else {
				if _, err := builder.WriteRune(c); err != nil {
					return nil, err
				}
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
				if _, err := builder.WriteRune(c); err != nil {
					return nil, err
				}
			}
		}
	}

	if builder.Len() > 0 {
		row = append(row, builder.String())
	}
	if len(row) > 0 {
		result = append(result, row)
	}

	return result, nil
}
