package runtimecfg

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type SourceConfig struct {
	Twitter  SourceTwitterConfig  `toml:"twitter" mapstructure:"twitter" json:"twitter" yaml:"twitter"`
	Proxy    string               `toml:"proxy" mapstructure:"proxy" json:"proxy" yaml:"proxy"`
	Pixiv    SourcePixivConfig    `toml:"pixiv" mapstructure:"pixiv" json:"pixiv" yaml:"pixiv"`
	Bilibili SourceBilibiliConfig `toml:"bilibili" mapstructure:"bilibili" json:"bilibili" yaml:"bilibili"`
	Danbooru SourceDanbooruConfig `toml:"danbooru" mapstructure:"danbooru" json:"danbooru" yaml:"danbooru"`
	Kemono   SourceKemonoConfig   `toml:"kemono" mapstructure:"kemono" json:"kemono" yaml:"kemono"`
	Yandere  SourceYandereConfig  `toml:"yandere" mapstructure:"yandere" json:"yandere" yaml:"yandere"`
	Nhentai  SourceNhentaiConfig  `toml:"nhentai" mapstructure:"nhentai" json:"nhentai" yaml:"nhentai"`
	Remote   SourceRemoteConfig   `toml:"remote" mapstructure:"remote" json:"remote" yaml:"remote"`
}

type SourceRemoteConfig struct {
	Token    string                `toml:"token" mapstructure:"token" json:"token" yaml:"token"`
	Fetch    RemoteOperationConfig `toml:"fetch" mapstructure:"fetch" json:"fetch" yaml:"fetch"`
	Download RemoteOperationConfig `toml:"download" mapstructure:"download" json:"download" yaml:"download"`
}

type RemoteOperationConfig struct {
	Endpoints []string `toml:"endpoints" mapstructure:"endpoints" json:"endpoints" yaml:"endpoints"`
	Timeout   uint     `toml:"timeout" mapstructure:"timeout" json:"timeout" yaml:"timeout"`
}

const maxRemoteDurationSeconds = uint64((1<<63 - 1) / int64(time.Second))

func (c RemoteOperationConfig) Validate() error {
	for i, endpoint := range c.Endpoints {
		u, err := url.Parse(endpoint)
		if err != nil || u == nil || (u.Scheme != "http" && u.Scheme != "https") ||
			u.Hostname() == "" || u.Opaque != "" || u.User != nil ||
			u.RawQuery != "" || u.ForceQuery || strings.Contains(endpoint, "#") {
			return fmt.Errorf("endpoints[%d] must be an HTTP(S) API base URL without credentials, query or fragment", i)
		}
		if u.Port() != "" {
			port, err := strconv.ParseUint(u.Port(), 10, 16)
			if err != nil || port == 0 {
				return fmt.Errorf("endpoints[%d] has an invalid port", i)
			}
		}
	}
	if c.Timeout == 0 || uint64(c.Timeout) > maxRemoteDurationSeconds {
		return fmt.Errorf("timeout must be positive and fit a duration in seconds")
	}
	return nil
}

type SourcePixivConfig struct {
	ImgProxy string         `toml:"img_proxy" mapstructure:"img_proxy" json:"img_proxy" yaml:"img_proxy"`
	RssURLs  []string       `toml:"rss_urls" mapstructure:"rss_urls" json:"rss_urls" yaml:"rss_urls"`
	Cookies  []CookieConfig `toml:"cookies" mapstructure:"cookies" json:"cookies" yaml:"cookies"`
	Disable  bool           `toml:"disable" mapstructure:"disable" json:"disable" yaml:"disable"`
}

type SourceTwitterConfig struct {
	FxTwitterDomain string `toml:"fx_twitter_domain" mapstructure:"fx_twitter_domain" json:"fx_twitter_domain" yaml:"fx_twitter_domain"`
	Disable         bool   `toml:"disable" mapstructure:"disable" json:"disable" yaml:"disable"`
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
