package runtimecfg

import (
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/spf13/viper"
)

type Config struct {
	Debug   bool   `toml:"debug" mapstructure:"debug" json:"debug" yaml:"debug"`
	WSRVURL string `toml:"wsrv_url" mapstructure:"wsrv_url" json:"wsrv_url" yaml:"wsrv_url"`
	Web     struct {
		Enable  bool   `toml:"enable" mapstructure:"enable" json:"enable" yaml:"enable"`
		Address string `toml:"address" mapstructure:"address" json:"address" yaml:"address"`
	} `toml:"web" mapstructure:"web" json:"web" yaml:"web"`
	API      apiConfig      `toml:"api" mapstructure:"api" json:"api" yaml:"api"`
	Auth     authConfig     `toml:"auth" mapstructure:"auth" json:"auth" yaml:"auth"`
	Fetcher  fetcherConfig  `toml:"fetcher" mapstructure:"fetcher" json:"fetcher" yaml:"fetcher"`
	Log      logConfig      `toml:"log" mapstructure:"log" json:"log" yaml:"log"`
	Source   sourceConfigs  `toml:"source" mapstructure:"source" json:"source" yaml:"source"`
	Storage  storageConfigs `toml:"storage" mapstructure:"storage" json:"storage" yaml:"storage"`
	Telegram telegramConfig `toml:"telegram" mapstructure:"telegram" json:"telegram" yaml:"telegram"`
	Database databaseConfig `toml:"database" mapstructure:"database" json:"database" yaml:"database"`
	Search   searchConfig   `toml:"search" mapstructure:"search" json:"search" yaml:"search"`
	Tagger   taggerConfig   `toml:"tagger" mapstructure:"tagger" json:"tagger" yaml:"tagger"`

	Mirate migrateConfig `toml:"migrate" mapstructure:"migrate" json:"migrate" yaml:"migrate"`
}

type migrateConfig struct {
	Target string `toml:"target" mapstructure:"target" json:"target" yaml:"target"` // pgsql, mysql, sqlite
	DSN    string `toml:"dsn" mapstructure:"dsn" json:"dsn" yaml:"dsn"`
}

type fetcherConfig struct {
	MaxConcurrent int `toml:"max_concurrent" mapstructure:"max_concurrent" json:"max_concurrent" yaml:"max_concurrent"`
	Limit         int `toml:"limit" mapstructure:"limit" json:"limit" yaml:"limit"`
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

	viper.SetDefault("wsrv_url", "https://wsrv.nl")

	viper.SetDefault("api.address", "0.0.0.0:39080")
	viper.SetDefault("api.site_name", "ManyACG")
	viper.SetDefault("api.site_title", "ManyACG - ACG Picture Collection")
	viper.SetDefault("api.site_description", "Many illustrations and pictures of ACG")
	viper.SetDefault("api.allowed_origins", []string{"*"})
	viper.SetDefault("api.realm", "ManyACG")
	viper.SetDefault("api.token_expire", 86400*14)
	viper.SetDefault("api.refresh_token_expire", 86400*30)
	viper.SetDefault("api.geoip_db", "geoip.mmdb")

	viper.SetDefault("fetcher.max_concurrent", 5)
	viper.SetDefault("fetcher.limit", 50)

	viper.SetDefault("log.level", "TRACE")
	viper.SetDefault("log.file_path", "logs/ManyACG.log")
	viper.SetDefault("log.backup_num", 7)

	viper.SetDefault("source.pixiv.enable", true)
	viper.SetDefault("source.twitter.enable", true)
	viper.SetDefault("source.bilibili.enable", true)
	viper.SetDefault("source.danbooru.enable", true)
	viper.SetDefault("source.kemono.enable", true)
	viper.SetDefault("source.kemono.worker", 5)
	viper.SetDefault("source.yandere.enable", true)
	viper.SetDefault("source.nhentai.enable", true)
	viper.SetDefault("source.pixiv.intervel", 60)
	viper.SetDefault("source.pixiv.sleep", 1)
	viper.SetDefault("source.twitter.fx_twitter_domain", "fxtwitter.com")
	viper.SetDefault("source.twitter.sleep", 1)
	viper.SetDefault("source.twitter.intervel", 60)

	viper.SetDefault("storage.cache_dir", "./cache")
	viper.SetDefault("storage.cache_ttl", 86400)
	viper.SetDefault("storage.local.path", "./manyacg")
	viper.SetDefault("storage.alist.token_expire", 86400)
	viper.SetDefault("storage.regular_format", "webp")
	viper.SetDefault("storage.thumb_format", "webp")
	viper.SetDefault("storage.telegram.api_url", "https://api.telegram.org")

	viper.SetDefault("telegram.sleep", 3)
	viper.SetDefault("telegram.api_url", "https://api.telegram.org")
	viper.SetDefault("telegram.retry.max_attempts", 5)
	viper.SetDefault("telegram.retry.exponent_base", 2.0)
	viper.SetDefault("telegram.retry.start_delay", 3)
	viper.SetDefault("telegram.retry.max_delay", 300)

	viper.SetDefault("storage.telegram.api_url", "https://api.telegram.org")
	viper.SetDefault("storage.telegram.retry.max_attempts", 5)
	viper.SetDefault("storage.telegram.retry.exponent_base", 2.0)
	viper.SetDefault("storage.telegram.retry.start_delay", 3)
	viper.SetDefault("storage.telegram.retry.max_delay", 300)

	viper.SetDefault("database.database", "manyacg")
	viper.SetDefault("database.max_staleness", 120)

	viper.SetDefault("search.meilisearch.index", "manyacg")
	viper.SetDefault("search.meilisearch.embedder", "default")

	if err := viper.ReadInConfig(); err != nil {
		fmt.Printf("error when reading config: %s\n", err)
		os.Exit(1)
	}
	c := Config{}
	if err := viper.Unmarshal(c); err != nil {
		fmt.Printf("error when unmarshal config: %s\n", err)
		os.Exit(1)
	}

	if len(c.Telegram.Admins) == 0 {
		fmt.Println("please set at least one admin in config file (telegram.admins)")
		os.Exit(1)
	}
	return c
}
