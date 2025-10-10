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

	App AppConfig `toml:"app" mapstructure:"app" json:"app" yaml:"app"`
	Web struct {
		Enable  bool   `toml:"enable" mapstructure:"enable" json:"enable" yaml:"enable"`
		Address string `toml:"address" mapstructure:"address" json:"address" yaml:"address"`
	} `toml:"web" mapstructure:"web" json:"web" yaml:"web"`
	API  apiConfig  `toml:"api" mapstructure:"api" json:"api" yaml:"api"`
	Auth authConfig `toml:"auth" mapstructure:"auth" json:"auth" yaml:"auth"`

	// some common packages config
	Log        LogConfig        `toml:"log" mapstructure:"log" json:"log" yaml:"log"`
	HttpClient HttpClientConfig `toml:"http_client" mapstructure:"http_client" json:"http_client" yaml:"http_client"`

	// infrastructures config
	Cache    CacheConfig    `toml:"cache" mapstructure:"cache" json:"cache" yaml:"cache"`
	Search   SearchConfig   `toml:"search" mapstructure:"search" json:"search" yaml:"search"`
	Tagging  TaggingConfig  `toml:"tagging" mapstructure:"tagging" json:"tagging" yaml:"tagging"`
	Database databaseConfig `toml:"database" mapstructure:"database" json:"database" yaml:"database"`
	Source   SourceConfig   `toml:"source" mapstructure:"source" json:"source" yaml:"source"`
	Storage  StorageConfig  `toml:"storage" mapstructure:"storage" json:"storage" yaml:"storage"`
	Wsrv     WsrvConfig     `toml:"wsrv" mapstructure:"wsrv" json:"wsrv" yaml:"wsrv"`

	// interfaces
	Telegram  TelegramConfig  `toml:"telegram" mapstructure:"telegram" json:"telegram" yaml:"telegram"`
	Scheduler SchedulerConfig `toml:"scheduler" mapstructure:"scheduler" json:"scheduler" yaml:"scheduler"`
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
		"storage.regular_length":               2560,
		"storage.regular_format":               "webp",
		"storage.thumb_length":                 1280,
		"storage.thumb_format":                 "avif",
		"storage.cache_dir":                    "./imgcache",
		"storage.cache_ttl":                    86400,

		"cache.type":                   "ristretto",
		"cache.default_ttl":            60 * 60,
		"cache.ristretto.num_counters": 1e5,
		"cache.ristretto.max_cost":     1e6,
		"cache.ristretto.buffer_items": 64,

		"database.type": "sqlite",
		"database.dsn":  "file:data/manyacg.db?_pragma=journal_mode(WAL)&_pragma=synchronous(NORMAL)&_pragma=busy_timeout(5000)&_txlock=deferred",
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
