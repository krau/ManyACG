package cmd

import (
	"context"
	"os"
	"time"

	"github.com/krau/ManyACG/common"
	"github.com/krau/ManyACG/config"
	"github.com/krau/ManyACG/dao"
	"github.com/krau/ManyACG/service"

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
		common.Init()
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		common.Logger.Info("Start migrating")
		dao.InitDB(ctx)
		defer func() {
			if err := dao.Client.Disconnect(ctx); err != nil {
				common.Logger.Fatal(err)
				os.Exit(1)
			}
		}()
		service.TidyArtist(context.Background())
	},
}

// TODO: remove this command after v0.66.0 is released
var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Migrate database From v0.64.1 To v0.65.0",
	Run: func(cmd *cobra.Command, args []string) {
		config.InitConfig()
		common.Init()
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		common.Logger.Info("Start migrating")
		dao.InitDB(ctx)
		defer func() {
			if err := dao.Client.Disconnect(ctx); err != nil {
				common.Logger.Fatal(err)
				os.Exit(1)
			}
		}()
		if err := dao.AddAliasToAllTags(ctx); err != nil {
			common.Logger.Fatal(err)
			os.Exit(1)
		}
		common.Logger.Info("Migrate completed")
	},
}

func init() {
	rootCmd.AddCommand(dbCmd)
	dbCmd.AddCommand(artistCmd)
	dbCmd.AddCommand(migrateCmd)
	artistCmd.AddCommand(tidyArtistCmd)
}
