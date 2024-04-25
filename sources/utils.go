package sources

import (
	"ManyACG-Bot/sources/twitter"
	"ManyACG-Bot/types"
	"regexp"
	"strings"
)

var (
	PixivSourceURLRegexp   *regexp.Regexp = regexp.MustCompile(`https://(?:www\.)?pixiv\.net/(?:artworks|i)/(\d+)`)
	TwitterSourceURLRegexp *regexp.Regexp = regexp.MustCompile(`https://(?:twitter|x)\.com/([^/]+)/status/(\d+)`)
	AllSourceURLRegexp     *regexp.Regexp = regexp.MustCompile(`https://(?:www\.)?(pixiv\.net/(?:artworks|i)/\d+|(?:twitter|x)\.com/[^/]+/status/\d+)`)
)

var (
	SourceURLRegexps map[string]*regexp.Regexp = map[string]*regexp.Regexp{
		string(types.SourceTypePixiv):   PixivSourceURLRegexp,
		string(types.SourceTypeTwitter): TwitterSourceURLRegexp,
	}
)

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
				tweetPath, err := twitter.GetTweetPath(text)
				if err != nil {
					return ""
				}
				return "https://twitter.com/" + tweetPath
			}
			return reg.FindString(text)
		}
	}
	return ""
}

func GetPixivRegularURL(original string) string {
	photoURL := strings.Replace(original, "img-original", "img-master", 1)
	photoURL = strings.Replace(photoURL, ".jpg", "_master1200.jpg", 1)
	photoURL = strings.Replace(photoURL, ".png", "_master1200.jpg", 1)
	return photoURL
}
