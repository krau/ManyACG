package kemono

import (
	"encoding/json"
	"fmt"
	"strings"
)

func getPostPath(sourceURL string) string {
	return strings.TrimPrefix(sourceURL, "https://kemono.su")
}

func getAuthorProfile(service, creatorId string) (*KemonoCreatorProfileResp, error) {
	apiURL := apiBaseURL + fmt.Sprintf("/%s/user/%s/profile", service, creatorId)
	resp, err := reqClient.R().Get(apiURL)
	if err != nil {
		return nil, err
	}
	var kemonoResp KemonoCreatorProfileResp
	if err := json.Unmarshal(resp.Bytes(), &kemonoResp); err != nil {
		return nil, err
	}
	return &kemonoResp, nil
}

func isImage(path string) bool {
	return strings.HasSuffix(path, ".jpg") || strings.HasSuffix(path, ".jpeg") || strings.HasSuffix(path, ".png")
}
