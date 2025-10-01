package runtimecfg

type databaseConfig struct {
	Type         string `toml:"type" mapstructure:"type" json:"type" yaml:"type"`
	DSN          string `toml:"dsn" mapstructure:"dsn" json:"dsn" yaml:"dsn"`
}
