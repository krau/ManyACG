package runtimecfg

type CacheConfig struct {
	// Type       string `toml:"type" mapstructure:"type" json:"type" yaml:"type"`
	DefaultTTL int `toml:"default_ttl" mapstructure:"default_ttl" json:"default_ttl" yaml:"default_ttl"` // seconds
	// // bigcache, ristretto(default), redis
	// BigCache struct {
	// 	Eviction int `toml:"eviction" mapstructure:"eviction" json:"eviction" yaml:"eviction"` // seconds
	// } `toml:"bigcache" mapstructure:"bigcache" json:"bigcache" yaml:"bigcache"`
	Ristretto struct {
		NumCounters int64 `toml:"num_counters" mapstructure:"num_counters" json:"num_counters" yaml:"num_counters"`
		MaxCost     int64 `toml:"max_cost" mapstructure:"max_cost" json:"max_cost" yaml:"max_cost"`
	} `toml:"ristretto" mapstructure:"ristretto" json:"ristretto" yaml:"ristretto"`
	Redis struct {
		InitAddress []string `toml:"init_address" mapstructure:"init_address" json:"init_address" yaml:"init_address"`
	} `toml:"redis" mapstructure:"redis" json:"redis" yaml:"redis"`
}
