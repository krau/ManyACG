package runtimecfg

type databaseConfig struct {
	URI          string `toml:"uri" mapstructure:"uri" json:"uri" yaml:"uri"`
	Host         string `toml:"host" mapstructure:"host" json:"host" yaml:"host"`
	Port         int    `toml:"port" mapstructure:"port" json:"port" yaml:"port"`
	User         string `toml:"user" mapstructure:"user" json:"user" yaml:"user"`
	Password     string `toml:"password" mapstructure:"password" json:"password" yaml:"password"`
	Database     string `toml:"database" mapstructure:"database" json:"database" yaml:"database"`
	MaxStaleness int    `toml:"max_staleness" mapstructure:"max_staleness" json:"max_staleness" yaml:"max_staleness"`
	Type         string `toml:"type" mapstructure:"type" json:"type" yaml:"type"`
	DSN          string `toml:"dsn" mapstructure:"dsn" json:"dsn" yaml:"dsn"`
}
