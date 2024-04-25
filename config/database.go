package config

type databaseConfig struct {
	URI      string `toml:"uri" mapstructure:"uri" json:"uri" yaml:"uri"`
	Host     string `toml:"host" mapstructure:"host" json:"host" yaml:"host"`
	Port     int    `toml:"port" mapstructure:"port" json:"port" yaml:"port"`
	User     string `toml:"user" mapstructure:"user" json:"user" yaml:"user"`
	Password string `toml:"password" mapstructure:"password" json:"password" yaml:"password"`
	Database string `toml:"database"`
}
