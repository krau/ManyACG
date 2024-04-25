package config

type sourceConfigs struct {
	Proxy   string              `toml:"proxy" mapstructure:"proxy" json:"proxy" yaml:"proxy"`
	Pixiv   SourcePixivConfig   `toml:"pixiv" mapstructure:"pixiv" json:"pixiv" yaml:"pixiv"`
	Twitter SourceTwitterConfig `toml:"twitter" mapstructure:"twitter" json:"twitter" yaml:"twitter"`
}

type SourcePixivConfig struct {
	Enable   bool           `toml:"enable" mapstructure:"enable" json:"enable" yaml:"enable"`
	Proxy    string         `toml:"proxy" mapstructure:"proxy" json:"proxy" yaml:"proxy"`
	URLs     []string       `toml:"urls" mapstructure:"urls" json:"urls" yaml:"urls"`
	Intervel int            `toml:"intervel" mapstructure:"intervel" json:"intervel" yaml:"intervel"`
	Sleep    uint           `toml:"sleep" mapstructure:"sleep" json:"sleep" yaml:"sleep"`
	Cookies  []cookieConfig `toml:"cookies" mapstructure:"cookies" json:"cookies" yaml:"cookies"`
}

type SourceTwitterConfig struct {
	Enable          bool   `toml:"enable" mapstructure:"enable" json:"enable" yaml:"enable"`
	FxTwitterDomain string `toml:"fx_twitter_domain" mapstructure:"fx_twitter_domain" json:"fx_twitter_domain" yaml:"fx_twitter_domain"`
}
