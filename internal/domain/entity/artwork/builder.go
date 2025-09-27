package artwork

import (
	"errors"

	"github.com/krau/ManyACG/internal/common"
	"github.com/krau/ManyACG/pkg/objectuuid"
)

type Builder struct {
	artwork Artwork
}

func NewBuilder(id objectuuid.ObjectUUID) *Builder {
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

func (b *Builder) SourceType(sourceType common.SourceType) *Builder {
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

func (b *Builder) ArtistID(artistID objectuuid.ObjectUUID) *Builder {
	b.artwork.ArtistID = artistID
	return b
}

func (b *Builder) TagIDs(tagIDs *objectuuid.ObjectUUIDs) *Builder {
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
	for _, pic := range b.artwork.Pictures {
		if pic.ArtworkID != b.artwork.ID {
			return nil, errors.New("picture's artwork ID does not match")
		}
		if pic.Original == "" {
			return nil, errors.New("picture original URL is required")
		}
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

func (b *Builder) MustBuild() *Artwork {
	artwork, err := b.Build()
	if err != nil {
		panic(err)
	}
	return artwork
}