package nhentai

import (
	"errors"
	"regexp"
	"strings"
)

var (
	nhentaiSourceURLRegexp  = regexp.MustCompile(`nhentai\.net/g/\d+`)
	sourceURLPrefix         = "https://nhentai.net/g/"
	apiURLPrefix            = "https://nhentai.net/api/gallery/"
	ErrorInvalidNhentaiURL  = errors.New("invalid nhentai url")
	ErrNhentaiResponseError = errors.New("nhentai response error")
	numberRegexp            = regexp.MustCompile(`\d+`)
	originalUrlFormat       = "https://i%d.nhentai.net/galleries/%s/%d.%s"
)

func GetGalleryID(url string) string {
	matchUrl := nhentaiSourceURLRegexp.FindString(url)
	if matchUrl == "" {
		return ""
	}
	return numberRegexp.FindString(strings.Split(matchUrl, "/")[len(strings.Split(matchUrl, "/"))-1])
}

type nhentaiApiResp struct {
	Title struct {
		English  string `json:"english"`
		Japanese string `json:"japanese"`
		Pretty   string `json:"pretty"`
	} `json:"title"`
	MediaID string `json:"media_id"`
	Images  struct {
		Pages []struct {
			T string `json:"t"`
			W int    `json:"w"`
			H int    `json:"h"`
		} `json:"pages"`
	} `json:"images"`
	Tags []struct {
		Type string `json:"type"`
		Name string `json:"name"`
	} `json:"tags"`
}

var nhentaiExtsMap = map[string]string{
	"j": "jpg",
	"p": "png",
	"g": "gif",
	"w": "webp",
}

// func (n *Nhentai) crawlGallery(galleryID string) (*types.Artwork, error) {
// 	url := sourceURLPrefix + galleryID
// 	docResp, err := reqClient.R().Get(url)
// 	if err != nil {
// 		return nil, err
// 	}
// 	if docResp.IsErrorState() {
// 		return nil, fmt.Errorf("%w: %s", ErrNhentaiResponseError, docResp.Status)
// 	}

// 	doc, err := goquery.NewDocumentFromReader(docResp.Body)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to parse html: %w", err)
// 	}

// 	pictures := make([]*types.Picture, 0)

// 	pictureThumbUrls := make([]string, 0)
// 	doc.Find("#thumbnail-container > div > div > a > img").Each(func(i int, s *goquery.Selection) {
// 		thumbUrl, _ := s.Attr("data-src")
// 		if thumbUrl != "" {
// 			pictureThumbUrls = append(pictureThumbUrls, thumbUrl)
// 		}
// 	})
// 	if len(pictureThumbUrls) == 0 {
// 		return nil, fmt.Errorf("%w: %s", ErrNhentaiResponseError, "failed to parse gallery pictures")
// 	}

// 	apiResp, err := reqClient.R().Get(apiURLPrefix + galleryID)
// 	if err != nil {
// 		return nil, err
// 	}
// 	if apiResp.IsErrorState() {
// 		return nil, fmt.Errorf("%w: %s", ErrNhentaiResponseError, apiResp.Status)
// 	}

// 	var apiData nhentaiApiResp
// 	if err := json.Unmarshal(apiResp.Bytes(), &apiData); err != nil {
// 		return nil, fmt.Errorf("failed to parse api response: %w", err)
// 	}

// 	title := apiData.Title.Pretty
// 	description := apiData.Title.English + " " + apiData.Title.Japanese
// 	if title == "" {
// 		title = description
// 	}
// 	tags := make([]string, 0)
// 	var artistName, artistUid string
// 	for _, tag := range apiData.Tags {
// 		if artistName == "" && tag.Type == "artist" {
// 			artistName = tag.Name
// 			artistUid = tag.Name
// 			continue
// 		}
// 		if tag.Type == "tag" {
// 			tags = append(tags, tag.Name)
// 		}
// 	}
// 	if artistName == "" {
// 		artistName = "Nhentai"
// 		artistUid = "1"
// 	}
// 	originalUrls := make([]string, 0)
// 	for i, page := range apiData.Images.Pages {
// 		ext, ok := nhentaiExtsMap[page.T]
// 		if !ok {
// 			return nil, fmt.Errorf("%w: %s %s", ErrNhentaiResponseError, "unknown image type", page.T)
// 		}
// 		originalUrls = append(originalUrls, fmt.Sprintf(originalUrlFormat, rand.Intn(4)+1, apiData.MediaID, i+1, ext))
// 	}
// 	if len(originalUrls) != len(pictureThumbUrls) {
// 		return nil, fmt.Errorf("%w: %s", ErrNhentaiResponseError, "thumbnail count not equal to original count")
// 	}
// 	for i, thumbUrl := range pictureThumbUrls {
// 		pictures = append(pictures, &types.Picture{
// 			Index:     uint(i),
// 			Thumbnail: thumbUrl,
// 			Original:  originalUrls[i],
// 		})
// 	}

// 	return &types.Artwork{
// 		Title:       title,
// 		Description: description,
// 		Artist: &types.Artist{
// 			Name:     artistName,
// 			UID:      artistUid,
// 			Username: artistName,
// 			Type:     types.SourceTypeNhentai,
// 		},
// 		Tags:       tags,
// 		SourceURL:  url,
// 		Pictures:   pictures,
// 		R18:        true,
// 		SourceType: types.SourceTypeNhentai,
// 	}, nil
// }

// func getOriginalUrl(thumbUrl string) (string, error) {
// 	if thumbUrl == "" {
// 		return "", errors.New("empty thumb url")
// 	}
// 	if strings.HasSuffix(thumbUrl, "t.webp") {
// 		parts := strings.Split(thumbUrl, ".")
// 		if len(parts) != 4 {
// 			return "", fmt.Errorf("invalid thumb url: %s", thumbUrl)
// 		}
// 		parts[0] = fmt.Sprintf("https://i%d", rand.Intn(4)+1)
// 		parts[len(parts)-2] = parts[len(parts)-2][:len(parts[len(parts)-2])-1]
// 		return strings.Join(parts, "."), nil
// 	}
// 	parts := strings.Split(thumbUrl, "/")
// 	if len(parts) != 6 {
// 		return "", fmt.Errorf("invalid thumb url: %s", thumbUrl)
// 	}
// 	ext := parts[len(parts)-1]
// 	picIndex := numberRegexp.FindString(ext)
// 	if picIndex == "" {
// 		return "", fmt.Errorf("invalid thumb url: %s", thumbUrl)
// 	}
// 	galleryID := parts[len(parts)-2]
// 	return fmt.Sprintf("https://i%d.nhentai.net/galleries/%s/%s.webp", rand.Intn(4)+1, galleryID, picIndex), nil
// }
