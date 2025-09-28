package runtimecfg

type taggerConfig struct {
	Enable  bool   `toml:"enable" mapstructure:"enable" json:"enable" yaml:"enable"`
	Host    string `toml:"host" mapstructure:"host" json:"host" yaml:"host"`
	Token   string `toml:"token" mapstructure:"token" json:"token" yaml:"token"`
	Timeout int    `toml:"timeout" mapstructure:"timeout" json:"timeout" yaml:"timeout"`
	TagNew  bool   `toml:"tagnew" mapstructure:"tagnew" json:"tagnew" yaml:"tagnew"`
}
