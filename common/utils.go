package common

func DownloadFromURL(url string) ([]byte, error) {
	resp, err := Cilent.R().Get(url)
	if err != nil {
		return nil, err
	}
	return resp.Bytes(), nil
}
