package artwork

import (
	"ManyACG/adapter"
	"ManyACG/common"
	manyacgErrors "ManyACG/errors"
	"ManyACG/model"
	"ManyACG/service"
	"ManyACG/storage"
	"ManyACG/types"
	"errors"
	"net/http"

	"github.com/gabriel-vasile/mimetype"
	"github.com/gin-gonic/gin"
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
		common.GinErrorResponse(ctx, err, http.StatusInternalServerError, "Failed to get random artworks")
		return
	}
	if len(artworks) == 0 {
		common.GinErrorResponse(ctx, err, http.StatusNotFound, "Artworks not found")
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
		common.GinErrorResponse(ctx, err, http.StatusInternalServerError, "Failed to get random artwork")
		return
	}
	if len(artwork) == 0 {
		common.GinErrorResponse(ctx, err, http.StatusNotFound, "Artwork not found")
		return
	}

	data, err := storage.GetFile(ctx, artwork[0].Pictures[0].StorageInfo.Regular)
	if err != nil {
		common.GinErrorResponse(ctx, err, http.StatusInternalServerError, "Failed to get file")
		return
	}
	mimeType := mimetype.Detect(data)
	ctx.Data(http.StatusOK, mimeType.String(), data)
}

func GetArtwork(ctx *gin.Context) {
	id := ctx.Param("id")
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		common.GinErrorResponse(ctx, err, http.StatusBadRequest, "Invalid artwork ID")
		return
	}

	artwork, err := service.GetArtworkByID(ctx, objectID)
	hasKey := ctx.GetBool("auth")
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			common.GinErrorResponse(ctx, err, http.StatusNotFound, "Artwork not found")
			return
		}
		common.GinErrorResponse(ctx, err, http.StatusInternalServerError, "Failed to get artwork")
		return
	}
	if artwork == nil {
		common.GinErrorResponse(ctx, err, http.StatusNotFound, "Artwork not found")
		return
	}
	if artwork.R18 && !hasKey {
		if !checkR18Permission(ctx) {
			return
		}
	}
	ctx.JSON(http.StatusOK, ResponseFromArtwork(artwork, hasKey))
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

	if request.ArtistID != "" {
		artistID, err := primitive.ObjectIDFromHex(request.ArtistID)
		if err != nil {
			common.GinErrorResponse(ctx, err, http.StatusBadRequest, "Invalid artist ID")
		}
		artworks, err := service.GetArtworksByArtistID(ctx, artistID, r18Type, request.Page, request.PageSize, adapter.OnlyLoadPicture())
		if err != nil {
			common.GinErrorResponse(ctx, err, http.StatusInternalServerError, "Failed to get artworks by artist")
		}
		if len(artworks) == 0 {
			common.GinErrorResponse(ctx, err, http.StatusNotFound, "Artworks not found")
			return
		}
		ctx.JSON(http.StatusOK, ResponseFromArtworks(artworks, hasKey))
		return
	}

	artworks, err := service.GetLatestArtworks(ctx, r18Type, request.Page, request.PageSize, adapter.OnlyLoadPicture())

	if err != nil {
		common.GinErrorResponse(ctx, err, http.StatusInternalServerError, "Failed to get artwork list")
	}

	if len(artworks) == 0 {
		common.GinErrorResponse(ctx, err, http.StatusNotFound, "Artworks not found")
		return
	}
	ctx.JSON(http.StatusOK, ResponseFromArtworks(artworks, hasKey))
}

func GetArtworkCount(ctx *gin.Context) {
	var request R18Request
	if err := ctx.ShouldBind(&request); err != nil {
		common.GinBindError(ctx, err)
		return
	}
	r18Type := types.R18Type(request.R18)
	count, err := service.GetArtworkCount(ctx, r18Type)
	if err != nil {
		common.GinErrorResponse(ctx, err, http.StatusInternalServerError, "Failed to get artwork count")
		return
	}
	ctx.JSON(http.StatusOK, &common.RestfulCommonResponse[int64]{
		Status:  http.StatusOK,
		Message: "Success",
		Data:    count,
	})
}

func LikeArtwork(ctx *gin.Context) {
	artworkID := ctx.MustGet("artwork_id").(primitive.ObjectID)
	user := ctx.MustGet("user").(*model.UserModel)
	userID := user.ID
	err := service.CreateLike(ctx, userID, artworkID)
	if err != nil {
		if errors.Is(err, manyacgErrors.ErrLikeExists) {
			common.GinErrorResponse(ctx, err, http.StatusBadRequest, "You have liked this artwork today")
			return
		}
		common.GinErrorResponse(ctx, err, http.StatusInternalServerError, "Failed to like artwork")
		return
	}
	ctx.JSON(http.StatusOK, &common.RestfulCommonResponse[any]{
		Status:  http.StatusOK,
		Message: "Success",
	})
}

func GetArtworkFavoriteStatus(ctx *gin.Context) {
	artworkID := ctx.MustGet("artwork_id").(primitive.ObjectID)
	user := ctx.MustGet("user").(*model.UserModel)
	userID := user.ID
	favorite, err := service.GetFavorite(ctx, userID, artworkID)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			ctx.JSON(http.StatusOK, &common.RestfulCommonResponse[bool]{
				Status:  http.StatusOK,
				Message: "Success",
				Data:    false,
			})
			return
		}
		common.GinErrorResponse(ctx, err, http.StatusInternalServerError, "Failed to get favorite status")
		return
	}
	ctx.JSON(http.StatusOK, &common.RestfulCommonResponse[bool]{
		Status:  http.StatusOK,
		Message: "Success",
		Data:    favorite != nil,
	})
}

func FavoriteArtwork(ctx *gin.Context) {
	artworkID := ctx.MustGet("artwork_id").(primitive.ObjectID)
	user := ctx.MustGet("user").(*model.UserModel)
	userID := user.ID
	_, err := service.CreateFavorite(ctx, userID, artworkID)
	if err != nil {
		common.GinErrorResponse(ctx, err, http.StatusInternalServerError, "Failed to favorite artwork")
		return
	}
	ctx.JSON(http.StatusOK, &common.RestfulCommonResponse[any]{
		Status:  http.StatusOK,
		Message: "Success",
	})
}

func UnfavoriteArtwork(ctx *gin.Context) {
	artworkID := ctx.MustGet("artwork_id").(primitive.ObjectID)
	user := ctx.MustGet("user").(*model.UserModel)
	userID := user.ID
	err := service.DeleteFavorite(ctx, userID, artworkID)
	if err != nil {
		common.GinErrorResponse(ctx, err, http.StatusInternalServerError, "Failed to unfavorite artwork")
		return
	}
	ctx.JSON(http.StatusOK, &common.RestfulCommonResponse[any]{
		Status:  http.StatusOK,
		Message: "Success",
	})
}
