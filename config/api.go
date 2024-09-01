package config

type apiConfig struct {
	Enable         bool     `toml:"enable" mapstructure:"enable" json:"enable" yaml:"enable"`
	SiteURL        string   `toml:"site_url" mapstructure:"site_url" json:"site_url" yaml:"site_url"`
	Address        string   `toml:"address" mapstructure:"address" json:"address" yaml:"address"`
	Key            string   `toml:"key" mapstructure:"key" json:"key" yaml:"key"`
	AllowedOrigins []string `toml:"allowed_origins" mapstructure:"allowed_origins" json:"allowed_origins" yaml:"allowed_origins"`

	Realm              string `toml:"realm" mapstructure:"realm" json:"realm" yaml:"realm"`
	Secret             string `toml:"secret" mapstructure:"secret" json:"secret" yaml:"secret"`
	TokenExpire        int    `toml:"token_expire" mapstructure:"token_expire" json:"token_expire" yaml:"token_expire"`
	RefreshTokenExpire int    `toml:"refresh_token_expire" mapstructure:"refresh_token_expire" json:"refresh_token_expire" yaml:"refresh_token_expire"`
}
