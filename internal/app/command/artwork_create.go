package command

import (
	"context"
	"errors"

	"github.com/krau/ManyACG/internal/common/decorator"
	"github.com/krau/ManyACG/internal/domain/entity/artist"
	"github.com/krau/ManyACG/internal/domain/entity/artwork"
	"github.com/krau/ManyACG/internal/domain/entity/tag"
	"github.com/krau/ManyACG/internal/domain/repo"
	"github.com/krau/ManyACG/internal/shared"
	"github.com/krau/ManyACG/pkg/objectuuid"
)

type ArtworkCreation struct {
	shared.ArtworkInfo
}

type ArtworkCreationResult struct {
	ArtworkID  objectuuid.ObjectUUID
	ArtistID   objectuuid.ObjectUUID
	PictureIDs *objectuuid.ObjectUUIDs
	TagIDs     *objectuuid.ObjectUUIDs
}

type CreateArtworkHandler decorator.CommandHandler[ArtworkCreation]

type createArtworkHandler struct {
	txRepo repo.TransactionRepo
}

func (h *createArtworkHandler) findOrUpsertArtist(ctx context.Context, repos *repo.TransactionRepos, info shared.ArtistInfo) (*artist.Artist, error) {
	artistEnt, err := repos.ArtistRepo.FindBySourceAndUID(ctx, string(info.Type), info.UID)
	if errors.Is(err, repo.ErrNotFound) {
		artistEnt = artist.NewArtist(objectuuid.New(), info.Name, info.Type, info.UID, info.Username)
		if err := repos.ArtistRepo.Save(ctx, artistEnt); err != nil {
			return nil, err
		}
	}
	if err != nil {
		return nil, err
	}
	artistEnt.Update(info.Name, info.Username)
	if err := repos.ArtistRepo.Save(ctx, artistEnt); err != nil {
		return nil, err
	}
	return artistEnt, nil
}

func (h *createArtworkHandler) findOrCreateTag(ctx context.Context, repos *repo.TransactionRepos, name string) (*tag.Tag, error) {
	tagEnt, err := repos.TagRepo.FindByNameWithAlias(ctx, name)
	if errors.Is(err, repo.ErrNotFound) {
		tagEnt = tag.NewTag(objectuuid.New(), name, nil)
		if err := repos.TagRepo.Save(ctx, tagEnt); err != nil {
			return nil, err
		}
	}
	if err != nil {
		return nil, err
	}
	return tagEnt, nil
}

func (h *createArtworkHandler) Handle(ctx context.Context, cmd ArtworkCreation) error {
	err := h.txRepo.WithTransaction(ctx, func(repos *repo.TransactionRepos) error {
		artistEnt, err := h.findOrUpsertArtist(ctx, repos, cmd.Artist)
		if err != nil {
			return err
		}
		tagIDs := objectuuid.NewObjectUUIDs()
		for _, tagName := range cmd.TagNames {
			tagEnt, err := h.findOrCreateTag(ctx, repos, tagName)
			if err != nil {
				return err
			}
			tagIDs.UnsafeAdd(tagEnt.ID)
		}
		_, err = repos.ArtworkRepo.FindByURL(ctx, cmd.SourceURL)
		if err == nil {
			return errors.New("artwork with the same source URL already exists")
		}
		if !errors.Is(err, repo.ErrNotFound) {
			return err
		}
		artworkID := objectuuid.New()
		pics := make([]artwork.Picture, 0, len(cmd.Pictures))
		for i, picInfo := range cmd.Pictures {
			picID := objectuuid.New()
			pics = append(pics, artwork.Picture{
				ID:           picID,
				ArtworkID:    artworkID,
				Index:        uint(i),
				Original:     picInfo.Original,
				Thumbnail:    picInfo.Thumbnail,
				ThumbHash:    picInfo.ThumbHash,
				Width:        picInfo.Width,
				Height:       picInfo.Height,
				Phash:        picInfo.Phash,
				TelegramInfo: picInfo.TelegramInfo,
				StorageInfo:  picInfo.StorageInfo,
			})
		}
		artworkEnt, err := artwork.NewBuilder(artworkID).
			Title(cmd.Title).
			Description(cmd.Description).
			R18(cmd.R18).
			SourceType(cmd.SourceType).
			SourceURL(cmd.SourceURL).
			LikeCount(0).
			ArtistID(artistEnt.ID).
			TagIDs(tagIDs).
			Pictures(pics).
			Build()
		if err != nil {
			return err
		}
		// result = &ArtworkCreationResult{
		// 	ArtworkID:  artworkEnt.ID,
		// 	ArtistID:   artistEnt.ID,
		// 	PictureIDs: objectuuid.NewObjectUUIDs(picIDs...),
		// 	TagIDs:     tagIDs,
		// }
		return repos.ArtworkRepo.Save(ctx, artworkEnt)
	})
	return err
}

func NewCreateArtworkHandler(txRepo repo.TransactionRepo) CreateArtworkHandler {
	return &createArtworkHandler{txRepo: txRepo}
}
