package config

type storageConfigs struct {
	OriginalType string              `toml:"original_type" mapstructure:"original_type" json:"original_type" yaml:"original_type"`
	RegularType  string              `toml:"regular_type" mapstructure:"regular_type" json:"regular_type" yaml:"regular_type"`
	ThumbType    string              `toml:"thumb_type" mapstructure:"thumb_type" json:"thumb_type" yaml:"thumb_type"`
	CacheDir     string              `toml:"cache_dir" mapstructure:"cache_dir" json:"cache_dir" yaml:"cache_dir"`
	CacheTTL     uint                `toml:"cache_ttl" mapstructure:"cache_ttl" json:"cache_ttl" yaml:"cache_ttl"`
	Webdav       StorageWebdavConfig `toml:"webdav" mapstructure:"webdav" json:"webdav" yaml:"webdav"`
	Local        StorageLocalConfig  `toml:"local" mapstructure:"local" json:"local" yaml:"local"`
	Alist        StorageAlistConfig  `toml:"alist" mapstructure:"alist" json:"alist" yaml:"alist"`
}

type StorageWebdavConfig struct {
	Enable   bool   `toml:"enable" mapstructure:"enable" json:"enable" yaml:"enable"`
	URL      string `toml:"url" mapstructure:"url" json:"url" yaml:"url"`
	Username string `toml:"username" mapstructure:"username" json:"username" yaml:"username"`
	Password string `toml:"password" mapstructure:"password" json:"password" yaml:"password"`
	Path     string `toml:"path" mapstructure:"path" json:"path" yaml:"path"`
}

type StorageLocalConfig struct {
	Enable bool   `toml:"enable" mapstructure:"enable" json:"enable" yaml:"enable"`
	Path   string `toml:"path" mapstructure:"path" json:"path" yaml:"path"`
}

type StorageAlistConfig struct {
	Enable       bool   `toml:"enable" mapstructure:"enable" json:"enable" yaml:"enable"`
	URL          string `toml:"url" mapstructure:"url" json:"url" yaml:"url"`
	CdnURL       string `toml:"cdn_url" mapstructure:"cdn_url" json:"cdn_url" yaml:"cdn_url"`
	Username     string `toml:"username" mapstructure:"username" json:"username" yaml:"username"`
	Password     string `toml:"password" mapstructure:"password" json:"password" yaml:"password"`
	Path         string `toml:"path" mapstructure:"path" json:"path" yaml:"path"`
	PathPassword string `toml:"path_password" mapstructure:"path_password" json:"path_password" yaml:"path_password"`
	TokenExpire  int    `toml:"token_expire" mapstructure:"token_expire" json:"token_expire" yaml:"token_expire"`
}
