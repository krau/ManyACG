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

	// title: #info > h1 > span.pretty
	title := doc.Find("#info > h1 > span.pretty").Text()

	// description: #info > h2 > span.pretty
	description := doc.Find("#info > h2 > span.pretty").Text()

	// artist: #tags > div:nth-child(4) > span > a > span.name
	artist := doc.Find("#tags > div:nth-child(4) > span > a > span.name").Text()

	// tags: #tags > div:nth-child(3) > span 下面的 a 下面的 span.name
	tags := make([]string, 0)
	doc.Find("#tags > div:nth-child(3) > span > a > span.name").Each(func(i int, s *goquery.Selection) {
		tags = append(tags, s.Text())
	})

	if title == "" || description == "" || artist == "" || len(tags) == 0 {
		return nil, fmt.Errorf("%w: %s", ErrNhentaiResponseError, "failed to parse gallery info")
	}

	pictures := make([]*types.Picture, 0)

	// pictures: #thumbnail-container > div > div:nth-child(1) > a > img 标签的 data-src 属性

	doc.Find("#thumbnail-container > div > div > a > img").Each(func(i int, s *goquery.Selection) {
		thumbUrl, _ := s.Attr("data-src")
		pictures = append(pictures, &types.Picture{
			Index:     uint(i),
			Thumbnail: thumbUrl,
			Original:  getOriginalUrl(thumbUrl),
		})
	})

	if len(pictures) == 0 {
		return nil, fmt.Errorf("%w: %s", ErrNhentaiResponseError, "failed to parse gallery pictures")
	}

	return &types.Artwork{
		Title:       title,
		Description: description,
		Artist: &types.Artist{
			Name:     artist,
			UID:      artist,
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

func getOriginalUrl(thumbUrl string) string {
	if thumbUrl == "" {
		return ""
	}
	parts := strings.Split(thumbUrl, ".")
	parts[0] = fmt.Sprintf("https://i%d", rand.Intn(4)+1)
	parts[len(parts)-2] = parts[len(parts)-2][:len(parts[len(parts)-2])-1]
	return strings.Join(parts, ".")
}
