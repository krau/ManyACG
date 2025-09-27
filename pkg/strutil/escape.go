package strutil

import (
	"html"
	"regexp"
	"strings"

	"github.com/duke-git/lancet/v2/strutil"
	"github.com/duke-git/lancet/v2/validator"
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
	"'", "_",
	"`", "_",
	"\t", "_",
	"\r", "_",
	"\n", "_",
)

func SanitizeFileName(fileName string) string {
	fname := strutil.RemoveWhiteSpace(fileNameReplacer.Replace(fileName), true)
	fname = strings.Map(func(r rune) rune {
		if r < 0x20 || r == 0x7F {
			return '_'
		}
		if validator.IsPrintable(string(r)) {
			return r
		}
		return '_'
	}, fname)
	return fname
}

var markdownRe = regexp.MustCompile("([" + regexp.QuoteMeta(`\_*[]()~`+"`"+`>#+-=|{}.!`) + "])")

func EscapeMarkdown(text string) string {
	return markdownRe.ReplaceAllString(text, "\\$1")
}

func EscapeHTML(text string) string {
	return html.EscapeString(text)
}
