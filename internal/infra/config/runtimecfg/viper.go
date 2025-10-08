package runtimecfg

import (
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/spf13/viper"
)

type Config struct {
	Migrate struct {
		Target string `toml:"target" mapstructure:"target" json:"target" yaml:"target"`
		DSN    string `toml:"dsn" mapstructure:"dsn" json:"dsn" yaml:"dsn"`
	} `toml:"migrate" mapstructure:"migrate" json:"migrate" yaml:"migrate"`

	App  AppConfig  `toml:"app" mapstructure:"app" json:"app" yaml:"app"`
	Wsrv WsrvConfig `toml:"wsrv" mapstructure:"wsrv" json:"wsrv" yaml:"wsrv"`
	Web  struct {
		Enable  bool   `toml:"enable" mapstructure:"enable" json:"enable" yaml:"enable"`
		Address string `toml:"address" mapstructure:"address" json:"address" yaml:"address"`
	} `toml:"web" mapstructure:"web" json:"web" yaml:"web"`
	API        apiConfig        `toml:"api" mapstructure:"api" json:"api" yaml:"api"`
	Auth       authConfig       `toml:"auth" mapstructure:"auth" json:"auth" yaml:"auth"`
	Log        logConfig        `toml:"log" mapstructure:"log" json:"log" yaml:"log"`
	Telegram   TelegramConfig   `toml:"telegram" mapstructure:"telegram" json:"telegram" yaml:"telegram"`
	HttpClient HttpClientConfig `toml:"http_client" mapstructure:"http_client" json:"http_client" yaml:"http_client"`

	// infrastructures config
	Cache    CacheConfig    `toml:"cache" mapstructure:"cache" json:"cache" yaml:"cache"`
	Search   SearchConfig   `toml:"search" mapstructure:"search" json:"search" yaml:"search"`
	Tagging  TaggingConfig  `toml:"tagging" mapstructure:"tagging" json:"tagging" yaml:"tagging"`
	Database databaseConfig `toml:"database" mapstructure:"database" json:"database" yaml:"database"`
	Source   SourceConfig   `toml:"source" mapstructure:"source" json:"source" yaml:"source"`
	Storage  StorageConfig  `toml:"storage" mapstructure:"storage" json:"storage" yaml:"storage"`
}

type AppConfig struct {
	// Something globally used in app
	Debug bool `toml:"debug" mapstructure:"debug" json:"debug" yaml:"debug"`
}

type WsrvConfig struct {
	URL string `toml:"url" mapstructure:"url" json:"url" yaml:"url"`
}

var (
	cfg      Config
	loadOnce sync.Once
)

func Get() Config {
	loadOnce.Do(func() {
		cfg = loadConfig()
	})
	return cfg
}

func loadConfig() Config {

	viper.SetConfigName("config")
	viper.AddConfigPath(".")
	viper.AddConfigPath("/etc/manyacg/")
	viper.SetConfigType("toml")
	viper.SetEnvPrefix("manyacg")
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	defaults := map[string]any{
		"wsrv.url": "https://wsrv.nl",

		"storage.cache_dir": "./imgcache",
		"storage.cache_ttl": 86400,

		"telegram.api_url":             "https://api.telegram.org",
		"telegram.retry.max_attempts":  5,
		"telegram.retry.exponent_base": 2.0,
		"telegram.retry.start_delay":   3,
		"telegram.retry.max_delay":     300,

		"storage.telegram.api_url":             "https://api.telegram.org",
		"storage.telegram.retry.max_attempts":  5,
		"storage.telegram.retry.exponent_base": 2.0,
		"storage.telegram.retry.start_delay":   3,
		"storage.telegram.retry.max_delay":     300,
	}

	for key, value := range defaults {
		viper.SetDefault(key, value)
	}

	// viper.SetDefault("api.address", "0.0.0.0:39080")
	// viper.SetDefault("api.site_name", "ManyACG")
	// viper.SetDefault("api.site_title", "ManyACG - ACG Picture Collection")
	// viper.SetDefault("api.site_description", "Many illustrations and pictures of ACG")
	// viper.SetDefault("api.allowed_origins", []string{"*"})
	// viper.SetDefault("api.realm", "ManyACG")
	// viper.SetDefault("api.token_expire", 86400*14)
	// viper.SetDefault("api.refresh_token_expire", 86400*30)
	// viper.SetDefault("api.geoip_db", "geoip.mmdb")

	// viper.SetDefault("fetcher.max_concurrent", 5)
	// viper.SetDefault("fetcher.limit", 50)

	// viper.SetDefault("log.file_path", "logs/manyacg.log")
	// viper.SetDefault("log.backup_num", 7)

	// viper.SetDefault("source.pixiv.enable", true)
	// viper.SetDefault("source.twitter.enable", true)
	// viper.SetDefault("source.bilibili.enable", true)
	// viper.SetDefault("source.danbooru.enable", true)
	// viper.SetDefault("source.kemono.enable", true)
	// viper.SetDefault("source.kemono.worker", 5)
	// viper.SetDefault("source.yandere.enable", true)
	// viper.SetDefault("source.nhentai.enable", true)
	// viper.SetDefault("source.pixiv.intervel", 60)
	// viper.SetDefault("source.pixiv.sleep", 1)
	// viper.SetDefault("source.twitter.fx_twitter_domain", "fxtwitter.com")
	// viper.SetDefault("source.twitter.sleep", 1)
	// viper.SetDefault("source.twitter.intervel", 60)

	// viper.SetDefault("storage.cache_dir", "./cache")
	// viper.SetDefault("storage.cache_ttl", 86400)
	// viper.SetDefault("storage.local.path", "./manyacg")
	// viper.SetDefault("storage.alist.token_expire", 86400)
	// viper.SetDefault("storage.regular_format", "webp")
	// viper.SetDefault("storage.thumb_format", "webp")
	// viper.SetDefault("storage.telegram.api_url", "https://api.telegram.org")

	// viper.SetDefault("telegram.sleep", 3)
	// viper.SetDefault("telegram.api_url", "https://api.telegram.org")
	// viper.SetDefault("telegram.retry.max_attempts", 5)
	// viper.SetDefault("telegram.retry.exponent_base", 2.0)
	// viper.SetDefault("telegram.retry.start_delay", 3)
	// viper.SetDefault("telegram.retry.max_delay", 300)

	// viper.SetDefault("storage.telegram.api_url", "https://api.telegram.org")
	// viper.SetDefault("storage.telegram.retry.max_attempts", 5)
	// viper.SetDefault("storage.telegram.retry.exponent_base", 2.0)
	// viper.SetDefault("storage.telegram.retry.start_delay", 3)
	// viper.SetDefault("storage.telegram.retry.max_delay", 300)

	// viper.SetDefault("database.database", "manyacg")
	// viper.SetDefault("database.max_staleness", 120)

	// viper.SetDefault("search.meilisearch.index", "manyacg")
	// viper.SetDefault("search.meilisearch.embedder", "default")

	if err := viper.ReadInConfig(); err != nil {
		fmt.Printf("error when reading config: %s\n", err)
		os.Exit(1)
	}
	c := Config{}
	if err := viper.Unmarshal(&c); err != nil {
		fmt.Printf("error when unmarshal config: %s\n", err)
		os.Exit(1)
	}

	if len(c.Telegram.Admins) == 0 {
		fmt.Println("please set at least one admin in config file (telegram.admins)")
		os.Exit(1)
	}
	return c
}
