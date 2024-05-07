package sources

import (
	. "ManyACG-Bot/logger"
	"ManyACG-Bot/sources/twitter"
	"ManyACG-Bot/types"
	"regexp"
	"strings"
)

var (
	PixivSourceURLRegexp   *regexp.Regexp = regexp.MustCompile(`pixiv\.net/(?:artworks/|i/|member_illust\.php\?(?:[\w=&]*\&|)illust_id=)(\d+)`)
	TwitterSourceURLRegexp *regexp.Regexp = regexp.MustCompile(`(?:twitter|x)\.com/([^/]+)/status/(\d+)`)
)

var (
	SourceURLRegexps map[string]*regexp.Regexp = map[string]*regexp.Regexp{
		string(types.SourceTypePixiv):   PixivSourceURLRegexp,
		string(types.SourceTypeTwitter): TwitterSourceURLRegexp,
	}
)

// MatchSourceURL returns the source URL of the text.
func MatchSourceURL(text string) string {
	text = strings.ReplaceAll(text, "\n", " ")
	for name, reg := range SourceURLRegexps {
		if reg.MatchString(text) {
			switch name {
			case string(types.SourceTypePixiv):
				url := reg.FindString(text)
				pid := strings.Split(url, "/")[len(strings.Split(url, "/"))-1]
				return "https://www.pixiv.net/artworks/" + pid
			case string(types.SourceTypeTwitter):
				url := reg.FindString(text)
				tweetPath := twitter.GetTweetPath(url)
				if tweetPath == "" {
					Logger.Warnf("Matched Twitter URL but failed to get tweet path. URL: %s", url)
					return ""
				}
				return "https://twitter.com/" + tweetPath
			}
			return reg.FindString(text)
		}
	}
	return ""
}

// MatchesSourceURL returns whether the text contains a source URL.
func MatchesSourceURL(text string) bool {
	text = strings.ReplaceAll(text, "\n", " ")
	for _, reg := range SourceURLRegexps {
		if reg.MatchString(text) {
			return true
		}
	}
	return false
}

func GetPixivRegularURL(original string) string {
	photoURL := strings.Replace(original, "img-original", "img-master", 1)
	photoURL = strings.Replace(photoURL, ".jpg", "_master1200.jpg", 1)
	photoURL = strings.Replace(photoURL, ".png", "_master1200.jpg", 1)
	return photoURL
}
