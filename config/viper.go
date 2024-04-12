package config

import "github.com/spf13/viper"

type Config struct {
	Log      logConfig      `toml:"log" mapstructure:"log" json:"log" yaml:"log"`
	Source   sourceConfigs  `toml:"source" mapstructure:"source" json:"source" yaml:"source"`
	Telegram TelegramConfig `toml:"telegram" mapstructure:"telegram" json:"telegram" yaml:"telegram"`
}

type logConfig struct {
	Level     string `toml:"level" mapstructure:"level" json:"level" yaml:"level"`
	FilePath  string `toml:"file_path" mapstructure:"file_path" json:"file_path" yaml:"file_path"`
	BackupNum uint   `toml:"backup_num" mapstructure:"backup_num" json:"backup_num" yaml:"backup_num"`
}

type sourceConfigs struct {
	Pixiv SourcePixivConfig `toml:"pixiv" mapstructure:"pixiv" json:"pixiv" yaml:"pixiv"`
}

type SourcePixivConfig struct {
	Enable   bool     `toml:"enable" mapstructure:"enable" json:"enable" yaml:"enable"`
	URLs     []string `toml:"urls" mapstructure:"urls" json:"urls" yaml:"urls"`
	Intervel uint     `toml:"intervel" mapstructure:"intervel" json:"intervel" yaml:"intervel"`
}

type TelegramConfig struct {
	Token    string `toml:"token" mapstructure:"token" json:"token" yaml:"token"`
	ChatID   int64  `toml:"chat_id" mapstructure:"chat_id" json:"chat_id" yaml:"chat_id"`
	Username string `toml:"username" mapstructure:"username" json:"username" yaml:"username"`
}

var Cfg *Config

func init() {
	viper.SetConfigName("config")
	viper.AddConfigPath(".")
	viper.SetConfigType("toml")

	viper.SetDefault("log.level", "TRACE")
	viper.SetDefault("log.file_path", "logs/ManyACG-Bot.log")
	viper.SetDefault("log.backup_num", 7)

	if err := viper.ReadInConfig(); err != nil {
		panic(err)
	}
	Cfg = &Config{}
	if err := viper.Unmarshal(Cfg); err != nil {
		panic(err)
	}
}
