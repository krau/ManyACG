package common

import (
	"regexp"
	"sync"

	"github.com/krau/ManyACG/common"
	"github.com/krau/ManyACG/errs"
	"github.com/krau/ManyACG/types"
)

var (
	Sources          = make(map[types.SourceType]types.Source)
	allSources       = make(map[types.SourceType]types.Source) // all sources, including disabled ones
	SourceURLRegexps = make(map[types.SourceType]*regexp.Regexp)
	registeyMu       sync.RWMutex
)

func RegisterSource(sourceType types.SourceType, source types.Source) {
	registeyMu.Lock()
	defer registeyMu.Unlock()
	Sources[sourceType] = source
	SourceURLRegexps[sourceType] = source.GetSourceURLRegexp()
	allSources[sourceType] = source
}

func GetFileName(artwork *types.Artwork, picture *types.Picture) (string, error) {
	source := allSources[artwork.SourceType]
	if source == nil {
		return "", errs.ErrSourceNotSupported
	}
	fileName := source.GetFileName(artwork, picture)
	return common.SanitizeFileName(fileName), nil
}
