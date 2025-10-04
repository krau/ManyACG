package strutil

import (
	"fmt"
	"net/url"
	"path"
	"path/filepath"
	"strings"
)

func GetFileExtFromURL(rawurl string) (string, error) {
	parsedUrl, err := url.Parse(rawurl)
	if err != nil {
		return "", err
	}
	if parsedUrl.Path == "" || strings.HasSuffix(parsedUrl.Path, "/") {
		return "", fmt.Errorf("no file name in URL")
	}

	base := path.Base(parsedUrl.Path)
	base, err = url.PathUnescape(base)
	if err != nil {
		return "", err
	}
	ext := filepath.Ext(base)
	if ext == "" {
		return "", fmt.Errorf("no file extension found")
	}
	return ext, nil
}
