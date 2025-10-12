package runtimecfg

type RestConfig struct {
	Enable bool   `toml:"enable" mapstructure:"enable" json:"enable" yaml:"enable"`
	Addr   string `toml:"addr" mapstructure:"addr" json:"addr" yaml:"addr"`
	// rate limit
	Limit            LimiterConfig     `toml:"limit" mapstructure:"limit" json:"limit" yaml:"limit"`
	Site             SiteConfig        `toml:"site" mapstructure:"site" json:"site" yaml:"site"`
	StoragePathRules []StoragePathRule `toml:"storage_path_rule" mapstructure:"storage_path_rule" json:"storage_path_rule" yaml:"storage_path_rule"`
	GeoIPDB          string            `toml:"geoip_db" mapstructure:"geoip_db" json:"geoip_db" yaml:"geoip_db"`
}

type StoragePathRule struct {
	Path        string `toml:"path" mapstructure:"path" json:"path" yaml:"path"`
	StorageType string `toml:"storage_type" mapstructure:"storage_type" json:"storage_type" yaml:"storage_type"`
	TrimPrefix  string `toml:"trim_prefix" mapstructure:"trim_prefix" json:"trim_prefix" yaml:"trim_prefix"`
	JoinPrefix  string `toml:"join_prefix" mapstructure:"join_prefix" json:"join_prefix" yaml:"join_prefix"`
}

type SiteConfig struct {
	Title string `toml:"title" mapstructure:"title" json:"title" yaml:"title"`
	Desc  string `toml:"desc" mapstructure:"desc" json:"desc" yaml:"desc"`
	Name  string `toml:"name" mapstructure:"name" json:"name" yaml:"name"`
	Email string `toml:"email" mapstructure:"email" json:"email" yaml:"email"`
	URL   string `toml:"url" mapstructure:"url" json:"url" yaml:"url"`
}

type LimiterConfig struct {
	Enable bool `toml:"enable" mapstructure:"enable" json:"enable" yaml:"enable"`
	// in seconds
	Expiration int `toml:"expiration" mapstructure:"expiration" json:"expiration" yaml:"expiration"`
	Max        int `toml:"max" mapstructure:"max" json:"max" yaml:"max"`
}
