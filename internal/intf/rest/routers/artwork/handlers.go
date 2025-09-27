package artwork

import (
	"errors"
	"net/http"

	"github.com/krau/ManyACG/api/restful/utils"
	"github.com/krau/ManyACG/common"
	"github.com/krau/ManyACG/errs"
	"github.com/krau/ManyACG/service"
	"github.com/krau/ManyACG/types"

	"github.com/gin-gonic/gin"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func GetArtwork(ctx *gin.Context) {
	id := ctx.Param("id")
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		utils.GinErrorResponse(ctx, err, http.StatusBadRequest, "Invalid artwork ID")
		return
	}

	artwork, err := service.GetArtworkByID(ctx, objectID)
	hasKey := ctx.GetBool("auth")
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			utils.GinErrorResponse(ctx, err, http.StatusNotFound, "Artwork not found")
			return
		}
		utils.GinErrorResponse(ctx, err, http.StatusInternalServerError, "Failed to get artwork")
		return
	}
	if artwork == nil {
		utils.GinErrorResponse(ctx, errs.ErrNotFoundArtworks, http.StatusNotFound, "Artwork not found")
		return
	}
	// if artwork.R18 && !hasKey {
	// 	if !checkR18Permission(ctx) {
	// 		return
	// 	}
	// }
	ctx.JSON(http.StatusOK, ResponseFromArtwork(artwork, hasKey))
}

func GetArtworkCount(ctx *gin.Context) {
	var request R18Request
	if err := ctx.ShouldBind(&request); err != nil {
		utils.GinBindError(ctx, err)
		return
	}
	r18Type := types.R18Type(request.R18)
	count, err := service.GetArtworkCount(ctx, r18Type)
	if err != nil {
		common.Logger.Errorf("Failed to get artwork count: %v", err)
		utils.GinErrorResponse(ctx, err, http.StatusInternalServerError, "Failed to get artwork count")
		return
	}
	ctx.JSON(http.StatusOK, &utils.RestfulCommonResponse[int64]{
		Status:  http.StatusOK,
		Message: "Success",
		Data:    count,
	})
}

func LikeArtwork(ctx *gin.Context) {
	artworkID := ctx.MustGet("artwork_id").(primitive.ObjectID)
	user := ctx.MustGet("user").(*types.UserModel)
	userID := user.ID
	err := service.CreateLike(ctx, userID, artworkID)
	if err != nil {
		if errors.Is(err, errs.ErrLikeExists) {
			utils.GinErrorResponse(ctx, err, http.StatusBadRequest, "You have liked this artwork today")
			return
		}
		utils.GinErrorResponse(ctx, err, http.StatusInternalServerError, "Failed to like artwork")
		return
	}
	ctx.JSON(http.StatusOK, &utils.RestfulCommonResponse[any]{
		Status:  http.StatusOK,
		Message: "Success",
	})
}

func GetArtworkLikeStatus(ctx *gin.Context) {
	artworkID := ctx.MustGet("artwork_id").(primitive.ObjectID)
	user := ctx.MustGet("user").(*types.UserModel)
	userID := user.ID
	like, err := service.GetLike(ctx, userID, artworkID)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			ctx.JSON(http.StatusOK, &utils.RestfulCommonResponse[bool]{
				Status:  http.StatusOK,
				Message: "Success",
				Data:    false,
			})
			return
		}
		utils.GinErrorResponse(ctx, err, http.StatusInternalServerError, "Failed to get like status")
		return
	}
	ctx.JSON(http.StatusOK, &utils.RestfulCommonResponse[bool]{
		Status:  http.StatusOK,
		Message: "Success",
		Data:    like != nil,
	})
}

func GetArtworkFavoriteStatus(ctx *gin.Context) {
	artworkID := ctx.MustGet("artwork_id").(primitive.ObjectID)
	user := ctx.MustGet("user").(*types.UserModel)
	userID := user.ID
	favorite, err := service.GetFavorite(ctx, userID, artworkID)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			ctx.JSON(http.StatusOK, &utils.RestfulCommonResponse[bool]{
				Status:  http.StatusOK,
				Message: "Success",
				Data:    false,
			})
			return
		}
		utils.GinErrorResponse(ctx, err, http.StatusInternalServerError, "Failed to get favorite status")
		return
	}
	ctx.JSON(http.StatusOK, &utils.RestfulCommonResponse[bool]{
		Status:  http.StatusOK,
		Message: "Success",
		Data:    favorite != nil,
	})
}

func FavoriteArtwork(ctx *gin.Context) {
	artworkID := ctx.MustGet("artwork_id").(primitive.ObjectID)
	user := ctx.MustGet("user").(*types.UserModel)
	userID := user.ID
	_, err := service.CreateFavorite(ctx, userID, artworkID)
	if err != nil {
		utils.GinErrorResponse(ctx, err, http.StatusInternalServerError, "Failed to favorite artwork")
		return
	}
	ctx.JSON(http.StatusOK, &utils.RestfulCommonResponse[any]{
		Status:  http.StatusOK,
		Message: "Success",
	})
}

func UnfavoriteArtwork(ctx *gin.Context) {
	artworkID := ctx.MustGet("artwork_id").(primitive.ObjectID)
	user := ctx.MustGet("user").(*types.UserModel)
	userID := user.ID
	err := service.DeleteFavorite(ctx, userID, artworkID)
	if err != nil {
		utils.GinErrorResponse(ctx, err, http.StatusInternalServerError, "Failed to unfavorite artwork")
		return
	}
	ctx.JSON(http.StatusOK, &utils.RestfulCommonResponse[any]{
		Status:  http.StatusOK,
		Message: "Success",
	})
}

func FetchArtwork(ctx *gin.Context) {
	var request FetchArtworkRequest
	if err := ctx.ShouldBind(&request); err != nil {
		utils.GinBindError(ctx, err)
		return
	}
	apiKey := ctx.MustGet("api_key").(*types.ApiKeyModel)
	sourceURL := sources.FindSourceURL(request.URL)
	if sourceURL == "" {
		utils.GinErrorResponse(ctx, errs.ErrSourceNotSupported, http.StatusBadRequest, "Source not supported")
		return
	}
	var artwork *types.Artwork
	var err error
	if !request.NoCache {
		artwork, err = service.GetArtworkByURLWithCacheFetch(ctx, sourceURL)
		if err != nil {
			utils.GinErrorResponse(ctx, err, http.StatusInternalServerError, "Failed to fetch artwork")
			return
		}
	} else {
		artwork, err = sources.GetArtworkInfo(sourceURL)
		if errors.Is(err, errs.ErrSourceNotSupported) {
			utils.GinErrorResponse(ctx, err, http.StatusBadRequest, "Source not supported")
			return
		}
		if err != nil {
			utils.GinErrorResponse(ctx, err, http.StatusInternalServerError, "Failed to fetch artwork")
			return
		}
		if _, err := service.GetCachedArtworkByURL(ctx, sourceURL); err != nil {
			service.CreateCachedArtwork(ctx, artwork, types.ArtworkStatusCached)
		}
	}
	if artwork == nil {
		utils.GinErrorResponse(ctx, errs.ErrNotFoundArtworks, http.StatusInternalServerError, "Artwork is nil")
		return
	}
	cacheId, err := service.CreateCallbackData(ctx, artwork.SourceURL)
	if err != nil {
		utils.GinErrorResponse(ctx, err, http.StatusInternalServerError, "Failed to create cache id")
		return
	}
	ctx.JSON(http.StatusOK, ResponseFromFetchedArtwork(artwork, cacheId))
	if err := service.IncreaseApiKeyUsed(ctx, apiKey.Key); err != nil {
		common.Logger.Errorf("Failed to increase api key used: %v", err)
	}
}
