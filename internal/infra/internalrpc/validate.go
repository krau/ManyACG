package internalrpc

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"mime"
	"net"
	"net/url"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/krau/ManyACG/internal/common/httpclient"
	"github.com/krau/ManyACG/internal/infra/config/runtimecfg"
	"github.com/krau/ManyACG/internal/infra/internalrpc/internalpb"
	"github.com/krau/ManyACG/internal/infra/source"
	"github.com/krau/ManyACG/internal/model/dto"
	"github.com/krau/ManyACG/internal/shared"
)

// These bounds accommodate large galleries and descriptions while bounding edge-controlled metadata.
const (
	maxResultMedia       = 4096
	maxResultIndex       = 65535 // Kemono video indices can refer to a sparse attachment list.
	maxResultDimension   = 100000
	maxResultURLBytes    = 8192
	maxResultTitleBytes  = 16 << 10
	maxResultTextBytes   = 1 << 20
	maxResultArtistBytes = 4 << 10
	maxResultTags        = 4096
	maxResultTagBytes    = 4 << 10
	maxResultFrames      = 65536
	maxResultFrameBytes  = 1024
	maxResultMIMEBytes   = 256
	maxResultDurationMS  = 7 * 24 * 60 * 60 * 1000
)

type artworkResultIdentity struct {
	source shared.SourceType
	id     string
}

func validateArtworkResult(
	ctx context.Context,
	requestedURL string,
	message *internalpb.FetchedArtwork,
	namers map[shared.SourceType]source.ArtworkNamer,
	cfg runtimecfg.SourceConfig,
) (*dto.FetchedArtwork, error) {
	if message == nil {
		return nil, errors.New("internal agent returned a null artwork")
	}
	requested, _, err := identifyArtworkResult(requestedURL, namers)
	if err != nil {
		return nil, fmt.Errorf("invalid requested artwork URL: %w", err)
	}
	returned, canonicalURL, err := identifyArtworkResult(message.SourceUrl, namers)
	if err != nil {
		return nil, fmt.Errorf("internal agent returned an invalid source_url: %w", err)
	}
	if returned != requested || message.SourceType != string(requested.source) {
		return nil, errors.New("internal agent returned a different artwork or source")
	}
	if !boundedResultString(message.Title, maxResultTitleBytes) || !boundedResultString(message.Description, maxResultTextBytes) {
		return nil, errors.New("internal agent returned invalid artwork text")
	}
	artist := message.Artist
	if artist == nil || artist.Type != string(requested.source) || !safeResultArtistUID(artist.Uid) {
		return nil, errors.New("internal agent returned an invalid artist identity or source")
	}
	if !boundedResultString(artist.Name, maxResultArtistBytes) || !boundedResultString(artist.Username, maxResultArtistBytes) ||
		!boundedResultString(artist.Uid, maxResultArtistBytes) {
		return nil, errors.New("internal agent returned invalid artist fields")
	}
	if len(message.Tags) > maxResultTags {
		return nil, errors.New("internal agent returned too many tags")
	}
	for _, tag := range message.Tags {
		if !boundedResultString(tag, maxResultTagBytes) {
			return nil, errors.New("internal agent returned an invalid tag")
		}
	}
	mediaCount := len(message.Pictures) + len(message.UgoiraMetas) + len(message.Videos)
	if mediaCount == 0 || mediaCount > maxResultMedia {
		return nil, errors.New("internal agent returned an invalid media count")
	}
	if len(message.UgoiraMetas) != 0 && requested.source != shared.SourceTypePixiv {
		return nil, errors.New("internal agent returned ugoira for a non-Pixiv source")
	}
	pictureIndices := make(map[uint32]struct{}, len(message.Pictures))
	for i, picture := range message.Pictures {
		if picture == nil {
			return nil, fmt.Errorf("internal agent pictures[%d] is null", i)
		}
		if err := checkResultIndex(picture.Index, uint32(len(message.Pictures)-1), pictureIndices); err != nil {
			return nil, fmt.Errorf("internal agent pictures[%d]: %w", i, err)
		}
		if picture.Width > maxResultDimension || picture.Height > maxResultDimension {
			return nil, fmt.Errorf("internal agent pictures[%d] has invalid dimensions", i)
		}
	}
	ugoiraIndices := make(map[uint32]struct{}, len(message.UgoiraMetas))
	for i, meta := range message.UgoiraMetas {
		if meta == nil || meta.Data == nil {
			return nil, fmt.Errorf("internal agent ugoira_metas[%d] has null data", i)
		}
		if err := checkResultIndex(meta.Index, uint32(len(message.UgoiraMetas)-1), ugoiraIndices); err != nil {
			return nil, fmt.Errorf("internal agent ugoira_metas[%d]: %w", i, err)
		}
		data := meta.Data
		if data.Width < 0 || data.Height < 0 || data.Width > maxResultDimension || data.Height > maxResultDimension {
			return nil, fmt.Errorf("internal agent ugoira_metas[%d] has invalid dimensions", i)
		}
		if !validResultUgoiraMIME(data.MimeType) || len(data.Frames) == 0 || len(data.Frames) > maxResultFrames {
			return nil, fmt.Errorf("internal agent ugoira_metas[%d] has invalid MIME or frame count", i)
		}
		for j, frame := range data.Frames {
			if frame == nil || frame.Delay <= 0 || !safeResultFramePath(frame.File) {
				return nil, fmt.Errorf("internal agent ugoira_metas[%d].frames[%d] is invalid", i, j)
			}
		}
	}
	videoIndices := make(map[uint32]struct{}, len(message.Videos))
	videoIndexLimit := uint32(len(message.Videos))
	if videoIndexLimit > 0 {
		videoIndexLimit--
	}
	if requested.source == shared.SourceTypeKemono {
		videoIndexLimit = maxResultIndex
	}
	for i, video := range message.Videos {
		if video == nil {
			return nil, fmt.Errorf("internal agent videos[%d] is null", i)
		}
		if err := checkResultIndex(video.Index, videoIndexLimit, videoIndices); err != nil {
			return nil, fmt.Errorf("internal agent videos[%d]: %w", i, err)
		}
		if video.Width > maxResultDimension || video.Height > maxResultDimension || video.Duration > maxResultDurationMS ||
			!validResultVideoMIME(video.MimeType) {
			return nil, fmt.Errorf("internal agent videos[%d] has invalid dimensions, duration or MIME", i)
		}
	}

	proxyHost := resultPixivProxyHost(cfg.Pixiv.ImgProxy)
	validatedHosts := make(map[string]struct{})
	checkURL := func(raw string, required bool) error {
		if err := ctx.Err(); err != nil {
			return err
		}
		if raw == "" && !required {
			return nil
		}
		u, err := parseResultURL(raw)
		if err != nil {
			return err
		}
		host := strings.ToLower(u.Hostname())
		if !allowedResultMediaHost(requested.source, host, proxyHost) {
			return errors.New("media URL host is not an allowed source CDN")
		}
		if _, ok := validatedHosts[host]; ok {
			return nil
		}
		if err := httpclient.ValidatePublicURL(ctx, raw); err != nil {
			return err
		}
		validatedHosts[host] = struct{}{}
		return nil
	}
	for i, picture := range message.Pictures {
		if err := checkURL(picture.Original, true); err != nil {
			return nil, fmt.Errorf("internal agent pictures[%d].original: %w", i, err)
		}
		if err := checkURL(picture.Thumbnail, false); err != nil {
			return nil, fmt.Errorf("internal agent pictures[%d].thumbnail: %w", i, err)
		}
	}
	for i, meta := range message.UgoiraMetas {
		data := meta.Data
		for _, field := range []struct {
			name     string
			value    string
			required bool
		}{
			{"original_zip", data.OriginalZip, true},
			{"thumb_zip", data.ThumbZip, false},
			{"poster_original", data.PosterOriginal, false},
			{"poster_thumb", data.PosterThumb, false},
		} {
			if err := checkURL(field.value, field.required); err != nil {
				return nil, fmt.Errorf("internal agent ugoira_metas[%d].%s: %w", i, field.name, err)
			}
		}
	}
	for i, video := range message.Videos {
		if err := checkURL(video.Url, true); err != nil {
			return nil, fmt.Errorf("internal agent videos[%d].url: %w", i, err)
		}
		if err := checkURL(video.Poster, false); err != nil {
			return nil, fmt.Errorf("internal agent videos[%d].poster: %w", i, err)
		}
	}
	artwork := artworkFromProto(message)
	artwork.SourceURL = canonicalURL
	return artwork, nil
}

func boundedResultString(value string, maximum int) bool {
	return len(value) <= maximum && utf8.ValidString(value) && !strings.ContainsRune(value, '\x00')
}

func parseResultURL(raw string) (*url.URL, error) {
	if raw == "" || !boundedResultString(raw, maxResultURLBytes) || strings.TrimSpace(raw) != raw || strings.Contains(raw, "#") || forbiddenResultURLCharacters(raw) {
		return nil, errors.New("URL is empty or invalid")
	}
	decoded, err := url.PathUnescape(raw)
	if err != nil || forbiddenResultURLCharacters(decoded) {
		return nil, errors.New("URL has invalid encoding")
	}
	u, err := url.Parse(raw)
	if err != nil || u == nil || u.Opaque != "" || (u.Scheme != "http" && u.Scheme != "https") || u.Hostname() == "" || u.User != nil || u.Fragment != "" {
		return nil, errors.New("URL must be an absolute HTTP(S) URL without userinfo or fragment")
	}
	if strings.HasSuffix(u.Host, ":") || u.Port() != "" && !((u.Scheme == "http" && u.Port() == "80") || (u.Scheme == "https" && u.Port() == "443")) {
		return nil, errors.New("URL requires the scheme's default port")
	}
	if net.ParseIP(u.Hostname()) != nil || !validResultHostname(u.Hostname()) {
		return nil, errors.New("URL must use a source domain name")
	}
	if forbiddenResultURLCharacters(u.RawPath) || forbiddenResultURLCharacters(u.Path) || strings.Contains(strings.ToLower(u.RawPath), "%2f") {
		return nil, errors.New("URL has an ambiguous path")
	}
	return u, nil
}

func identifyArtworkResult(raw string, namers map[shared.SourceType]source.ArtworkNamer) (artworkResultIdentity, string, error) {
	u, err := parseResultURL(raw)
	if err != nil {
		return artworkResultIdentity{}, "", err
	}
	for sourceType, namer := range namers {
		if namer == nil {
			continue
		}
		route, identity, ok := resultArtworkRoute(sourceType, u)
		if !ok {
			continue
		}
		canonical, ok := namer.MatchesSourceURL(route)
		if !ok {
			continue
		}
		canonicalURL, err := parseResultURL(canonical)
		if err != nil {
			continue
		}
		_, canonicalIdentity, ok := resultArtworkRoute(sourceType, canonicalURL)
		if ok && canonicalIdentity == identity {
			return artworkResultIdentity{source: sourceType, id: canonicalIdentity}, canonical, nil
		}
	}
	return artworkResultIdentity{}, "", errors.New("URL is not an artwork recognized by a local source")
}

// Namer regexes also extract links from text. Check the actual host and complete path first.
func resultArtworkRoute(sourceType shared.SourceType, u *url.URL) (route, identity string, ok bool) {
	host := strings.ToLower(u.Hostname())
	parts := strings.Split(strings.TrimSuffix(strings.TrimPrefix(u.Path, "/"), "/"), "/")
	switch sourceType {
	case shared.SourceTypePixiv:
		if host != "pixiv.net" && host != "www.pixiv.net" && host != "m.pixiv.net" {
			return "", "", false
		}
		if len(parts) > 1 && resultPixivLanguage(parts[0]) {
			parts = parts[1:]
		}
		if len(parts) == 2 && (parts[0] == "artworks" || parts[0] == "i") && resultDecimalID(parts[1]) {
			identity = resultNormalizedID(parts[1])
		} else if len(parts) == 1 && parts[0] == "member_illust.php" {
			ids, err := url.ParseQuery(u.RawQuery)
			if err != nil || len(ids["illust_id"]) != 1 || !resultDecimalID(ids.Get("illust_id")) {
				return "", "", false
			}
			identity = resultNormalizedID(ids.Get("illust_id"))
		} else {
			return "", "", false
		}
		return "https://www.pixiv.net/artworks/" + identity, identity, true
	case shared.SourceTypeTwitter:
		if !resultExactHost(host, "twitter.com", "www.twitter.com", "mobile.twitter.com", "m.twitter.com", "x.com", "www.x.com", "mobile.x.com", "m.x.com") ||
			len(parts) != 3 || !resultTwitterUsername(parts[0]) || parts[1] != "status" || !resultDecimalID(parts[2]) {
			return "", "", false
		}
		identity = resultNormalizedID(parts[2])
		return "https://x.com/" + parts[0] + "/status/" + identity, identity, true
	case shared.SourceTypeBilibili:
		if host == "t.bilibili.com" && len(parts) == 1 && resultDecimalID(parts[0]) {
			identity = resultNormalizedID(parts[0])
		} else if resultExactHost(host, "bilibili.com", "www.bilibili.com", "m.bilibili.com") && len(parts) == 2 && parts[0] == "opus" && resultDecimalID(parts[1]) {
			identity = resultNormalizedID(parts[1])
		} else {
			return "", "", false
		}
		return "https://t.bilibili.com/" + identity, identity, true
	case shared.SourceTypeDanbooru:
		if host != "danbooru.donmai.us" {
			return "", "", false
		}
		if len(parts) == 2 && parts[0] == "posts" && resultDecimalID(parts[1]) {
			identity = resultNormalizedID(parts[1])
		} else if len(parts) == 3 && parts[0] == "post" && parts[1] == "show" && resultDecimalID(parts[2]) {
			identity = resultNormalizedID(parts[2])
		} else {
			return "", "", false
		}
		return "https://danbooru.donmai.us/posts/" + identity, identity, true
	case shared.SourceTypeYandere:
		if host != "yande.re" || len(parts) != 3 || parts[0] != "post" || parts[1] != "show" || !resultDecimalID(parts[2]) {
			return "", "", false
		}
		identity = resultNormalizedID(parts[2])
		return "https://yande.re/post/show/" + identity, identity, true
	case shared.SourceTypeNhentai:
		if !resultExactHost(host, "nhentai.net", "www.nhentai.net") || len(parts) != 2 || parts[0] != "g" || !resultDecimalID(parts[1]) {
			return "", "", false
		}
		identity = resultNormalizedID(parts[1])
		return "https://nhentai.net/g/" + identity, identity, true
	case shared.SourceTypeKemono:
		if !resultExactHost(host, "kemono.cr", "kemono.su", "www.kemono.cr", "www.kemono.su") || len(parts) != 5 ||
			!resultServiceName(parts[0]) || parts[1] != "user" || !resultDecimalID(parts[2]) || parts[3] != "post" || !resultDecimalID(parts[4]) {
			return "", "", false
		}
		identity = parts[0] + "/user/" + resultNormalizedID(parts[2]) + "/post/" + resultNormalizedID(parts[4])
		return "https://kemono.cr/" + identity, identity, true
	}
	return "", "", false
}

func resultPixivLanguage(value string) bool {
	switch value {
	case "en", "ja", "ko", "zh", "zh-cn", "zh-tw":
		return true
	}
	return false
}

func resultDecimalID(value string) bool {
	if value == "" {
		return false
	}
	for _, c := range value {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}

func resultNormalizedID(value string) string {
	if trimmed := strings.TrimLeft(value, "0"); trimmed != "" {
		return trimmed
	}
	return "0"
}

func resultServiceName(value string) bool {
	if value == "" {
		return false
	}
	for _, c := range value {
		if (c < 'a' || c > 'z') && (c < 'A' || c > 'Z') && (c < '0' || c > '9') && c != '_' {
			return false
		}
	}
	return true
}

func resultTwitterUsername(value string) bool {
	return len(value) <= 64 && resultServiceName(value)
}

func resultExactHost(host string, allowed ...string) bool {
	for _, candidate := range allowed {
		if host == candidate {
			return true
		}
	}
	return false
}

func resultDomain(host, domain string) bool {
	return host == domain || strings.HasSuffix(host, "."+domain)
}

func allowedResultMediaHost(sourceType shared.SourceType, host, pixivProxyHost string) bool {
	switch sourceType {
	case shared.SourceTypePixiv:
		return resultDomain(host, "pximg.net") || (pixivProxyHost != "" && host == pixivProxyHost)
	case shared.SourceTypeTwitter:
		return resultDomain(host, "twimg.com")
	case shared.SourceTypeBilibili:
		return resultDomain(host, "hdslb.com") || resultDomain(host, "biliimg.com") || resultDomain(host, "bilivideo.com") || resultDomain(host, "bilivideo.cn")
	case shared.SourceTypeDanbooru:
		return resultDomain(host, "donmai.us")
	case shared.SourceTypeYandere:
		return resultDomain(host, "yande.re")
	case shared.SourceTypeNhentai:
		return resultDomain(host, "nhentai.net") || resultDomain(host, "nhentai.to")
	case shared.SourceTypeKemono:
		return resultDomain(host, "kemono.cr") || resultDomain(host, "kemono.su")
	}
	return false
}

func resultPixivProxyHost(configured string) string {
	if configured == "" || strings.TrimSpace(configured) != configured {
		return ""
	}
	if !strings.Contains(configured, "://") {
		configured = "https://" + configured
	}
	u, err := parseResultURL(configured)
	if err != nil {
		return ""
	}
	return strings.ToLower(u.Hostname())
}

func checkResultIndex(index, limit uint32, seen map[uint32]struct{}) error {
	if index > limit {
		return errors.New("media index exceeds the supported range")
	}
	if _, exists := seen[index]; exists {
		return errors.New("duplicate media index")
	}
	seen[index] = struct{}{}
	return nil
}

func safeResultFramePath(name string) bool {
	if name == "." || !boundedResultString(name, maxResultFrameBytes) || !fs.ValidPath(name) || strings.ContainsAny(name, "\\:") {
		return false
	}
	for _, c := range name {
		if unicode.IsControl(c) {
			return false
		}
	}
	return true
}

func validResultUgoiraMIME(value string) bool {
	if !boundedResultString(value, maxResultMIMEBytes) {
		return false
	}
	mediaType, _, err := mime.ParseMediaType(value)
	if err != nil {
		return false
	}
	switch mediaType {
	case "image/jpeg", "image/png", "image/webp", "image/gif", "application/zip", "application/x-zip-compressed":
		return true
	}
	return false
}

func validResultVideoMIME(value string) bool {
	// Kemono does not provide MIME metadata; FxTwitter can return a short format name.
	if value == "" {
		return true
	}
	if !boundedResultString(value, maxResultMIMEBytes) {
		return false
	}
	switch strings.ToLower(value) {
	case "mp4", "webm", "mov", "mkv", "avi", "mpeg", "m4v", "gif":
		return true
	}
	mediaType, _, err := mime.ParseMediaType(value)
	if err != nil {
		return false
	}
	switch mediaType {
	case "video/mp4", "video/webm", "video/quicktime", "video/x-matroska", "video/x-msvideo", "video/mpeg", "video/ogg", "video/x-m4v", "image/gif", "application/vnd.apple.mpegurl", "application/x-mpegurl":
		return true
	}
	return false
}

func safeResultArtistUID(uid string) bool {
	if strings.TrimSpace(uid) == "" || uid == "." || uid == ".." || !boundedResultString(uid, maxResultArtistBytes) || strings.ContainsAny(uid, "/\\:") {
		return false
	}
	for _, c := range uid {
		if unicode.IsControl(c) {
			return false
		}
	}
	return true
}

func forbiddenResultURLCharacters(value string) bool {
	for _, c := range value {
		if c == '\\' || unicode.IsControl(c) {
			return true
		}
	}
	return false
}

func validResultHostname(host string) bool {
	if host == "" || len(host) > 253 {
		return false
	}
	for label := range strings.SplitSeq(host, ".") {
		if label == "" || len(label) > 63 || label[0] == '-' || label[len(label)-1] == '-' {
			return false
		}
		for _, c := range label {
			if !(c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= '0' && c <= '9' || c == '-') {
				return false
			}
		}
	}
	return true
}
