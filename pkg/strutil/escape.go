package strutil

import (
	"regexp"
	"strings"
	"unicode"

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
		if unicode.IsControl(r) || unicode.IsSpace(r) {
			return '_'
		}
		if validator.IsPrintable(string(r)) {
			return r
		}
		return '_'
	}, fname)

	re := regexp.MustCompile(`_+`)
	fname = re.ReplaceAllString(fname, "_")

	fname = strings.Trim(fname, "_ ")

	return fname
}