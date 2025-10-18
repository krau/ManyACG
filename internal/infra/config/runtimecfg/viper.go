package runtimecfg

import (
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/spf13/viper"
)

type Config struct {
	App AppConfig `toml:"app" mapstructure:"app" json:"app" yaml:"app"`
	// some common packages config
	Log        LogConfig        `toml:"log" mapstructure:"log" json:"log" yaml:"log"`
	HttpClient HttpClientConfig `toml:"http_client" mapstructure:"http_client" json:"http_client" yaml:"http_client"`

	// infrastructures config
	Cache    CacheConfig    `toml:"cache" mapstructure:"cache" json:"cache" yaml:"cache"`
	KVDB     KVDBConfig     `toml:"kvdb" mapstructure:"kvdb" json:"kvdb" yaml:"kvdb"`
	Search   SearchConfig   `toml:"search" mapstructure:"search" json:"search" yaml:"search"`
	Tagging  TaggingConfig  `toml:"tagging" mapstructure:"tagging" json:"tagging" yaml:"tagging"`
	Database databaseConfig `toml:"database" mapstructure:"database" json:"database" yaml:"database"`
	Source   SourceConfig   `toml:"source" mapstructure:"source" json:"source" yaml:"source"`
	Storage  StorageConfig  `toml:"storage" mapstructure:"storage" json:"storage" yaml:"storage"`
	Wsrv     WsrvConfig     `toml:"wsrv" mapstructure:"wsrv" json:"wsrv" yaml:"wsrv"`

	// interfaces
	Telegram  TelegramConfig  `toml:"telegram" mapstructure:"telegram" json:"telegram" yaml:"telegram"`
	Scheduler SchedulerConfig `toml:"scheduler" mapstructure:"scheduler" json:"scheduler" yaml:"scheduler"`
	Rest      RestConfig      `toml:"rest" mapstructure:"rest" json:"rest" yaml:"rest"`
}

type KVDBConfig struct {
	Type string `toml:"type" mapstructure:"type" json:"type" yaml:"type"` // "bbolt"
	Path string `toml:"path" mapstructure:"path" json:"path" yaml:"path"`
}

type SchedulerConfig struct {
	Enable   bool `toml:"enable" mapstructure:"enable" json:"enable" yaml:"enable"`
	Interval uint `toml:"interval" mapstructure:"interval" json:"interval" yaml:"interval"`
	Limit    int  `toml:"limit" mapstructure:"limit" json:"limit" yaml:"limit"` // 0 or negative means no limit
}

type AppConfig struct {
	// Something globally used in app
	Debug bool `toml:"debug" mapstructure:"debug" json:"debug" yaml:"debug"`
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
		"log.file_path":  "logs/manyacg.log",
		"log.backup_num": 7,

		"wsrv.url": "https://wsrv.nl",

		"telegram.api_url":             "https://api.telegram.org",
		"telegram.retry.max_attempts":  5,
		"telegram.retry.exponent_base": 2.0,
		"telegram.retry.start_delay":   3,
		"telegram.retry.max_delay":     300,

		"rest.site.title":        "ManyACG - Kawaii is all you need",
		"rest.site.desc":         "ACG Image Collector and Gallery Server",
		"rest.site.name":         "ManyACG",
		"rest.cache.default_ttl": 600, // 10 minutes

		"storage.telegram.api_url":             "https://api.telegram.org",
		"storage.telegram.retry.max_attempts":  5,
		"storage.telegram.retry.exponent_base": 2.0,
		"storage.telegram.retry.start_delay":   3,
		"storage.telegram.retry.max_delay":     300,
		"storage.regular_length":               2560,
		"storage.regular_format":               "webp",
		"storage.thumb_length":                 500,
		"storage.thumb_format":                 "avif",
		"storage.cache_dir":                    "./imgcache",
		"storage.cache_ttl":                    60 * 60 * 4, // in seconds

		"source.pixiv.img_proxy":           "pximg.manyacg.top",
		"source.twitter.fx_twitter_domain": "fxtwitter.com",

		"cache.default_ttl":            86400,
		"cache.ristretto.num_counters": 1e5,
		"cache.ristretto.max_cost":     1e6,

		"database.type": "sqlite",
		"database.dsn":  `file:manyacg.db?_pragma=journal_mode(WAL)&_pragma=synchronous(NORMAL)&_pragma=busy_timeout(5000)&_txlock=deferred`,
	}

	for key, value := range defaults {
		viper.SetDefault(key, value)
	}

	if err := viper.ReadInConfig(); err != nil {
		fmt.Printf("error when reading config: %s\n", err)
		os.Exit(1)
	}
	c := Config{}
	if err := viper.Unmarshal(&c); err != nil {
		fmt.Printf("error when unmarshal config: %s\n", err)
		os.Exit(1)
	}
	return c
}
