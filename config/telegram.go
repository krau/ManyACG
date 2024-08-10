package config

type telegramConfig struct {
	Token    string  `toml:"token" mapstructure:"token" json:"token" yaml:"token"`
	APIURL   string  `toml:"api_url" mapstructure:"api_url" json:"api_url" yaml:"api_url"`
	Admins   []int64 `toml:"admins" mapstructure:"admins" json:"admins" yaml:"admins"`
	Channel  bool    `toml:"channel" mapstructure:"channel" json:"channel" yaml:"channel"`
	ChatID   int64   `toml:"chat_id" mapstructure:"chat_id" json:"chat_id" yaml:"chat_id"`
	Username string  `toml:"username" mapstructure:"username" json:"username" yaml:"username"`
	Sleep    uint    `toml:"sleep" mapstructure:"sleep" json:"sleep" yaml:"sleep"`
	GroupID  int64   `toml:"group_id" mapstructure:"group_id" json:"group_id" yaml:"group_id"`
}
