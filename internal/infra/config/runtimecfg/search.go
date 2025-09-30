package runtimecfg

import "errors"

type SearchConfig struct {
	Enable      bool              `toml:"enable" mapstructure:"enable" json:"enable" yaml:"enable"`
	Engine      string            `toml:"engine" mapstructure:"engine" json:"engine" yaml:"engine"`
	MeiliSearch MeiliSearchConfig `toml:"meilisearch" mapstructure:"meilisearch" json:"meilisearch" yaml:"meilisearch"`
}

type MeiliSearchConfig struct {
	Host     string `toml:"host" mapstructure:"host" json:"host" yaml:"host"`
	Key      string `toml:"key" mapstructure:"key" json:"key" yaml:"key"`
	Index    string `toml:"index" mapstructure:"index" json:"index" yaml:"index"`
	Embedder string `toml:"embedder" mapstructure:"embedder" json:"embedder" yaml:"embedder"`
}

func (c MeiliSearchConfig) Valid() error {
	if c.Host == "" {
		return errors.New("meilisearch host is empty")
	}
	if c.Index == "" {
		return errors.New("meilisearch index is empty")
	}
	return nil
}
