package config

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Debug    bool           `toml:"debug" mapstructure:"debug" json:"debug" yaml:"debug"`
	API      apiConfig      `toml:"api" mapstructure:"api" json:"api" yaml:"api"`
	Auth     authConfig     `toml:"auth" mapstructure:"auth" json:"auth" yaml:"auth"`
	Fetcher  fetcherConfig  `toml:"fetcher" mapstructure:"fetcher" json:"fetcher" yaml:"fetcher"`
	Log      logConfig      `toml:"log" mapstructure:"log" json:"log" yaml:"log"`
	Source   sourceConfigs  `toml:"source" mapstructure:"source" json:"source" yaml:"source"`
	Storage  storageConfigs `toml:"storage" mapstructure:"storage" json:"storage" yaml:"storage"`
	Telegram telegramConfig `toml:"telegram" mapstructure:"telegram" json:"telegram" yaml:"telegram"`
	Database databaseConfig `toml:"database" mapstructure:"database" json:"database" yaml:"database"`
}

type fetcherConfig struct {
	MaxConcurrent int `toml:"max_concurrent" mapstructure:"max_concurrent" json:"max_concurrent" yaml:"max_concurrent"`
	Limit         int `toml:"limit" mapstructure:"limit" json:"limit" yaml:"limit"`
}

var Cfg *Config

func InitConfig() {
	viper.SetConfigName("config")
	viper.AddConfigPath(".")
	viper.SetConfigType("toml")

	viper.SetDefault("api.address", "0.0.0.0:39080")
	viper.SetDefault("api.site_name", "ManyACG")
	viper.SetDefault("api.site_title", "ManyACG - ACG Picture Collection")
	viper.SetDefault("api.site_description", "Many illustrations and pictures of ACG")
	viper.SetDefault("api.site_email", "acg@unv.app")
	viper.SetDefault("api.allowed_origins", []string{"*"})
	viper.SetDefault("api.realm", "ManyACG")
	viper.SetDefault("api.token_expire", 86400*14)
	viper.SetDefault("api.refresh_token_expire", 86400*30)

	viper.SetDefault("fetcher.max_concurrent", 5)
	viper.SetDefault("fetcher.limit", 50)

	viper.SetDefault("log.level", "TRACE")
	viper.SetDefault("log.file_path", "logs/ManyACG.log")
	viper.SetDefault("log.backup_num", 7)

	viper.SetDefault("source.twitter.fx_twitter_domain", "fxtwitter.com")

	viper.SetDefault("storage.cache_dir", "./cache")
	viper.SetDefault("storage.cache_ttl", 86400)
	viper.SetDefault("storage.original_type", "local")
	viper.SetDefault("storage.regular_type", "local")
	viper.SetDefault("storage.thumb_type", "local")
	viper.SetDefault("storage.local.enable", true)
	viper.SetDefault("storage.local.path", "./manyacg")
	viper.SetDefault("storage.alist.token_expire", 86400)

	viper.SetDefault("telegram.sleep", 1)
	viper.SetDefault("telegram.api_url", "https://api.telegram.org")
	viper.SetDefault("telegram.retry.max_attempts", 5)
	viper.SetDefault("telegram.retry.exponent_base", 2.0)
	viper.SetDefault("telegram.retry.start_delay", 2*time.Second)
	viper.SetDefault("telegram.retry.max_delay", 1*time.Minute)

	viper.SetDefault("database.database", "manyacg")
	viper.SetDefault("database.max_staleness", 120)

	if err := viper.ReadInConfig(); err != nil {
		fmt.Printf("error when reading config: %s\n", err)
		os.Exit(1)
	}
	Cfg = &Config{}
	if err := viper.Unmarshal(Cfg); err != nil {
		fmt.Printf("error when unmarshal config: %s\n", err)
		os.Exit(1)
	}

	if len(Cfg.Telegram.Admins) == 0 {
		fmt.Println("please set at least one admin in config file (telegram.admins)")
		os.Exit(1)
	}
}
