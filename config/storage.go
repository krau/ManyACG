package config

type storageConfigs struct {
	Type     string              `toml:"type" mapstructure:"type" json:"type" yaml:"type"`
	CacheDir string              `toml:"cache_dir" mapstructure:"cache_dir" json:"cache_dir" yaml:"cache_dir"`
	CacheTTL uint                `toml:"cache_ttl" mapstructure:"cache_ttl" json:"cache_ttl" yaml:"cache_ttl"`
	Webdav   StorageWebdavConfig `toml:"webdav" mapstructure:"webdav" json:"webdav" yaml:"webdav"`
}

type StorageWebdavConfig struct {
	URL      string `toml:"url" mapstructure:"url" json:"url" yaml:"url"`
	Username string `toml:"username" mapstructure:"username" json:"username" yaml:"username"`
	Password string `toml:"password" mapstructure:"password" json:"password" yaml:"password"`
	Path     string `toml:"path" mapstructure:"path" json:"path" yaml:"path"`
}
