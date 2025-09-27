package config

type searchConfig struct {
	Enable      bool              `toml:"enable" mapstructure:"enable" json:"enable" yaml:"enable"`
	Engine      string            `toml:"engine" mapstructure:"engine" json:"engine" yaml:"engine"`
	MeiliSearch meiliSearchConfig `toml:"meilisearch" mapstructure:"meilisearch" json:"meilisearch" yaml:"meilisearch"`
}

type meiliSearchConfig struct {
	Host     string `toml:"host" mapstructure:"host" json:"host" yaml:"host"`
	Key      string `toml:"key" mapstructure:"key" json:"key" yaml:"key"`
	Index    string `toml:"index" mapstructure:"index" json:"index" yaml:"index"`
	Embedder string `toml:"embedder" mapstructure:"embedder" json:"embedder" yaml:"embedder"`
}
