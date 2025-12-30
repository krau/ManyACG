package runtimecfg

type databaseConfig struct {
	Type  string      `toml:"type" mapstructure:"type" json:"type" yaml:"type"`
	DSN   string      `toml:"dsn" mapstructure:"dsn" json:"dsn" yaml:"dsn"`
	Pgsql pgsqlConfig `toml:"pgsql" mapstructure:"pgsql" json:"pgsql" yaml:"pgsql"`
}

type pgsqlConfig struct {
	PGroonga bool `toml:"pgroonga" mapstructure:"pgroonga" json:"pgroonga" yaml:"pgroonga"`
	// PGroonga index isn't crash safe. You need to run REINDEX when your PGroonga index is broken by crash.
	ReindexPGroonga bool `toml:"reindex_pgroonga" mapstructure:"reindex_pgroonga" json:"reindex_pgroonga" yaml:"reindex_pgroonga"`
}
