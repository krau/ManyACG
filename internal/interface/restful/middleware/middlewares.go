package middleware

import (
	"errors"
	"net/http"

	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/krau/ManyACG/api/restful/utils"
	"github.com/krau/ManyACG/common"
	"github.com/krau/ManyACG/config"
	"github.com/krau/ManyACG/pkg/objectuuid"

	"github.com/krau/ManyACG/internal/service"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"

	"time"

	"github.com/chenyahui/gin-cache/persist"
	"github.com/go-redis/redis/v8"
)

var CacheStore persist.CacheStore

func Init() {
	// Init JWT
	var err error
	JWTAuthMiddleware, err = jwt.New(JWTInitParamas())
	if err != nil {
		common.Logger.Panicf("JWT init error: %v", err)
	}
	if err := JWTAuthMiddleware.MiddlewareInit(); err != nil {
		common.Logger.Panicf("JWT middleware init error: %v", err)
	}

	// Init Cache
	cacheConfig := config.Cfg.API.Cache
	if cacheConfig.Enable {
		if cacheConfig.Redis {
			opt, err := redis.ParseURL(cacheConfig.URL)
			if err != nil {
				common.Logger.Panicf("Failed to parse redis url: %v", err)
			}
			CacheStore = persist.NewRedisStore(redis.NewClient(opt))
		} else {
			CacheStore = persist.NewMemoryStore(time.Duration(cacheConfig.MemoryTTL) * time.Second)
		}
	}
}

func GetCacheDuration(route string) time.Duration {
	if CacheStore == nil {
		return time.Second
	}
	ttl, ok := config.Cfg.API.Cache.TTL[route]
	if !ok {
		return time.Second
	}
	return time.Duration(ttl) * time.Second
}

func CheckAdminKey(ctx *gin.Context) {
	if config.Cfg.Debug {
		ctx.Set("auth", true)
		ctx.Next()
		return
	}
	keyHeader := ctx.GetHeader("X-ADMIN-API-KEY")
	if keyHeader == config.Cfg.API.Key {
		ctx.Set("auth", true)
		ctx.Next()
		return
	}
	ctx.Set("auth", false)
	ctx.Next()
}

func AdminKeyRequired(ctx *gin.Context) {
	if ctx.GetBool("auth") {
		ctx.Next()
		return
	}
	ctx.JSON(http.StatusUnauthorized, gin.H{
		"status":  http.StatusUnauthorized,
		"message": "Unauthorized",
	})
	ctx.Abort()
}

func ValidatePictureID(ctx *gin.Context) {
	pictureID := ctx.Param("id")
	objectID, err := objectuuid.FromObjectIDHex(pictureID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"status":  http.StatusBadRequest,
			"message": "Invalid ID",
		})
		ctx.Abort()
		return
	}
	picture, err := service.GetPictureByID(ctx, objectID)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			ctx.JSON(http.StatusNotFound, gin.H{
				"status":  http.StatusNotFound,
				"message": "Picture not found",
			})
		} else {
			common.Logger.Errorf("Failed to get picture: %v", err)
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"status":  http.StatusInternalServerError,
				"message": "Failed to get picture",
			})
		}
		ctx.Abort()
		return
	}
	ctx.Set("picture", picture)
	ctx.Next()

}

func ValidateParamObjectID(ctx *gin.Context) {
	id := ctx.Param("id")
	objectID, err := objectuuid.FromObjectIDHex(id)
	if err != nil {
		utils.GinErrorResponse(ctx, err, http.StatusBadRequest, "Invalid ID")
		return
	}
	ctx.Set("object_id", objectID)
	ctx.Next()
}

func CheckApiKey(ctx *gin.Context) {
	key := ctx.GetHeader("X-API-KEY")
	if key == "" {
		utils.GinErrorResponse(ctx, errors.New("api key is required"), http.StatusUnauthorized, "Unauthorized")
		ctx.Abort()
		return
	}
	apiKey, err := service.GetApiKeyByKey(ctx, key)
	if err != nil {
		utils.GinErrorResponse(ctx, err, http.StatusUnauthorized, "Unauthorized")
		ctx.Abort()
		return
	}
	if apiKey.Used >= apiKey.Quota {
		utils.GinErrorResponse(ctx, errors.New("api key quota exceeded"), http.StatusForbidden, "Forbidden")
		ctx.Abort()
		return
	}
	ctx.Set("api_key", apiKey)
	ctx.Next()
}
