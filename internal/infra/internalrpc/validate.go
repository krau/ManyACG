package internalrpc

import (
	"errors"
	"fmt"
	"strings"
	"unicode"

	"github.com/krau/ManyACG/internal/infra/internalrpc/internalpb"
	"github.com/krau/ManyACG/internal/infra/source"
	"github.com/krau/ManyACG/internal/model/dto"
	"github.com/krau/ManyACG/internal/shared"
)

func validateArtworkResult(
	requestedURL string,
	message *internalpb.FetchedArtwork,
	namers map[shared.SourceType]source.ArtworkNamer,
) (*dto.FetchedArtwork, error) {
	if message == nil {
		return nil, errors.New("internal agent returned a null artwork")
	}
	namer := namers[shared.SourceType(message.SourceType)]
	if namer == nil {
		return nil, errors.New("internal agent returned an unknown source")
	}
	requested, requestedOK := namer.MatchesSourceURL(requestedURL)
	returned, returnedOK := namer.MatchesSourceURL(message.SourceUrl)
	if !requestedOK || !returnedOK || requested == "" || requested != returned {
		return nil, errors.New("internal agent returned a different artwork or source")
	}
	artist := message.Artist
	if artist == nil || artist.Type != message.SourceType || !safeResultArtistUID(artist.Uid) {
		return nil, errors.New("internal agent returned an invalid artist")
	}
	if len(message.Pictures)+len(message.UgoiraMetas)+len(message.Videos) == 0 {
		return nil, errors.New("internal agent returned no media")
	}
	for i, picture := range message.Pictures {
		if picture == nil || picture.Original == "" {
			return nil, fmt.Errorf("internal agent pictures[%d] has no original URL", i)
		}
	}
	for i, meta := range message.UgoiraMetas {
		if meta == nil || meta.Data == nil || meta.Data.OriginalZip == "" || len(meta.Data.Frames) == 0 {
			return nil, fmt.Errorf("internal agent ugoira_metas[%d] has incomplete data", i)
		}
		for j, frame := range meta.Data.Frames {
			if frame == nil || frame.File == "" || frame.Delay <= 0 {
				return nil, fmt.Errorf("internal agent ugoira_metas[%d].frames[%d] is invalid", i, j)
			}
		}
	}
	for i, video := range message.Videos {
		if video == nil || video.Url == "" {
			return nil, fmt.Errorf("internal agent videos[%d] has no URL", i)
		}
	}
	artwork := artworkFromProto(message)
	artwork.SourceURL = requested
	return artwork, nil
}

// Artist UIDs become directory components when media is stored.
func safeResultArtistUID(uid string) bool {
	return strings.TrimSpace(uid) != "" && uid != "." && uid != ".." &&
		!strings.ContainsAny(uid, "/\\:") && !strings.ContainsFunc(uid, unicode.IsControl)
}
