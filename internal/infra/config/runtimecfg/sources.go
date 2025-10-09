package runtimecfg

type SourceConfig struct {
	Proxy    string               `toml:"proxy" mapstructure:"proxy" json:"proxy" yaml:"proxy"`
	Pixiv    SourcePixivConfig    `toml:"pixiv" mapstructure:"pixiv" json:"pixiv" yaml:"pixiv"`
	Twitter  SourceTwitterConfig  `toml:"twitter" mapstructure:"twitter" json:"twitter" yaml:"twitter"`
	Bilibili SourceBilibiliConfig `toml:"bilibili" mapstructure:"bilibili" json:"bilibili" yaml:"bilibili"`
	Danbooru SourceDanbooruConfig `toml:"danbooru" mapstructure:"danbooru" json:"danbooru" yaml:"danbooru"`
	Kemono   SourceKemonoConfig   `toml:"kemono" mapstructure:"kemono" json:"kemono" yaml:"kemono"`
	Yandere  SourceYandereConfig  `toml:"yandere" mapstructure:"yandere" json:"yandere" yaml:"yandere"`
	Nhentai  SourceNhentaiConfig  `toml:"nhentai" mapstructure:"nhentai" json:"nhentai" yaml:"nhentai"`
}

type SourcePixivConfig struct {
	Disable  bool           `toml:"disable" mapstructure:"disable" json:"disable" yaml:"disable"`
	Proxy    string         `toml:"proxy" mapstructure:"proxy" json:"proxy" yaml:"proxy"`
	URLs     []string       `toml:"urls" mapstructure:"urls" json:"urls" yaml:"urls"`
	Intervel int            `toml:"intervel" mapstructure:"intervel" json:"intervel" yaml:"intervel"`
	Sleep    uint           `toml:"sleep" mapstructure:"sleep" json:"sleep" yaml:"sleep"`
	Cookies  []cookieConfig `toml:"cookies" mapstructure:"cookies" json:"cookies" yaml:"cookies"`
}

type SourceTwitterConfig struct {
	Disable         bool     `toml:"disable" mapstructure:"disable" json:"disable" yaml:"disable"`
	FxTwitterDomain string   `toml:"fx_twitter_domain" mapstructure:"fx_twitter_domain" json:"fx_twitter_domain" yaml:"fx_twitter_domain"`
	Sleep           uint     `toml:"sleep" mapstructure:"sleep" json:"sleep" yaml:"sleep"`
	Intervel        int      `toml:"intervel" mapstructure:"intervel" json:"intervel" yaml:"intervel"`
	URLs            []string `toml:"urls" mapstructure:"urls" json:"urls" yaml:"urls"`
}

type SourceBilibiliConfig struct {
	Disable bool `toml:"disable" mapstructure:"disable" json:"disable" yaml:"disable"`
}

type SourceDanbooruConfig struct {
	Disable bool `toml:"disable" mapstructure:"disable" json:"disable" yaml:"disable"`
}

type SourceKemonoConfig struct {
	Disable bool `toml:"disable" mapstructure:"disable" json:"disable" yaml:"disable"`
}

type SourceYandereConfig struct {
	Disable bool `toml:"disable" mapstructure:"disable" json:"disable" yaml:"disable"`
}

type SourceNhentaiConfig struct {
	Disable bool `toml:"disable" mapstructure:"disable" json:"disable" yaml:"disable"`
}
