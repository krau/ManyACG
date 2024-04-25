package common

import (
	"strings"
)

func DownloadFromURL(url string) ([]byte, error) {
	resp, err := Client.R().Get(url)
	if err != nil {
		return nil, err
	}
	return resp.Bytes(), nil
}

func ReplaceFileNameInvalidChar(fileName string) string {
	return strings.NewReplacer(
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
	).Replace(fileName)
}
