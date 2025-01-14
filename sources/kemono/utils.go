package kemono

import (
	"encoding/json"
	"fmt"
	"path"
	"strings"

	"github.com/duke-git/lancet/v2/strutil"
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

var imgSuffixes = []string{".jpg", ".jpeg", ".png", ".gif", ".webp"}

func isImage(kemonoPath string) bool {
	return strutil.HasSuffixAny(path.Ext(strings.Split(kemonoPath, "?")[0]), imgSuffixes)
}
