package routers

import (
	_ "embed"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/duke-git/lancet/v2/fileutil"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/feeds"
	"github.com/krau/ManyACG/adapter"
	"github.com/krau/ManyACG/api/restful/utils"
	"github.com/krau/ManyACG/common"
	"github.com/krau/ManyACG/config"
	"github.com/oschwald/geoip2-golang"

	"github.com/krau/ManyACG/service"
	"github.com/krau/ManyACG/types"
)

var (
	geoipReader *geoip2.Reader
)

func Init() {
	var err error
	if config.Cfg.API.GeoIPDB == "" {
		common.Logger.Warn("GeoIP database path is not set, skipping GeoIP initialization")
		return
	}
	if !fileutil.IsExist(config.Cfg.API.GeoIPDB) {
		common.Logger.Errorf("GeoIP database file does not exist: %s", config.Cfg.API.GeoIPDB)
		return
	}

	dbData, err := os.ReadFile(config.Cfg.API.GeoIPDB)
	if err != nil {
		common.Logger.Errorf("Failed to read GeoIP database file: %v", err)
		return
	}

	geoipReader, err = geoip2.FromBytes(dbData)
	if err != nil {
		common.Logger.Errorf("Failed to open GeoIP database: %v", err)
	}
}

func GenerateAtom(ctx *gin.Context) {
	artworks, err := service.GetLatestArtworks(ctx, types.R18TypeNone, 1, 50)
	if err != nil {
		utils.GinErrorResponse(ctx, err, http.StatusInternalServerError, "Failed to get artworks")
		return
	}
	feed := &feeds.Feed{
		Title:       config.Cfg.API.SiteTitle,
		Link:        &feeds.Link{Href: config.Cfg.API.SiteURL},
		Description: config.Cfg.API.SiteDescription,
		Author:      &feeds.Author{Name: config.Cfg.API.SiteName, Email: config.Cfg.API.SiteEmail},
		Created:     time.Now(),
		Items:       adapter.ConvertToFeedItems(ctx, artworks),
	}
	atom, err := feed.ToAtom()
	if err != nil {
		common.Logger.Errorf("Failed to generate Atom feed: %v", err)
		utils.GinErrorResponse(ctx, err, http.StatusInternalServerError, "Failed to generate Atom feed")
		return
	}
	ctx.Data(http.StatusOK, "application/xml", []byte(atom))
}

func MyIP(ctx *gin.Context) {
	ip := ctx.ClientIP()
	if geoipReader == nil {
		ctx.JSON(http.StatusOK, gin.H{
			"ip":          ip,
			"country":     "无法获取",
			"countryName": "无法获取",
		})
		return
	}
	record, err := geoipReader.Country(net.ParseIP(ip))
	if err != nil {
		common.Logger.Errorf("Failed to get geoip record: %v", err)
		utils.GinErrorResponse(ctx, err, http.StatusInternalServerError, "Failed to get geoip record")
		return
	}
	country := record.Country.IsoCode
	countryName := record.Country.Names["zh-CN"]
	if countryName == "" {
		countryName = record.Country.Names["en"]
	}
	if countryName == "" {
		countryName = "未知"
	}
	common.Logger.Infof("IP: %s, Country: %s (%s)", ip, country, countryName)
	ctx.JSON(http.StatusOK, gin.H{
		"ip":          ip,
		"country":     country,
		"countryName": countryName,
	})
}
