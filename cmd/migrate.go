package cmd

import (
	"ManyACG/config"
	"ManyACG/dao"
	"ManyACG/logger"
	"ManyACG/service"
	"context"
	"os"
	"time"

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
	config.InitConfig()
	logger.InitLogger()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	logger.Logger.Info("Start migrating")
	dao.InitDB(ctx)
	defer func() {
		if err := dao.Client.Disconnect(ctx); err != nil {
			logger.Logger.Fatal(err)
			os.Exit(1)
		}
	}()
	service.Migrate(context.Background())
}
