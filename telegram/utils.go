package telegram

import (
	"regexp"
	"strings"
)

func escapeMarkdown(text string) string {
	escapeChars := `\_*[]()~` + "`" + ">#+-=|{}.!"
	re := regexp.MustCompile("([" + regexp.QuoteMeta(escapeChars) + "])")
	return re.ReplaceAllString(text, "\\$1")
}

func replaceChars(input string, oldChars []string, newChar string) string {
	for _, char := range oldChars {
		input = strings.ReplaceAll(input, char, newChar)
	}
	return input
}
