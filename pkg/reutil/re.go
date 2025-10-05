package reutil

import "regexp"

var (
	numberRegexp *regexp.Regexp = regexp.MustCompile(`\d+`)
)

func GetLatestNumberFromString(s string) (string, bool) {
	matches := numberRegexp.FindAllString(s, -1)
	if len(matches) == 0 {
		return "", false
	}
	return matches[len(matches)-1], true
}
