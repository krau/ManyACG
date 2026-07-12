package runtimecfg

type ImsearchConfig struct {
	Enable           bool    `toml:"enable" mapstructure:"enable" json:"enable" yaml:"enable"`
	DataDir          string  `toml:"data_dir" mapstructure:"data_dir" json:"data_dir" yaml:"data_dir"`
	Distance         int     `toml:"distance" mapstructure:"distance" json:"distance" yaml:"distance"`
	Count            int     `toml:"count" mapstructure:"count" json:"count" yaml:"count"`
	K                int     `toml:"k" mapstructure:"k" json:"k" yaml:"k"`
	NProbe           int     `toml:"nprobe" mapstructure:"nprobe" json:"nprobe" yaml:"nprobe"`
	NFeatures        int     `toml:"nfeatures" mapstructure:"nfeatures" json:"nfeatures" yaml:"nfeatures"`
	MaxHeight        int     `toml:"max_height" mapstructure:"max_height" json:"max_height" yaml:"max_height"`
	MaxWidth         int     `toml:"max_width" mapstructure:"max_width" json:"max_width" yaml:"max_width"`
	AutoBuild        bool    `toml:"auto_build" mapstructure:"auto_build" json:"auto_build" yaml:"auto_build"`
	BuildDebounceSec int     `toml:"build_debounce_sec" mapstructure:"build_debounce_sec" json:"build_debounce_sec" yaml:"build_debounce_sec"`
	MinMatches       int     `toml:"min_matches" mapstructure:"min_matches" json:"min_matches" yaml:"min_matches"`
	MinScore         float32 `toml:"min_score" mapstructure:"min_score" json:"min_score" yaml:"min_score"`
}
