package nhentai

import (
	"errors"
	"fmt"
	"math/rand"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/krau/ManyACG/types"
)

var (
	nhentaiSourceURLRegexp  = regexp.MustCompile(`nhentai\.net/g/\d+`)
	sourceURLPrefix         = "https://nhentai.net/g/"
	ErrorInvalidNhentaiURL  = errors.New("invalid nhentai url")
	ErrNhentaiResponseError = errors.New("nhentai response error")
	numberRegexp            = regexp.MustCompile(`\d+`)
)

func GetGalleryID(url string) string {
	matchUrl := nhentaiSourceURLRegexp.FindString(url)
	if matchUrl == "" {
		return ""
	}
	return strings.Split(matchUrl, "/")[len(strings.Split(matchUrl, "/"))-1]
}

func (n *Nhentai) crawlGallery(galleryID string) (*types.Artwork, error) {
	url := sourceURLPrefix + galleryID
	resp, err := reqClient.R().Get(url)
	if err != nil {
		return nil, err
	}
	if resp.IsErrorState() {
		return nil, fmt.Errorf("%w: %s", ErrNhentaiResponseError, resp.Status)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse html: %w", err)
	}

	title := doc.Find("#info > h1.title > span.pretty").Text()

	description := doc.Find("#info > h2 > span.pretty").Text()

	artist := doc.Find("#tags > div:nth-child(4) > span > a > span.name").First().Text()
	artistUid := artist
	if artist == "" {
		artist = "Nhentai"
		artistUid = "1"
	}

	tags := make([]string, 0)
	doc.Find("section#tags .tag-container .tags .tag .name").Each(func(i int, s *goquery.Selection) {
		tags = append(tags, s.Text())
	})

	if title == "" && description == "" && artist == "" && len(tags) == 0 {
		return nil, fmt.Errorf("%w: %s", ErrNhentaiResponseError, "failed to parse gallery info")
	}

	pictures := make([]*types.Picture, 0)

	pictureThumbUrls := make([]string, 0)
	doc.Find("#thumbnail-container > div > div > a > img").Each(func(i int, s *goquery.Selection) {
		thumbUrl, _ := s.Attr("data-src")
		if thumbUrl != "" {
			pictureThumbUrls = append(pictureThumbUrls, thumbUrl)
		}
	})
	for i, thumbUrl := range pictureThumbUrls {
		orgUrl, err := getOriginalUrl(thumbUrl)
		if err != nil {
			return nil, fmt.Errorf("%w: %s", ErrNhentaiResponseError, err)
		}
		pictures = append(pictures, &types.Picture{
			Index:     uint(i),
			Thumbnail: thumbUrl,
			Original:  orgUrl,
		})
	}

	if len(pictures) == 0 {
		return nil, fmt.Errorf("%w: %s", ErrNhentaiResponseError, "failed to parse gallery pictures")
	}

	return &types.Artwork{
		Title:       title,
		Description: description,
		Artist: &types.Artist{
			Name:     artist,
			UID:      artistUid,
			Username: artist,
			Type:     types.SourceTypeNhentai,
		},
		Tags:       tags,
		SourceURL:  url,
		Pictures:   pictures,
		R18:        true,
		SourceType: types.SourceTypeNhentai,
	}, nil
}

func getOriginalUrl(thumbUrl string) (string, error) {
	if thumbUrl == "" {
		return "", errors.New("empty thumb url")
	}
	if strings.HasSuffix(thumbUrl, "t.webp") {
		parts := strings.Split(thumbUrl, ".")
		if len(parts) != 4 {
			return "", fmt.Errorf("invalid thumb url: %s", thumbUrl)
		}
		parts[0] = fmt.Sprintf("https://i%d", rand.Intn(4)+1)
		parts[len(parts)-2] = parts[len(parts)-2][:len(parts[len(parts)-2])-1]
		return strings.Join(parts, "."), nil
	}
	parts := strings.Split(thumbUrl, "/")
	if len(parts) != 6 {
		return "", fmt.Errorf("invalid thumb url: %s", thumbUrl)
	}
	ext := parts[len(parts)-1]
	picIndex := numberRegexp.FindString(ext)
	if picIndex == "" {
		return "", fmt.Errorf("invalid thumb url: %s", thumbUrl)
	}
	galleryID := parts[len(parts)-2]
	return fmt.Sprintf("https://i%d.nhentai.net/galleries/%s/%s.webp", rand.Intn(4)+1, galleryID, picIndex), nil
}
