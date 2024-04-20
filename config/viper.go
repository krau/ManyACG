package config

import "github.com/spf13/viper"

type Config struct {
	API      apiConfig      `toml:"api" mapstructure:"api" json:"api" yaml:"api"`
	Fetcher  fetcherConfig  `toml:"fetcher" mapstructure:"fetcher" json:"fetcher" yaml:"fetcher"`
	Log      logConfig      `toml:"log" mapstructure:"log" json:"log" yaml:"log"`
	Source   sourceConfigs  `toml:"source" mapstructure:"source" json:"source" yaml:"source"`
	Storage  storageConfigs `toml:"storage" mapstructure:"storage" json:"storage" yaml:"storage"`
	Telegram telegramConfig `toml:"telegram" mapstructure:"telegram" json:"telegram" yaml:"telegram"`
	Database databaseConfig `toml:"database" mapstructure:"database" json:"database" yaml:"database"`
}

type apiConfig struct {
	Enable  bool   `toml:"enable" mapstructure:"enable" json:"enable" yaml:"enable"`
	Address string `toml:"address" mapstructure:"address" json:"address" yaml:"address"`
	Auth    bool   `toml:"auth" mapstructure:"auth" json:"auth" yaml:"auth"`
	Token   string `toml:"token" mapstructure:"token" json:"token" yaml:"token"`
}

type fetcherConfig struct {
	MaxConcurrent int `toml:"max_concurrent" mapstructure:"max_concurrent" json:"max_concurrent" yaml:"max_concurrent"`
	Limit         int `toml:"limit" mapstructure:"limit" json:"limit" yaml:"limit"`
}

type logConfig struct {
	Level     string `toml:"level" mapstructure:"level" json:"level" yaml:"level"`
	FilePath  string `toml:"file_path" mapstructure:"file_path" json:"file_path" yaml:"file_path"`
	BackupNum uint   `toml:"backup_num" mapstructure:"backup_num" json:"backup_num" yaml:"backup_num"`
}

type sourceConfigs struct {
	Pixiv SourcePixivConfig `toml:"pixiv" mapstructure:"pixiv" json:"pixiv" yaml:"pixiv"`
}

type SourcePixivConfig struct {
	Enable   bool     `toml:"enable" mapstructure:"enable" json:"enable" yaml:"enable"`
	Proxy    string   `toml:"proxy" mapstructure:"proxy" json:"proxy" yaml:"proxy"`
	URLs     []string `toml:"urls" mapstructure:"urls" json:"urls" yaml:"urls"`
	Intervel uint     `toml:"intervel" mapstructure:"intervel" json:"intervel" yaml:"intervel"`
	Cookies  string   `toml:"cookies" mapstructure:"cookies" json:"cookies" yaml:"cookies"`
}

type storageConfigs struct {
	Type   string              `toml:"type" mapstructure:"type" json:"type" yaml:"type"`
	Webdav StorageWebdavConfig `toml:"webdav" mapstructure:"webdav" json:"webdav" yaml:"webdav"`
}

type StorageWebdavConfig struct {
	URL      string `toml:"url" mapstructure:"url" json:"url" yaml:"url"`
	Username string `toml:"username" mapstructure:"username" json:"username" yaml:"username"`
	Password string `toml:"password" mapstructure:"password" json:"password" yaml:"password"`
	Path     string `toml:"path" mapstructure:"path" json:"path" yaml:"path"`
	CacheDir string `toml:"cache_dir" mapstructure:"cache_dir" json:"cache_dir" yaml:"cache_dir"`
	CacheTTL uint   `toml:"cache_ttl" mapstructure:"cache_ttl" json:"cache_ttl" yaml:"cache_ttl"`
}

type telegramConfig struct {
	Token    string  `toml:"token" mapstructure:"token" json:"token" yaml:"token"`
	ChatID   int64   `toml:"chat_id" mapstructure:"chat_id" json:"chat_id" yaml:"chat_id"`
	Username string  `toml:"username" mapstructure:"username" json:"username" yaml:"username"`
	Sleep    uint    `toml:"sleep" mapstructure:"sleep" json:"sleep" yaml:"sleep"`
	Admins   []int64 `toml:"admins" mapstructure:"admins" json:"admins" yaml:"admins"`
}

type databaseConfig struct {
	URI      string `toml:"uri" mapstructure:"uri" json:"uri" yaml:"uri"`
	Host     string `toml:"host" mapstructure:"host" json:"host" yaml:"host"`
	Port     int    `toml:"port" mapstructure:"port" json:"port" yaml:"port"`
	User     string `toml:"user" mapstructure:"user" json:"user" yaml:"user"`
	Password string `toml:"password" mapstructure:"password" json:"password" yaml:"password"`
	Database string `toml:"database"`
}

var Cfg *Config

func init() {
	viper.SetConfigName("config")
	viper.AddConfigPath(".")
	viper.SetConfigType("toml")

	viper.SetDefault("fetcher.max_concurrent", 5)
	viper.SetDefault("fetcher.limit", 30)

	viper.SetDefault("log.level", "TRACE")
	viper.SetDefault("log.file_path", "logs/ManyACG-Bot.log")
	viper.SetDefault("log.backup_num", 7)

	viper.SetDefault("Database.databse", "manyacg")

	if err := viper.ReadInConfig(); err != nil {
		panic(err)
	}
	Cfg = &Config{}
	if err := viper.Unmarshal(Cfg); err != nil {
		panic(err)
	}
}
