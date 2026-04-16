package handlers

import (
	"github.com/gofiber/fiber/v3"
	"github.com/krau/ManyACG/internal/infra/kvstor"
	"github.com/krau/ManyACG/internal/interface/rest/common"
	"github.com/krau/ManyACG/internal/service"
	"github.com/krau/ManyACG/internal/shared"
	"github.com/krau/ManyACG/pkg/log"
	"github.com/unvgo/ouid"
)

type RequestFetchArtwork struct {
	URL string `query:"url" form:"url" json:"url" validate:"required" message:"url is required and must be a valid url"`
}

type ResponseFetchArtwork struct {
	CacheID     string                       `json:"cache_id"`
	Title       string                       `json:"title"`
	Description string                       `json:"description"`
	SourceURL   string                       `json:"source_url"`
	R18         bool                         `json:"r18"`
	Tags        []string                     `json:"tags"`
	Artist      *ResponseFetchedArtist       `json:"artist"`
	SourceType  shared.SourceType            `json:"source_type"`
	Pictures    []*ResponseFetchedPicture    `json:"pictures"`
	Videos      []*ResponseFetchedVideo      `json:"videos,omitempty"`
	UgoiraMetas []*ResponseFetchedUgoiraMeta `json:"ugoira_metas,omitempty"`
}

type ResponseFetchedArtist struct {
	Name     string `json:"name"`
	Username string `json:"username"`
	UID      string `json:"uid"`
}

type ResponseFetchedPicture struct {
	Width     uint   `json:"width"`
	Height    uint   `json:"height"`
	Index     uint   `json:"index"`
	Thumbnail string `json:"thumbnail"`
	Original  string `json:"original"`
	FileName  string `json:"file_name"`
}

type ResponseFetchedVideo struct {
	URL      string `json:"url"`
	Poster   string `json:"poster"`
	MimeType string `json:"mime_type"`
	Index    uint   `json:"index"`
	Width    uint   `json:"width"`
	Height   uint   `json:"height"`
	Duration uint   `json:"duration"`
}

type ResponseFetchedUgoiraMeta struct {
	Data  shared.UgoiraMetaData `json:"data"`
	Index uint                  `json:"index"`
}

func fetchArtworkResponse(cacheID string, art shared.ArtworkLike, serv *service.Service) *ResponseFetchArtwork {
	pics := make([]*ResponseFetchedPicture, 0, len(art.GetPictures()))
	for _, pic := range art.GetPictures() {
		width, height := pic.GetSize()
		pics = append(pics, &ResponseFetchedPicture{
			Width:     width,
			Height:    height,
			Index:     pic.GetIndex(),
			Thumbnail: pic.GetThumbnail(),
			Original:  pic.GetOriginal(),
			FileName:  serv.PrettyFileName(art, pic),
		})
	}
	ugoiraMetas := make([]*ResponseFetchedUgoiraMeta, 0, len(art.GetUgoiraMetas()))
	for _, meta := range art.GetUgoiraMetas() {
		ugoiraMetas = append(ugoiraMetas, &ResponseFetchedUgoiraMeta{
			Data:  meta.GetData(),
			Index: meta.GetIndex(),
		})
	}
	videos := make([]*ResponseFetchedVideo, 0, len(art.GetVideos()))
	for _, video := range art.GetVideos() {
		videos = append(videos, &ResponseFetchedVideo{
			URL:      video.GetURL(),
			Poster:   video.GetPoster(),
			MimeType: video.GetMimeType(),
			Index:    video.GetIndex(),
			Width:    video.GetWidth(),
			Height:   video.GetHeight(),
			Duration: video.GetDuration(),
		})
	}
	return &ResponseFetchArtwork{
		CacheID:     cacheID,
		Title:       art.GetTitle(),
		Description: art.GetDescription(),
		SourceURL:   art.GetSourceURL(),
		R18:         art.GetR18(),
		Tags:        art.GetTags(),
		Artist: &ResponseFetchedArtist{
			Name:     art.GetArtist().GetName(),
			Username: art.GetArtist().GetUserName(),
			UID:      art.GetArtist().GetUID(),
		},
		SourceType:  art.GetType(),
		Pictures:    pics,
		Videos:      videos,
		UgoiraMetas: ugoiraMetas,
	}
}

func HandleFetchArtwork(ctx fiber.Ctx) error {
	key := ctx.Get("X-API-KEY")
	if key == "" {
		return common.NewError(fiber.StatusUnauthorized, "api key is required")
	}
	serv := common.MustGetState[*service.Service](ctx, common.StateKeyService)
	keyEnt, err := serv.GetApiKeyByKey(ctx, key)
	if err != nil {
		return common.NewError(fiber.StatusUnauthorized, "invalid api key")
	}
	if !keyEnt.HasPermission(shared.PermissionFetchArtwork) {
		return common.NewError(fiber.StatusForbidden, "no permission to fetch artwork")
	}
	if !keyEnt.CanUse() {
		return common.NewError(fiber.StatusForbidden, "api key quota exceeded")
	}

	req := new(RequestFetchArtwork)
	if err := ctx.Bind().All(req); err != nil {
		return err
	}
	sourceURL := serv.FindSourceURL(req.URL)
	if sourceURL == "" {
		return common.NewError(fiber.StatusBadRequest, "no valid source url found")
	}
	artwork, err := serv.GetOrFetchCachedArtwork(ctx, sourceURL)
	if err != nil {
		return err
	}
	serv.IncreaseApiKeyUsed(ctx, key)
	cacheID := ouid.New().Hex()
	if err := kvstor.Set(ctx, cacheID, artwork.GetSourceURL()); err != nil {
		log.Warn("failed to set cacheid", "data", artwork.GetSourceURL(), "err", err)
	}
	resp := fetchArtworkResponse(cacheID, artwork, serv)
	return ctx.JSON(common.NewSuccess(resp))
}
