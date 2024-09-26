package config

type authConfig struct {
	Resend resendConfig `toml:"resend" mapstructure:"resend" json:"resend" yaml:"resend"`
}

type resendConfig struct {
	APIKey  string `toml:"api_key" mapstructure:"api_key" json:"api_key" yaml:"api_key"`
	From    string `toml:"from" mapstructure:"from" json:"from" yaml:"from"`
	Subject string `toml:"subject" mapstructure:"subject" json:"subject" yaml:"subject"`
}
