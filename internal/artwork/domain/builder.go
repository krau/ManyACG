package domain

import (
	"errors"

	"github.com/krau/ManyACG/internal/shared"
)

type Builder struct {
	artwork Artwork
}

func NewBuilder(id ArtworkID) *Builder {
	return &Builder{
		artwork: Artwork{
			ID: id,
		},
	}
}

func (b *Builder) Title(title string) *Builder {
	b.artwork.Title = title
	return b
}

func (b *Builder) Description(description string) *Builder {
	b.artwork.Description = description
	return b
}

func (b *Builder) R18(r18 bool) *Builder {
	b.artwork.R18 = r18
	return b
}

func (b *Builder) SourceType(sourceType shared.SourceType) *Builder {
	b.artwork.SourceType = sourceType
	return b
}

func (b *Builder) SourceURL(sourceURL string) *Builder {
	b.artwork.SourceURL = sourceURL
	return b
}

func (b *Builder) LikeCount(likeCount uint) *Builder {
	b.artwork.LikeCount = likeCount
	return b
}

func (b *Builder) ArtistID(artistID ArtistID) *Builder {
	b.artwork.ArtistID = artistID
	return b
}

func (b *Builder) TagIDs(tagIDs *TagIDs) *Builder {
	b.artwork.TagIDs = tagIDs
	return b
}

func (b *Builder) Pictures(pictures []Picture) *Builder {
	b.artwork.Pictures = pictures
	return b
}

func (b *Builder) Build() (*Artwork, error) {
	if b.artwork.ArtistID.IsZero() {
		return nil, errors.New("artist ID is required")
	}
	if len(b.artwork.Pictures) == 0 {
		return nil, errors.New("at least one picture is required")
	}
	if b.artwork.Title == "" {
		return nil, errors.New("title is required")
	}
	if b.artwork.SourceURL == "" {
		return nil, errors.New("source URL is required")
	}
	if b.artwork.SourceType == "" {
		return nil, errors.New("source type is required")
	}
	return &b.artwork, nil
}
