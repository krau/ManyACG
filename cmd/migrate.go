package cmd

import (
	"github.com/krau/ManyACG/internal/infra/config/runtimecfg"
	"github.com/krau/ManyACG/internal/infra/database"
	"github.com/spf13/cobra"
)

var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Initialize or update the database schema",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		return database.Migrate(cmd.Context(), runtimecfg.Get())
	},
}

func init() {
	rootCmd.AddCommand(migrateCmd)
}
