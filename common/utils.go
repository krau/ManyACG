package common

import (
	"regexp"
	"strings"
)

var (
	PixivSourceURLRegexp *regexp.Regexp = regexp.MustCompile(`https://(?:www\.)?pixiv\.net/(?:artworks|i)/(\d+)`)
	AllSourceURLRegexp   *regexp.Regexp = regexp.MustCompile(`https://(?:www\.)?pixiv\.net/(?:artworks|i)/(\d+)`)
)

var (
	SourceURLRegexps []*regexp.Regexp = []*regexp.Regexp{
		PixivSourceURLRegexp,
	}
)

func DownloadFromURL(url string) ([]byte, error) {
	resp, err := Client.R().Get(url)
	if err != nil {
		return nil, err
	}
	return resp.Bytes(), nil
}

func MatchSourceURL(text string) string {
	for _, reg := range SourceURLRegexps {
		if reg.MatchString(text) {
			if reg == PixivSourceURLRegexp {
				url := reg.FindString(text)
				pid := strings.Split(url, "/")[len(strings.Split(url, "/"))-1]
				return "https://www.pixiv.net/artworks/" + pid
			}
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
