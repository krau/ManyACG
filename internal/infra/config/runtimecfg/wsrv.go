package runtimecfg

type WsrvConfig struct {
	URL string `toml:"url" mapstructure:"url" json:"url" yaml:"url"`
}
