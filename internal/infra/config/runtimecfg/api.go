package runtimecfg

type apiConfig struct {
	Enable  bool   `toml:"enable" mapstructure:"enable" json:"enable" yaml:"enable"`
	Metrics bool   `toml:"metrics" mapstructure:"metrics" json:"metrics" yaml:"metrics"`
	Address string `toml:"address" mapstructure:"address" json:"address" yaml:"address"`
	Key     string `toml:"key" mapstructure:"key" json:"key" yaml:"key"`
	MustKey bool   `toml:"must_key" mapstructure:"must_key" json:"must_key" yaml:"must_key"`

	SiteURL         string `toml:"site_url" mapstructure:"site_url" json:"site_url" yaml:"site_url"`
	SiteName        string `toml:"site_name" mapstructure:"site_name" json:"site_name" yaml:"site_name"`
	SiteTitle       string `toml:"site_title" mapstructure:"site_title" json:"site_title" yaml:"site_title"`
	SiteDescription string `toml:"site_description" mapstructure:"site_description" json:"site_description" yaml:"site_description"`
	SiteEmail       string `toml:"site_email" mapstructure:"site_email" json:"site_email" yaml:"site_email"`

	AllowedOrigins     []string `toml:"allowed_origins" mapstructure:"allowed_origins" json:"allowed_origins" yaml:"allowed_origins"`
	Realm              string   `toml:"realm" mapstructure:"realm" json:"realm" yaml:"realm"`
	Secret             string   `toml:"secret" mapstructure:"secret" json:"secret" yaml:"secret"`
	TokenExpire        int      `toml:"token_expire" mapstructure:"token_expire" json:"token_expire" yaml:"token_expire"`
	RefreshTokenExpire int      `toml:"refresh_token_expire" mapstructure:"refresh_token_expire" json:"refresh_token_expire" yaml:"refresh_token_expire"`

	PathRules []ApiPathRule `toml:"path_rules" mapstructure:"path_rules" json:"path_rules" yaml:"path_rules"`

	Cache   apiCacheConfig `toml:"cache" mapstructure:"cache" json:"cache" yaml:"cache"`
	GeoIPDB string         `toml:"geoip_db" mapstructure:"geoip_db" json:"geoip_db" yaml:"geoip_db"`
}

type ApiPathRule struct {
	Path        string `toml:"path" mapstructure:"path" json:"path" yaml:"path"`
	StorageType string `toml:"storage_type" mapstructure:"storage_type" json:"storage_type" yaml:"storage_type"`
	TrimPrefix  string `toml:"trim_prefix" mapstructure:"trim_prefix" json:"trim_prefix" yaml:"trim_prefix"`
	JoinPrefix  string `toml:"join_prefix" mapstructure:"join_prefix" json:"join_prefix" yaml:"join_prefix"`
}

type apiCacheConfig struct {
	Enable    bool           `toml:"enable" mapstructure:"enable" json:"enable" yaml:"enable"`
	Redis     bool           `toml:"redis" mapstructure:"redis" json:"redis" yaml:"redis"`
	URL       string         `toml:"url" mapstructure:"url" json:"url" yaml:"url"`
	TTL       map[string]int `toml:"ttl" mapstructure:"ttl" json:"ttl" yaml:"ttl"`
	MemoryTTL int            `toml:"memory_ttl" mapstructure:"memory_ttl" json:"memory_ttl" yaml:"memory_ttl"`
}
