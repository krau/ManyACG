package kemono

import (
	"context"
	"encoding/json"
	"fmt"
	"path"
	"strings"

	"github.com/duke-git/lancet/v2/strutil"
)

func getPostPath(sourceURL string) string {
	return strings.TrimPrefix(sourceURL, kemonoDomainBase)
}

func (k *Kemono) getAuthorProfile(ctx context.Context, service, creatorId string) (*KemonoCreatorProfileResp, error) {
	apiURL := apiBaseURL + fmt.Sprintf("/%s/user/%s/profile", service, creatorId)
	resp, err := k.reqClient.R().SetContext(ctx).Get(apiURL)
	if err != nil {
		return nil, err
	}
	var kemonoResp KemonoCreatorProfileResp
	if err := json.Unmarshal(resp.Bytes(), &kemonoResp); err != nil {
		return nil, err
	}
	return &kemonoResp, nil
}

var imgSuffixes = []string{".jpg", ".jpeg", ".png", ".webp"}

func isImage(kemonoPath string) bool {
	return strutil.HasSuffixAny(path.Ext(strings.Split(kemonoPath, "?")[0]), imgSuffixes)
}
