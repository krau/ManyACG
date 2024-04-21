package common

import (
	"regexp"
	"strings"
)

func DownloadFromURL(url string) ([]byte, error) {
	resp, err := Cilent.R().Get(url)
	if err != nil {
		return nil, err
	}
	return resp.Bytes(), nil
}

func MatchSourceURL(text string) string {
	return regexp.MustCompile(`https://www.pixiv.net/artworks/(\d+)`).FindString(text)
}

func GetPixivRegularURL(original string) string {
	photoURL := strings.Replace(original, "img-original", "img-master", 1)
	photoURL = strings.Replace(photoURL, ".jpg", "_master1200.jpg", 1)
	photoURL = strings.Replace(photoURL, ".png", "_master1200.jpg", 1)
	return photoURL
}
