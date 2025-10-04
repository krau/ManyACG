package runtimecfg

type HttpClientConfig struct {
	Proxy string `toml:"proxy" mapstructure:"proxy" json:"proxy" yaml:"proxy"`
}
