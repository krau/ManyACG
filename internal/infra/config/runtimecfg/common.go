package runtimecfg


type CookieConfig struct {
	Name  string `toml:"name" mapstructure:"name" json:"name" yaml:"name"`
	Value string `toml:"value" mapstructure:"value" json:"value" yaml:"value"`
}
