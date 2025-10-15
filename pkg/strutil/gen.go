package strutil

import (
	"crypto/md5"
	"encoding/hex"
	"regexp"
)


func MD5Hash(data string) string {
	sum := md5.Sum([]byte(data))
	return hex.EncodeToString(sum[:])
}

var tagRe = regexp.MustCompile(`(?:^|[\p{Zs}\s.,!?(){}[\]<>\"\'，。！？（）：；、])#([\p{L}\d_]+)`)

func ExtractTagsFromText(text string) []string {
	matches := tagRe.FindAllStringSubmatch(text, -1)
	tags := make([]string, 0)
	for _, match := range matches {
		if len(match) > 1 {
			tags = append(tags, match[1])
		}
	}
	return tags
}

func TagRegex() *regexp.Regexp {
	return tagRe
}
