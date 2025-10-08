package runtimecfg

type TaggingConfig struct {
	Enable     bool   `toml:"enable" mapstructure:"enable" json:"enable" yaml:"enable"`
	TagNew     bool   `toml:"tagnew" mapstructure:"tagnew" json:"tagnew" yaml:"tagnew"`
	Engine     string `toml:"engine" mapstructure:"engine" json:"engine" yaml:"engine"`
	Konatagger struct {
		Host    string `toml:"host" mapstructure:"host" json:"host" yaml:"host"`
		Token   string `toml:"token" mapstructure:"token" json:"token" yaml:"token"`
		Timeout int    `toml:"timeout" mapstructure:"timeout" json:"timeout" yaml:"timeout"`
	} `toml:"konatagger" mapstructure:"konatagger" json:"konatagger" yaml:"konatagger"`
}
