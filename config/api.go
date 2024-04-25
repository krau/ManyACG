package config

type apiConfig struct {
	Enable  bool   `toml:"enable" mapstructure:"enable" json:"enable" yaml:"enable"`
	Address string `toml:"address" mapstructure:"address" json:"address" yaml:"address"`
	Auth    bool   `toml:"auth" mapstructure:"auth" json:"auth" yaml:"auth"`
	Token   string `toml:"token" mapstructure:"token" json:"token" yaml:"token"`
}
