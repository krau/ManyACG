package config

type telegramConfig struct {
	Token    string  `toml:"token" mapstructure:"token" json:"token" yaml:"token"`
	ChatID   int64   `toml:"chat_id" mapstructure:"chat_id" json:"chat_id" yaml:"chat_id"`
	Username string  `toml:"username" mapstructure:"username" json:"username" yaml:"username"`
	Sleep    uint    `toml:"sleep" mapstructure:"sleep" json:"sleep" yaml:"sleep"`
	Admins   []int64 `toml:"admins" mapstructure:"admins" json:"admins" yaml:"admins"`
}
