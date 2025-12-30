package runtimecfg

type TelegramConfig struct {
	BotToken        string                      `toml:"bot_token" mapstructure:"bot_token" json:"bot_token" yaml:"bot_token"`
	APIURL          string                      `toml:"api_url" mapstructure:"api_url" json:"api_url" yaml:"api_url"`
	Username        string                      `toml:"username" mapstructure:"username" json:"username" yaml:"username"`
	CaptionTemplate string                      `toml:"caption_template" mapstructure:"caption_template" json:"caption_template" yaml:"caption_template"`
	Admins          []int64                     `toml:"admins" mapstructure:"admins" json:"admins" yaml:"admins"`
	ExtraTarget     []TelegramExtraTargetConfig `toml:"extra_target" mapstructure:"extra_target" json:"extra_target" yaml:"extra_target"`
	Retry           BotRetryConfig              `toml:"retry" mapstructure:"retry" json:"retry" yaml:"retry"`
	// Channel  bool    `toml:"channel" mapstructure:"channel" json:"channel" yaml:"channel"`
	ChatID  int64 `toml:"chat_id" mapstructure:"chat_id" json:"chat_id" yaml:"chat_id"`
	GroupID int64 `toml:"group_id" mapstructure:"group_id" json:"group_id" yaml:"group_id"`
	Disable bool  `toml:"disable" mapstructure:"disable" json:"disable" yaml:"disable"`
}

// 额外的发送的目标聊天, 将会在作品信息的操作键盘上显示发送到这些频道(但不落库)
type TelegramExtraTargetConfig struct {
	Title  string `toml:"title" mapstructure:"title" json:"title" yaml:"title"` // 在按钮上显示的标题
	ChatID int64  `toml:"chat_id" mapstructure:"chat_id" json:"chat_id" yaml:"chat_id"`
}

type BotRetryConfig struct {
	MaxAttempts  int     `toml:"max_attempts" mapstructure:"max_attempts" json:"max_attempts" yaml:"max_attempts"`
	ExponentBase float64 `toml:"exponent_base" mapstructure:"exponent_base" json:"exponent_base" yaml:"exponent_base"`
	StartDelay   int64   `toml:"start_delay" mapstructure:"start_delay" json:"start_delay" yaml:"start_delay"`
	MaxDelay     int64   `toml:"max_delay" mapstructure:"max_delay" json:"max_delay" yaml:"max_delay"`
}
