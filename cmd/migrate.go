package cmd

import (
	"ManyACG/service"
	"context"

	"github.com/spf13/cobra"
)

var migrateCmd = &cobra.Command{
	Use:     "migrate",
	Aliases: []string{"mig"},
	Short:   "Migrate database, please backup before running this command",
	Run: func(cmd *cobra.Command, args []string) {
		Migrate()
	},
}

func init() {
	rootCmd.AddCommand(migrateCmd)
}

func Migrate() {
	service.Migrate(context.Background())
}
