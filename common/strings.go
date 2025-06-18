package common

import (
	"crypto/md5"
	"encoding/hex"
	"html"
	"math/rand"
	"net/url"
	"regexp"
	"strings"

	"github.com/duke-git/lancet/v2/validator"
	"github.com/krau/ManyACG/config"
	"github.com/krau/ManyACG/types"
)

var fileNameReplacer = strings.NewReplacer(
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

func SanitizeFileName(fileName string) string {
	fname := strings.TrimSpace(fileNameReplacer.Replace(fileName))
	fname = strings.Map(func(r rune) rune {
		if r < 0x20 || r == 0x7F {
			return '_'
		}
		switch r {
		// invalid characters
		case '/', '\\',
			':', '*', '?', '"', '<', '>', '|':
			return '_'
		// empty
		case ' ', '\t', '\r', '\n':
			return '_'
		}
		if validator.IsPrintable(string(r)) {
			return r
		}
		return '_'
	}, fname)
	return strings.Join(strings.FieldsFunc(fname, func(r rune) bool {
		return r == '_' || r == '-' || r == '.' || r == ' '
	}), "_")
}

var markdownRe = regexp.MustCompile("([" + regexp.QuoteMeta(`\_*[]()~`+"`"+`>#+-=|{}.!`) + "])")

func EscapeMarkdown(text string) string {
	return markdownRe.ReplaceAllString(text, "\\$1")
}

func EscapeHTML(text string) string {
	return html.EscapeString(text)
}

func MD5Hash(data string) string {
	sum := md5.Sum([]byte(data))
	return hex.EncodeToString(sum[:])
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

var TagRe = regexp.MustCompile(`(?:^|[\p{Zs}\s.,!?(){}[\]<>\"\'，。！？（）：；、])#([\p{L}\d_]+)`)

func ExtractTagsFromText(text string) []string {
	matches := TagRe.FindAllStringSubmatch(text, -1)
	tags := make([]string, 0)
	for _, match := range matches {
		if len(match) > 1 {
			tags = append(tags, match[1])
		}
	}
	return tags
}

// 在api返回中重写图片路径, 用于拼接图片直链
//
// 例: rule.Type = "alist", rule.Path = "/pictures/", rule.JoinPrefix = "https://example.com/pictures/", rule.TrimPrefix = "/pictures/"
//
// -> https://example.com/pictures/1234567890abcdef.jpg
//
// 如果图片路径以rule.Path开头, 且rule.StorageType为空或与图片存储类型匹配, 则将图片路径转换为rule.JoinPrefix + 图片路径去掉rule.TrimPrefix的部分
//
// 否则返回原图片路径
func ApplyApiStoragePathRule(detail *types.StorageDetail) string {
	for _, rule := range config.Cfg.API.PathRules {
		if strings.HasPrefix(detail.Path, rule.Path) && (rule.StorageType == "" || rule.StorageType == string(detail.Type)) {
			parsedUrl, err := url.JoinPath(rule.JoinPrefix, strings.TrimPrefix(detail.Path, rule.TrimPrefix))
			if err != nil {
				Logger.Warnf("Failed to join path: %s", err)
				return detail.Path
			}
			return parsedUrl
		}
	}
	return detail.Path
}
