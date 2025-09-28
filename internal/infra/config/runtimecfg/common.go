package runtimecfg

type SourceCommonConfig struct {
	Enable   bool
	Intervel int
}

type cookieConfig struct {
	Name  string `toml:"name" mapstructure:"name" json:"name" yaml:"name"`
	Value string `toml:"value" mapstructure:"value" json:"value" yaml:"value"`
}
