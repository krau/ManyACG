package config

type sourceConfigs struct {
	Proxy    string               `toml:"proxy" mapstructure:"proxy" json:"proxy" yaml:"proxy"`
	Pixiv    SourcePixivConfig    `toml:"pixiv" mapstructure:"pixiv" json:"pixiv" yaml:"pixiv"`
	Twitter  SourceTwitterConfig  `toml:"twitter" mapstructure:"twitter" json:"twitter" yaml:"twitter"`
	Bilibili SourceBilibiliConfig `toml:"bilibili" mapstructure:"bilibili" json:"bilibili" yaml:"bilibili"`
	Danbooru SourceDanbooruConfig `toml:"danbooru" mapstructure:"danbooru" json:"danbooru" yaml:"danbooru"`
	Kemono   SourceKemonoConfig   `toml:"kemono" mapstructure:"kemono" json:"kemono" yaml:"kemono"`
	Yandere  SourceYandereConfig  `toml:"yandere" mapstructure:"yandere" json:"yandere" yaml:"yandere"`
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
	Enable          bool     `toml:"enable" mapstructure:"enable" json:"enable" yaml:"enable"`
	FxTwitterDomain string   `toml:"fx_twitter_domain" mapstructure:"fx_twitter_domain" json:"fx_twitter_domain" yaml:"fx_twitter_domain"`
	Sleep           uint     `toml:"sleep" mapstructure:"sleep" json:"sleep" yaml:"sleep"`
	Intervel        int      `toml:"intervel" mapstructure:"intervel" json:"intervel" yaml:"intervel"`
	URLs            []string `toml:"urls" mapstructure:"urls" json:"urls" yaml:"urls"`
}

type SourceBilibiliConfig struct {
	Enable bool `toml:"enable" mapstructure:"enable" json:"enable" yaml:"enable"`
}

type SourceDanbooruConfig struct {
	Enable bool `toml:"enable" mapstructure:"enable" json:"enable" yaml:"enable"`
}

type SourceKemonoConfig struct {
	Enable  bool   `toml:"enable" mapstructure:"enable" json:"enable" yaml:"enable"`
	Session string `toml:"session" mapstructure:"session" json:"session" yaml:"session"`
}

type SourceYandereConfig struct {
	Enable bool `toml:"enable" mapstructure:"enable" json:"enable" yaml:"enable"`
}
