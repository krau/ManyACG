package artwork

import (
	"errors"
	"net/http"

	"github.com/krau/ManyACG/adapter"
	"github.com/krau/ManyACG/api/restful/utils"
	"github.com/krau/ManyACG/common"
	"github.com/krau/ManyACG/errs"
	"github.com/krau/ManyACG/service"
	"github.com/krau/ManyACG/storage"
	"github.com/krau/ManyACG/types"

	"github.com/gin-gonic/gin"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func RandomArtworks(ctx *gin.Context) {
	var request GetRandomArtworksRequest
	if err := ctx.ShouldBind(&request); err != nil {
		utils.GinBindError(ctx, err)
		return
	}

	hasKey := ctx.GetBool("auth")
	r18Type := types.R18Type(request.R18)

	artworks, err := service.GetRandomArtworks(ctx, r18Type, request.Limit)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			utils.GinErrorResponse(ctx, err, http.StatusNotFound, "Artworks not found")
			return
		}
		common.Logger.Errorf("Failed to get random artworks: %v", err)
		utils.GinErrorResponse(ctx, err, http.StatusInternalServerError, "Failed to get random artworks")
		return
	}
	if len(artworks) == 0 {
		utils.GinErrorResponse(ctx, errs.ErrNotFoundArtworks, http.StatusNotFound, "Artworks not found")
		return
	}
	ctx.JSON(http.StatusOK, ResponseFromArtworks(artworks, hasKey))
}

func RandomArtworkPreview(ctx *gin.Context) {
	var request GetRandomArtworksRequest
	if err := ctx.ShouldBind(&request); err != nil {
		utils.GinBindError(ctx, err)
		return
	}
	r18Type := types.R18Type(request.R18)
	artwork, err := service.GetRandomArtworks(ctx, r18Type, 1, adapter.OnlyLoadPicture())
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			utils.GinErrorResponse(ctx, err, http.StatusNotFound, "Artwork not found")
			return
		}
		utils.GinErrorResponse(ctx, err, http.StatusInternalServerError, "Failed to get random artwork")
		return
	}
	if len(artwork) == 0 {
		utils.GinErrorResponse(ctx, errs.ErrNotFoundArtworks, http.StatusNotFound, "Artwork not found")
		return
	}

	picture := artwork[0].Pictures[0]
	if picture.StorageInfo.Regular == nil {
		utils.GinErrorResponse(ctx, errors.New("picture not found"), http.StatusNotFound, "Picture not found")
		return
	}

	picUrl := utils.ApplyApiStoragePathRule(picture.StorageInfo.Regular)
	if picUrl == "" || picUrl == picture.StorageInfo.Regular.Path {
		storage.ServeFile(ctx, picture.StorageInfo.Regular)
		return
	}
	ctx.Redirect(http.StatusFound, picUrl)
}

func GetArtworkList(ctx *gin.Context) {
	var request GetArtworkListRequest
	if err := ctx.ShouldBind(&request); err != nil {
		utils.GinBindError(ctx, err)
		return
	}

	hasKey := ctx.GetBool("auth")
	r18Type := types.R18Type(request.R18)
	adapterOption := &types.AdapterOption{}
	if request.Simple {
		adapterOption = adapterOption.WithLoadArtist().WithOnlyIndexPicture()
	} else {
		adapterOption = adapter.LoadAll()
	}

	if request.ArtistID != "" {
		artistID, err := primitive.ObjectIDFromHex(request.ArtistID)
		if err != nil {
			utils.GinErrorResponse(ctx, err, http.StatusBadRequest, "Invalid artist ID")
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
		if request.Hybrid {
			getArtworkListHybrid(ctx, request.Keyword, 0.8, request.Page*request.PageSize, request.PageSize, r18Type, adapterOption)
			return
		}
		keywordSlice := common.ParseStringTo2DArray(request.Keyword, ",", "|")
		if len(keywordSlice) == 0 {
			utils.GinErrorResponse(ctx, errors.New("invalid keyword"), http.StatusBadRequest, "Invalid keyword")
			return
		}
		getArtworkListByKeyword(ctx, keywordSlice, r18Type, request.Page, request.PageSize, adapterOption)
		return
	}
	if request.SimilarTarget != "" {
		artworkID, err := primitive.ObjectIDFromHex(request.SimilarTarget)
		if err != nil {
			utils.GinErrorResponse(ctx, err, http.StatusBadRequest, "Invalid similar target ID")
			return
		}
		_, err = service.GetArtworkByID(ctx, artworkID, adapter.LoadNone())
		if err != nil {
			if errors.Is(err, mongo.ErrNoDocuments) {
				utils.GinErrorResponse(ctx, err, http.StatusNotFound, "Target Artwork not exist")
				return
			}
			utils.GinErrorResponse(ctx, err, http.StatusInternalServerError, "Failed to get target artwork")
			return
		}
		artworks, err := service.SearchSimilarArtworks(ctx, artworkID.Hex(), request.Page*request.PageSize, request.PageSize, r18Type, adapterOption)
		if err != nil {
			utils.GinErrorResponse(ctx, err, http.StatusInternalServerError, "Failed to get similar artworks")
			return
		}
		if len(artworks) == 0 {
			utils.GinErrorResponse(ctx, errs.ErrNotFoundArtworks, http.StatusNotFound, "Artworks not found")
			return
		}
		ctx.JSON(http.StatusOK, ResponseFromArtworks(artworks, hasKey))
		return
	}

	artworks, err := service.GetLatestArtworks(ctx, r18Type, request.Page, request.PageSize, adapterOption)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			utils.GinErrorResponse(ctx, err, http.StatusNotFound, "Artworks not found")
			return
		}
		utils.GinErrorResponse(ctx, err, http.StatusInternalServerError, "Failed to get artwork list")
		return
	}
	if len(artworks) == 0 {
		utils.GinErrorResponse(ctx, errs.ErrNotFoundArtworks, http.StatusNotFound, "Artworks not found")
		return
	}
	ctx.JSON(http.StatusOK, ResponseFromArtworks(artworks, hasKey))
}

func getArtworkListByArtist(ctx *gin.Context, artistID primitive.ObjectID, r18Type types.R18Type, page, pageSize int64, adapterOption ...*types.AdapterOption) {
	artworks, err := service.GetArtworksByArtistID(ctx, artistID, r18Type, page, pageSize, adapterOption...)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			utils.GinErrorResponse(ctx, err, http.StatusNotFound, "Artworks not found")
			return
		}
		utils.GinErrorResponse(ctx, err, http.StatusInternalServerError, "Failed to get artworks by artist")
		return
	}
	if len(artworks) == 0 {
		utils.GinErrorResponse(ctx, errs.ErrNotFoundArtworks, http.StatusNotFound, "Artworks not found")
		return
	}
	ctx.JSON(http.StatusOK, ResponseFromArtworks(artworks, ctx.GetBool("auth")))
}

func getArtworkListByTag(ctx *gin.Context, tag string, r18Type types.R18Type, page, pageSize int64, adapterOption ...*types.AdapterOption) {
	artworks, err := service.GetArtworksByTags(ctx, [][]string{{tag}}, r18Type, page, pageSize, adapterOption...)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			utils.GinErrorResponse(ctx, err, http.StatusNotFound, "Artworks not found")
			return
		}
		utils.GinErrorResponse(ctx, err, http.StatusInternalServerError, "Failed to get artworks by tag")
		return
	}
	if len(artworks) == 0 {
		utils.GinErrorResponse(ctx, errs.ErrNotFoundArtworks, http.StatusNotFound, "Artworks not found")
		return
	}
	ctx.JSON(http.StatusOK, ResponseFromArtworks(artworks, ctx.GetBool("auth")))
}

func getArtworkListByKeyword(ctx *gin.Context, keywordSlice [][]string, r18Type types.R18Type, page, pageSize int64, adapterOption ...*types.AdapterOption) {
	artworks, err := service.QueryArtworksByTextsPage(ctx, keywordSlice, r18Type, page, pageSize, adapterOption...)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			utils.GinErrorResponse(ctx, err, http.StatusNotFound, "Artworks not found")
			return
		}
		utils.GinErrorResponse(ctx, err, http.StatusInternalServerError, "Failed to get artworks by keyword")
		return
	}
	if len(artworks) == 0 {
		utils.GinErrorResponse(ctx, errs.ErrNotFoundArtworks, http.StatusNotFound, "Artworks not found")
		return
	}
	ctx.JSON(http.StatusOK, ResponseFromArtworks(artworks, ctx.GetBool("auth")))
}

func getArtworkListHybrid(ctx *gin.Context, queryText string, hybridSemanticRatio float64, offset, limit int64, r18Type types.R18Type, adapterOption ...*types.AdapterOption) {
	artworks, err := service.HybridSearchArtworks(ctx, queryText, hybridSemanticRatio, offset, limit, r18Type, adapterOption...)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			utils.GinErrorResponse(ctx, err, http.StatusNotFound, "Artworks not found")
			return
		}
		utils.GinErrorResponse(ctx, err, http.StatusInternalServerError, "Failed to hybrid search artworks")
		return
	}
	if len(artworks) == 0 {
		utils.GinErrorResponse(ctx, errs.ErrNotFoundArtworks, http.StatusNotFound, "Artworks not found")
		return
	}
	ctx.JSON(http.StatusOK, ResponseFromArtworks(artworks, ctx.GetBool("auth")))
}
