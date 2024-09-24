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

var dbCmd = &cobra.Command{
	Use:   "db",
	Short: "Database batch operations",
}

var artistCmd = &cobra.Command{
	Use:   "artist",
	Short: "Operations on artists",
}

var tidyArtistCmd = &cobra.Command{
	Use:   "tidy",
	Short: "Tidy artists",
	Long:  "清理没有任何 artwork 的 artist, 通过 source type 和 username 合并相同的 artist.",
	Run: func(cmd *cobra.Command, args []string) {
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
		service.TidyArtist(context.Background())
	},
}

func init() {
	rootCmd.AddCommand(dbCmd)
	dbCmd.AddCommand(artistCmd)
	artistCmd.AddCommand(tidyArtistCmd)
}