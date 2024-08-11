package config

type apiConfig struct {
	Enable         bool     `toml:"enable" mapstructure:"enable" json:"enable" yaml:"enable"`
	Address        string   `toml:"address" mapstructure:"address" json:"address" yaml:"address"`
	Token          string   `toml:"token" mapstructure:"token" json:"token" yaml:"token"`
	AllowedOrigins []string `toml:"allowed_origins" mapstructure:"allowed_origins" json:"allowed_origins" yaml:"allowed_origins"`
}
