package artwork

import (
	"errors"
	"net/http"

	"github.com/krau/ManyACG/adapter"
	"github.com/krau/ManyACG/common"
	manyacgErrors "github.com/krau/ManyACG/errors"
	"github.com/krau/ManyACG/service"
	"github.com/krau/ManyACG/storage"
	"github.com/krau/ManyACG/types"

	"github.com/gabriel-vasile/mimetype"
	"github.com/gin-gonic/gin"
	. "github.com/krau/ManyACG/logger"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func RandomArtworks(ctx *gin.Context) {
	var request GetRandomArtworksRequest
	if err := ctx.ShouldBind(&request); err != nil {
		common.GinBindError(ctx, err)
		return
	}

	hasKey := ctx.GetBool("auth")
	r18Type := types.R18Type(request.R18)
	if r18Type != types.R18TypeNone && !hasKey {
		if !checkR18Permission(ctx) {
			return
		}
	}

	artworks, err := service.GetRandomArtworks(ctx, r18Type, request.Limit)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			common.GinErrorResponse(ctx, err, http.StatusNotFound, "Artworks not found")
			return
		}
		Logger.Errorf("Failed to get random artworks: %v", err)
		common.GinErrorResponse(ctx, err, http.StatusInternalServerError, "Failed to get random artworks")
		return
	}
	if len(artworks) == 0 {
		common.GinErrorResponse(ctx, manyacgErrors.ErrNotFoundArtworks, http.StatusNotFound, "Artworks not found")
		return
	}
	ctx.JSON(http.StatusOK, ResponseFromArtworks(artworks, hasKey))
}

func RandomArtworkPreview(ctx *gin.Context) {
	var request GetRandomArtworksRequest
	if err := ctx.ShouldBind(&request); err != nil {
		common.GinBindError(ctx, err)
		return
	}
	hasKey := ctx.GetBool("auth")
	r18Type := types.R18Type(request.R18)
	if r18Type != types.R18TypeNone && !hasKey {
		if !checkR18Permission(ctx) {
			return
		}
	}
	artwork, err := service.GetRandomArtworks(ctx, r18Type, 1, adapter.OnlyLoadPicture())
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			common.GinErrorResponse(ctx, err, http.StatusNotFound, "Artwork not found")
			return
		}
		common.GinErrorResponse(ctx, err, http.StatusInternalServerError, "Failed to get random artwork")
		return
	}
	if len(artwork) == 0 {
		common.GinErrorResponse(ctx, manyacgErrors.ErrNotFoundArtworks, http.StatusNotFound, "Artwork not found")
		return
	}

	picture := artwork[0].Pictures[0]

	switch picture.StorageInfo.Regular.Type {
	case types.StorageTypeLocal:
		ctx.File(picture.StorageInfo.Regular.Path)
	case types.StorageTypeAlist:
		ctx.Redirect(http.StatusFound, common.ApplyPathRule(picture.StorageInfo.Regular.Path))
	default:
		data, err := storage.GetFile(ctx, picture.StorageInfo.Regular)
		if err != nil {
			common.GinErrorResponse(ctx, err, http.StatusInternalServerError, "Failed to get file")
			return
		}
		mimeType := mimetype.Detect(data)
		ctx.Data(http.StatusOK, mimeType.String(), data)
	}
}

func GetArtworkList(ctx *gin.Context) {
	var request GetArtworkListRequest
	if err := ctx.ShouldBind(&request); err != nil {
		common.GinBindError(ctx, err)
		return
	}

	hasKey := ctx.GetBool("auth")
	r18Type := types.R18Type(request.R18)
	if r18Type != types.R18TypeNone && !hasKey {
		if !checkR18Permission(ctx) {
			return
		}
	}
	adapterOption := &adapter.AdapterOption{}
	if request.Simple {
		adapterOption = adapterOption.WithLoadArtist().WithOnlyIndexPicture()
	} else {
		adapterOption = adapter.LoadAll()
	}

	if request.ArtistID != "" {
		artistID, err := primitive.ObjectIDFromHex(request.ArtistID)
		if err != nil {
			common.GinErrorResponse(ctx, err, http.StatusBadRequest, "Invalid artist ID")
			return
		}
		getArtworkListByArtist(ctx, artistID, r18Type, request.Page, request.PageSize, adapterOption)
		return
	}
	if request.Tag != "" {
		getArtworkListByTag(ctx, request.Tag, r18Type, request.Page, request.PageSize, adapterOption)
		return
	}
	if request.Keyword != "" {
		keywordSlice := common.ParseStringTo2DArray(request.Keyword, ",", ";")
		if len(keywordSlice) == 0 {
			common.GinErrorResponse(ctx, errors.New("invalid keyword"), http.StatusBadRequest, "Invalid keyword")
			return
		}
		getArtworkListByKeyword(ctx, keywordSlice, r18Type, request.Page, request.PageSize, adapterOption)
		return
	}

	artworks, err := service.GetLatestArtworks(ctx, r18Type, request.Page, request.PageSize, adapterOption)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			common.GinErrorResponse(ctx, err, http.StatusNotFound, "Artworks not found")
			return
		}
		common.GinErrorResponse(ctx, err, http.StatusInternalServerError, "Failed to get artwork list")
		return
	}
	if len(artworks) == 0 {
		common.GinErrorResponse(ctx, manyacgErrors.ErrNotFoundArtworks, http.StatusNotFound, "Artworks not found")
		return
	}
	ctx.JSON(http.StatusOK, ResponseFromArtworks(artworks, hasKey))
}

func getArtworkListByArtist(ctx *gin.Context, artistID primitive.ObjectID, r18Type types.R18Type, page, pageSize int64, adapterOption ...*adapter.AdapterOption) {
	artworks, err := service.GetArtworksByArtistID(ctx, artistID, r18Type, page, pageSize, adapterOption...)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			common.GinErrorResponse(ctx, err, http.StatusNotFound, "Artworks not found")
			return
		}
		common.GinErrorResponse(ctx, err, http.StatusInternalServerError, "Failed to get artworks by artist")
		return
	}
	if len(artworks) == 0 {
		common.GinErrorResponse(ctx, manyacgErrors.ErrNotFoundArtworks, http.StatusNotFound, "Artworks not found")
		return
	}
	ctx.JSON(http.StatusOK, ResponseFromArtworks(artworks, ctx.GetBool("auth")))
}

func getArtworkListByTag(ctx *gin.Context, tag string, r18Type types.R18Type, page, pageSize int64, adapterOption ...*adapter.AdapterOption) {
	artworks, err := service.GetArtworksByTags(ctx, [][]string{{tag}}, r18Type, page, pageSize, adapterOption...)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			common.GinErrorResponse(ctx, err, http.StatusNotFound, "Artworks not found")
			return
		}
		common.GinErrorResponse(ctx, err, http.StatusInternalServerError, "Failed to get artworks by tag")
		return
	}
	if len(artworks) == 0 {
		common.GinErrorResponse(ctx, manyacgErrors.ErrNotFoundArtworks, http.StatusNotFound, "Artworks not found")
		return
	}
	ctx.JSON(http.StatusOK, ResponseFromArtworks(artworks, ctx.GetBool("auth")))
}

func getArtworkListByKeyword(ctx *gin.Context, keywordSlice [][]string, r18Type types.R18Type, page, pageSize int64, adapterOption ...*adapter.AdapterOption) {
	artworks, err := service.QueryArtworksByTextsPage(ctx, keywordSlice, r18Type, page, pageSize, adapterOption...)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			common.GinErrorResponse(ctx, err, http.StatusNotFound, "Artworks not found")
			return
		}
		common.GinErrorResponse(ctx, err, http.StatusInternalServerError, "Failed to get artworks by keyword")
		return
	}
	if len(artworks) == 0 {
		common.GinErrorResponse(ctx, manyacgErrors.ErrNotFoundArtworks, http.StatusNotFound, "Artworks not found")
		return
	}
	ctx.JSON(http.StatusOK, ResponseFromArtworks(artworks, ctx.GetBool("auth")))
}
