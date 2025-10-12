package runtimecfg

type LogConfig struct {
	FilePath  string `toml:"file_path" mapstructure:"file_path" json:"file_path" yaml:"file_path"`
	BackupNum uint   `toml:"backup_num" mapstructure:"backup_num" json:"backup_num" yaml:"backup_num"`
}
