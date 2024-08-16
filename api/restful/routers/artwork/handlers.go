package artwork

import (
	"ManyACG/api/restful/utils"
	manyacgErrors "ManyACG/errors"
	. "ManyACG/logger"
	"ManyACG/model"
	"ManyACG/service"
	"ManyACG/types"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func RandomArtworks(ctx *gin.Context) {
	var request GetRandomArtworksRequest
	if err := ctx.ShouldBind(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"status":  http.StatusBadRequest,
			"message": utils.BindError(ctx, err),
		})
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
		ctx.JSON(http.StatusInternalServerError, &ArtworkResponse{
			Status: http.StatusInternalServerError,
			Message: func() string {
				if hasKey {
					return err.Error()
				}
				return "Failed to get random artworks"
			}(),
		})
		return
	}
	if len(artworks) == 0 {
		ctx.JSON(http.StatusNotFound, &ArtworkResponse{
			Status:  http.StatusNotFound,
			Message: "Artworks not found",
		})
		return
	}
	ctx.JSON(http.StatusOK, ResponseFromArtworks(artworks, hasKey))
}

func GetArtwork(ctx *gin.Context) {
	id := ctx.Param("id")
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, &ArtworkResponse{
			Status:  http.StatusBadRequest,
			Message: "Invalid ID",
		})
		return
	}

	artwork, err := service.GetArtworkByID(ctx, objectID)
	hasKey := ctx.GetBool("auth")
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, &ArtworkResponse{
			Status: http.StatusInternalServerError,
			Message: func() string {
				if hasKey {
					return err.Error()
				}
				return "Failed to get artwork"
			}(),
		})
	}
	if artwork == nil {
		ctx.JSON(http.StatusNotFound, &ArtworkResponse{
			Status:  http.StatusNotFound,
			Message: "Artwork not found",
		})
		return
	}
	ctx.JSON(http.StatusOK, ResponseFromArtwork(artwork, hasKey))
}

func GetLatestArtworks(ctx *gin.Context) {
	var request GetLatestArtworksRequest
	if err := ctx.ShouldBind(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"status":  http.StatusBadRequest,
			"message": utils.BindError(ctx, err),
		})
		return
	}

	hasKey := ctx.GetBool("auth")
	r18Type := types.R18Type(request.R18)
	if r18Type != types.R18TypeNone && !hasKey {
		if !checkR18Permission(ctx) {
			return
		}
	}

	artworks, err := service.GetLatestArtworks(ctx, r18Type, request.Page, request.PageSize)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, &ArtworkResponse{
			Status: http.StatusInternalServerError,
			Message: func() string {
				if hasKey {
					return err.Error()
				}
				return "Failed to get latest artworks"
			}(),
		})
	}
	if len(artworks) == 0 {
		ctx.JSON(http.StatusNotFound, &ArtworkResponse{
			Status:  http.StatusNotFound,
			Message: "Artworks not found",
		})
		return
	}
	ctx.JSON(http.StatusOK, ResponseFromArtworks(artworks, hasKey))
}

func GetArtworkCount(ctx *gin.Context) {
	var request R18Request
	if err := ctx.ShouldBind(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"status":  http.StatusBadRequest,
			"message": utils.BindError(ctx, err),
		})
		return
	}
	r18Type := types.R18Type(request.R18)
	count, err := service.GetArtworkCount(ctx, r18Type)
	if err != nil {
		Logger.Error(err)
		ctx.JSON(http.StatusInternalServerError, &ArtworkResponse{
			Status:  http.StatusInternalServerError,
			Message: "Failed to get artwork count",
		})
		return
	}
	ctx.JSON(http.StatusOK, &ArtworkResponse{
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
			ctx.JSON(http.StatusBadRequest, &ArtworkResponse{
				Status:  http.StatusBadRequest,
				Message: "You have liked this artwork today",
			})
			return
		}
		ctx.JSON(http.StatusInternalServerError, &ArtworkResponse{
			Status: http.StatusInternalServerError,
			Message: func() string {
				if ctx.GetBool("auth") {
					return err.Error()
				}
				return "Failed to like artwork"
			}(),
		})
		return
	}
	ctx.JSON(http.StatusOK, &ArtworkResponse{
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
			ctx.JSON(http.StatusOK, &ArtworkResponse{
				Status:  http.StatusOK,
				Message: "Success",
				Data:    false,
			})
			return
		}
		ctx.JSON(http.StatusInternalServerError, &ArtworkResponse{
			Status: http.StatusInternalServerError,
			Message: func() string {
				if ctx.GetBool("auth") {
					return err.Error()
				}
				return "Failed to get favorite status"
			}(),
		})
		return
	}
	ctx.JSON(http.StatusOK, &ArtworkResponse{
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
		ctx.JSON(http.StatusInternalServerError, &ArtworkResponse{
			Status: http.StatusInternalServerError,
			Message: func() string {
				if ctx.GetBool("auth") {
					return err.Error()
				}
				return "Failed to favorite artwork"
			}(),
		})
		return
	}
	ctx.JSON(http.StatusOK, &ArtworkResponse{
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
		ctx.JSON(http.StatusInternalServerError, &ArtworkResponse{
			Status: http.StatusInternalServerError,
			Message: func() string {
				if ctx.GetBool("auth") {
					return err.Error()
				}
				return "Failed to unfavorite artwork"
			}(),
		})
		return
	}
	ctx.JSON(http.StatusOK, &ArtworkResponse{
		Status:  http.StatusOK,
		Message: "Success",
	})
}
