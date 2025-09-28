package runtimecfg

type telegramConfig struct {
	Token    string         `toml:"token" mapstructure:"token" json:"token" yaml:"token"`
	APIURL   string         `toml:"api_url" mapstructure:"api_url" json:"api_url" yaml:"api_url"`
	Admins   []int64        `toml:"admins" mapstructure:"admins" json:"admins" yaml:"admins"`
	Channel  bool           `toml:"channel" mapstructure:"channel" json:"channel" yaml:"channel"`
	ChatID   int64          `toml:"chat_id" mapstructure:"chat_id" json:"chat_id" yaml:"chat_id"`
	Username string         `toml:"username" mapstructure:"username" json:"username" yaml:"username"`
	Sleep    uint           `toml:"sleep" mapstructure:"sleep" json:"sleep" yaml:"sleep"`
	GroupID  int64          `toml:"group_id" mapstructure:"group_id" json:"group_id" yaml:"group_id"`
	Retry    botRetryConfig `toml:"retry" mapstructure:"retry" json:"retry" yaml:"retry"`
}

type botRetryConfig struct {
	MaxAttempts  int     `toml:"max_attempts" mapstructure:"max_attempts" json:"max_attempts" yaml:"max_attempts"`
	ExponentBase float64 `toml:"exponent_base" mapstructure:"exponent_base" json:"exponent_base" yaml:"exponent_base"`
	StartDelay   int64   `toml:"start_delay" mapstructure:"start_delay" json:"start_delay" yaml:"start_delay"`
	MaxDelay     int64   `toml:"max_delay" mapstructure:"max_delay" json:"max_delay" yaml:"max_delay"`
}
