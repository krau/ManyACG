package handlers

import (
	"net"
	"os"
	"sync"

	"github.com/duke-git/lancet/v2/fileutil"
	"github.com/gofiber/fiber/v3"
	"github.com/krau/ManyACG/internal/infra/config/runtimecfg"
	"github.com/krau/ManyACG/internal/interface/rest/common"
	"github.com/krau/ManyACG/pkg/log"
	"github.com/oschwald/geoip2-golang"
)

var (
	loadIpOnce  sync.Once
	geoipReader *geoip2.Reader
)

func MyIP(cfg runtimecfg.RestConfig) fiber.Handler {
	loadIpOnce.Do(func() {
		if cfg.GeoIPDB == "" {
			log.Warn("GeoIP database path is not set, skipping GeoIP initialization")
			return
		}
		if !fileutil.IsExist(cfg.GeoIPDB) {
			log.Warn("GeoIP database file does not exist, skipping GeoIP initialization", "path", cfg.GeoIPDB)
			return
		}
		dbData, err := os.ReadFile(cfg.GeoIPDB)
		if err != nil {
			log.Error("failed to read GeoIP database file, skipping GeoIP initialization", "path", cfg.GeoIPDB, "err", err)
			return
		}
		geoipReader, err = geoip2.FromBytes(dbData)
		if err != nil {
			log.Error("failed to load GeoIP database, skipping GeoIP initialization", "path", cfg.GeoIPDB, "err", err)
			return
		}
		log.Debug("GeoIP database loaded", "path", cfg.GeoIPDB)
	})
	return func(ctx fiber.Ctx) error {
		ip := ctx.IP()
		if geoipReader == nil {
			ctx.JSON(fiber.Map{
				"ip":          ip,
				"country":     "unknown",
				"countryName": "unknown",
			})
			return nil
		}
		record, err := geoipReader.Country(net.ParseIP(ip))
		if err != nil {
			return common.NewError(fiber.StatusInternalServerError, "failed to lookup GeoIP record")
		}
		country := record.Country.IsoCode
		countryName := record.Country.Names["zh-CN"]
		if countryName == "" {
			countryName = record.Country.Names["en"]
		}
		if countryName == "" {
			countryName = "未知"
		}
		log.Infof("IP: %s, Country: %s (%s)", ip, country, countryName)
		ctx.JSON(fiber.Map{
			"ip":          ip,
			"country":     country,
			"countryName": countryName,
		})
		return nil
	}
}
