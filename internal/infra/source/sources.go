package source

import (
	"context"
	"errors"
	"strings"

	"github.com/krau/ManyACG/internal/shared"
)

type FetchedArtwork struct {
	shared.ArtworkInfo
}

type ArtworkSource interface {
	Init(ctx context.Context, cfg Config) error
	GetArtworkInfo(ctx context.Context, sourceUrl string) (*FetchedArtwork, error)
	MatchesSourceURL(sourceUrl string) (string, bool)
	FetchNewArtworks(ctx context.Context, limit int) ([]*FetchedArtwork, error)
}

func GetArtworkInfo(ctx context.Context, sourceURL string) (*FetchedArtwork, error) {
	for _, sou := range sources {
		if _, ok := sou.MatchesSourceURL(sourceURL); ok {
			return sou.GetArtworkInfo(ctx, sourceURL)
		}
	}
	return nil, errors.New("no supported source found")
}

func FindSourceURL(text string) string {
	text = strings.ReplaceAll(text, "\n", " ")
	for _, sou := range sources {
		if url, ok := sou.MatchesSourceURL(text); ok {
			return url
		}
	}
	return ""
}

// MatchesSourceURL returns whether the text contains a source URL.
func MatchesSourceURL(text string) bool {
	text = strings.ReplaceAll(text, "\n", " ")
	for _, sou := range sources {
		if _, ok := sou.MatchesSourceURL(text); ok {
			return true
		}
	}
	return false
}

// func isSourceEnabled(sourceType types.SourceType) bool {
// 	cfgValue := reflect.ValueOf(config.Get().Source)
// 	field := cfgValue.FieldByName(strings.Title(string(sourceType)))
// 	if !field.IsValid() {
// 		return false
// 	}
// 	if field.Kind() != reflect.Struct {
// 		return false
// 	}
// 	enableField := field.FieldByName("Enable")
// 	if !enableField.IsValid() {
// 		return false
// 	}
// 	if enableField.Kind() != reflect.Bool {
// 		return false
// 	}
// 	return enableField.Bool()
// }
