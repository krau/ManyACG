package runtimecfg

import (
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/spf13/viper"
)

type Config struct {
	Search     SearchConfig     `toml:"search" mapstructure:"search" json:"search" yaml:"search"`
	HttpClient HttpClientConfig `toml:"http_client" mapstructure:"http_client" json:"http_client" yaml:"http_client"`

	Wsrv WsrvConfig `toml:"wsrv" mapstructure:"wsrv" json:"wsrv" yaml:"wsrv"`

	// interfaces
	Telegram TelegramConfig `toml:"telegram" mapstructure:"telegram" json:"telegram" yaml:"telegram"`
	Source   SourceConfig   `toml:"source" mapstructure:"source" json:"source" yaml:"source"`
	Tagging  TaggingConfig  `toml:"tagging" mapstructure:"tagging" json:"tagging" yaml:"tagging"`
	Database databaseConfig `toml:"database" mapstructure:"database" json:"database" yaml:"database"`
	// some common packages config
	Log  LogConfig  `toml:"log" mapstructure:"log" json:"log" yaml:"log"`
	Rest RestConfig `toml:"rest" mapstructure:"rest" json:"rest" yaml:"rest"`

	// infrastructures config
	KVDB      KVDBConfig      `toml:"kvdb" mapstructure:"kvdb" json:"kvdb" yaml:"kvdb"`
	Storage   StorageConfig   `toml:"storage" mapstructure:"storage" json:"storage" yaml:"storage"`
	Scheduler SchedulerConfig `toml:"scheduler" mapstructure:"scheduler" json:"scheduler" yaml:"scheduler"`
	App       AppConfig       `toml:"app" mapstructure:"app" json:"app" yaml:"app"`
}

type KVDBConfig struct {
	Type      string `toml:"type" mapstructure:"type" json:"type" yaml:"type"` // bbolt, redis
	Path      string `toml:"path" mapstructure:"path" json:"path" yaml:"path"`
	Bucket    string `toml:"bucket" mapstructure:"bucket" json:"bucket" yaml:"bucket"`
	TTLBucket string `toml:"ttl_bucket" mapstructure:"ttl_bucket" json:"ttl_bucket" yaml:"ttl_bucket"`

	Redis          RedisConfig `toml:"redis" mapstructure:"redis" json:"redis" yaml:"redis"`
	TTLBatchLimit  int         `toml:"ttl_batch_limit" mapstructure:"ttl_batch_limit" json:"ttl_batch_limit" yaml:"ttl_batch_limit"`
	TTLSweepPeriod uint        `toml:"ttl_sweep_period" mapstructure:"ttl_sweep_period" json:"ttl_sweep_period" yaml:"ttl_sweep_period"` // in seconds

}

type RedisConfig struct {
	// URL like redis://user:pass@host:port/db?addr=... for cluster/sentinel
	URL      string `toml:"url" mapstructure:"url" json:"url" yaml:"url"`
	Prefix   string `toml:"prefix" mapstructure:"prefix" json:"prefix" yaml:"prefix"`
	Username string `toml:"username" mapstructure:"username" json:"username" yaml:"username"`
	Password string `toml:"password" mapstructure:"password" json:"password" yaml:"password"`
	// Addrs for standalone or cluster
	Addrs       []string `toml:"addrs" mapstructure:"addrs" json:"addrs" yaml:"addrs"`
	DB          int      `toml:"db" mapstructure:"db" json:"db" yaml:"db"`
	TLS         bool     `toml:"tls" mapstructure:"tls" json:"tls" yaml:"tls"`
	TLSInsecure bool     `toml:"tls_insecure" mapstructure:"tls_insecure" json:"tls_insecure" yaml:"tls_insecure"`
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

		"database.type": "sqlite",
		"database.dsn":  `file:manyacg.db?_pragma=journal_mode(WAL)&_pragma=synchronous(NORMAL)&_pragma=busy_timeout(5000)&_txlock=deferred`,

		"kvdb.type":             "bbolt",
		"kvdb.path":             "data/kvdb.bbolt",
		"kvdb.bucket":           "manyacg",
		"kvdb.ttl_bucket":       "manyacg_ttl",
		"kvdb.ttl_batch_limit":  1024,
		"kvdb.ttl_sweep_period": 60, // in seconds
		"kvdb.redis.prefix":     "manyacg:",
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
