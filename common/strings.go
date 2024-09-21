package common

import (
	"ManyACG/config"
	. "ManyACG/logger"
	"html"
	"math/rand"
	"net/url"
	"regexp"
	"strings"
)

var fileNameEscaper = strings.NewReplacer(
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
)

func EscapeFileName(fileName string) string {
	return fileNameEscaper.Replace(fileName)
}

var markdownRe = regexp.MustCompile("([" + regexp.QuoteMeta(`\_*[]()~`+"`"+`>#+-=|{}.!`) + "])")

func EscapeMarkdown(text string) string {
	return markdownRe.ReplaceAllString(text, "\\$1")
}

func EscapeHTML(text string) string {
	return html.EscapeString(text)
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

func ExtractTagsFromText(text string) []string {
	// 正则表达式匹配标签，\p{Zs} 匹配所有空格字符，包括全角空格
	re := regexp.MustCompile(`(?:^|[\p{Zs}\s.,!?(){}[\]<>\"\'，。！？（）：；、])#([\p{L}\d_]+)`)

	// 找到所有匹配的标签
	matches := re.FindAllStringSubmatch(text, -1)

	// 提取标签内容，不包含 # 号
	tags := make([]string, 0)
	for _, match := range matches {
		if len(match) > 1 {
			tags = append(tags, match[1])
		}
	}
	return tags
}

func ApplyPathRule(path string) string {
	for _, rule := range config.Cfg.API.PathRules {
		if strings.HasPrefix(path, rule.Path) {
			parsedUrl, err := url.JoinPath(rule.JoinPrefix, strings.TrimPrefix(path, rule.TrimPrefix))
			if err != nil {
				Logger.Errorf("Failed to parse url: %s", err)
				return rule.JoinPrefix + strings.TrimPrefix(path, rule.TrimPrefix)
			}
			return parsedUrl
		}
	}
	return path
}
