package config

type storageConfigs struct {
	OriginalType  string              `toml:"original_type" mapstructure:"original_type" json:"original_type" yaml:"original_type"`
	RegularType   string              `toml:"regular_type" mapstructure:"regular_type" json:"regular_type" yaml:"regular_type"`
	RegularFormat string              `toml:"regular_format" mapstructure:"regular_format" json:"regular_format" yaml:"regular_format"`
	ThumbType     string              `toml:"thumb_type" mapstructure:"thumb_type" json:"thumb_type" yaml:"thumb_type"`
	ThumbFormat   string              `toml:"thumb_format" mapstructure:"thumb_format" json:"thumb_format" yaml:"thumb_format"`
	CacheDir      string              `toml:"cache_dir" mapstructure:"cache_dir" json:"cache_dir" yaml:"cache_dir"`
	CacheTTL      uint                `toml:"cache_ttl" mapstructure:"cache_ttl" json:"cache_ttl" yaml:"cache_ttl"`
	Rules         []storageRuleConfig `toml:"rules" mapstructure:"rules" json:"rules" yaml:"rules"`
	Webdav        StorageWebdavConfig `toml:"webdav" mapstructure:"webdav" json:"webdav" yaml:"webdav"`
	Local         StorageLocalConfig  `toml:"local" mapstructure:"local" json:"local" yaml:"local"`
	Alist         StorageAlistConfig  `toml:"alist" mapstructure:"alist" json:"alist" yaml:"alist"`
}

type storageRuleConfig struct {
	/*
		Match: 进行 与 匹配
		Replace: 依次进行替换
		example:
		  match: {storage_type: "webdav", path_prefix: "/onedrive"}
		  replace: {rewrite_storage: "local", trim_prefix: "/onedrive", join_prefix: "/local/manyacg"}
		此规则被应用后, storage 在获取 webdav 存储驱动下的以 /onedrive 开头的图片时, 会去寻找 local 存储驱动下的以 /local/manyacg 开头的图片(路径前缀被替换)
	*/

	// Match
	StorageType string `toml:"storage_type" mapstructure:"storage_type" json:"storage_type" yaml:"storage_type"`
	PathPrefix  string `toml:"path_prefix" mapstructure:"path_prefix" json:"path_prefix" yaml:"path_prefix"`

	// Replace
	RewriteStorage string `toml:"rewrite_storage" mapstructure:"rewrite_storage" json:"rewrite_storage" yaml:"rewrite_storage"`
	TrimPrefix     string `toml:"trim_prefix" mapstructure:"trim_prefix" json:"trim_prefix" yaml:"trim_prefix"`
	JoinPrefix     string `toml:"join_prefix" mapstructure:"join_prefix" json:"join_prefix" yaml:"join_prefix"`
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
	Username     string `toml:"username" mapstructure:"username" json:"username" yaml:"username"`
	Password     string `toml:"password" mapstructure:"password" json:"password" yaml:"password"`
	Path         string `toml:"path" mapstructure:"path" json:"path" yaml:"path"`
	PathPassword string `toml:"path_password" mapstructure:"path_password" json:"path_password" yaml:"path_password"`
	TokenExpire  int    `toml:"token_expire" mapstructure:"token_expire" json:"token_expire" yaml:"token_expire"`
}
