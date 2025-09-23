package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"time"

	"github.com/krau/ManyACG/cmd/migrate"
	"github.com/krau/ManyACG/common"
	"github.com/krau/ManyACG/config"
	"github.com/krau/ManyACG/dao"

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

var tagCmd = &cobra.Command{
	Use:   "tag",
	Short: "Operations on tags",
}

var tidyArtistCmd = &cobra.Command{
	Use:   "tidy",
	Short: "Tidy artists",
	Long:  "清理没有任何 artwork 的 artist, 通过 source type 和 username 合并相同的 artist.",
	Run: func(cmd *cobra.Command, args []string) {
		config.InitConfig()
		common.Init()
		ctx, cancel := context.WithTimeout(context.TODO(), 30*time.Second)
		defer cancel()
		common.Logger.Info("Start migrating")
		dao.InitDB(ctx)
		defer func() {
			if err := dao.Client.Disconnect(ctx); err != nil {
				common.Logger.Fatal(err)
				os.Exit(1)
			}
		}()
		if err := dao.TidyArtist(context.TODO()); err != nil {
			common.Logger.Fatal(err)
			os.Exit(1)
		}
		common.Logger.Info("Tidy artist completed")
	},
}

var tidyTagCmd = &cobra.Command{
	Use:   "tidy",
	Short: "Tidy tags",
	Long:  "清理没有任何 artwork 的 tag",
	Run: func(cmd *cobra.Command, args []string) {
		config.InitConfig()
		common.Init()
		ctx, cancel := context.WithTimeout(context.TODO(), 30*time.Second)
		defer cancel()
		common.Logger.Info("Start migrating")
		dao.InitDB(ctx)
		defer func() {
			if err := dao.Client.Disconnect(ctx); err != nil {
				common.Logger.Fatal(err)
				os.Exit(1)
			}
		}()
		if err := dao.TidyTag(context.TODO()); err != nil {
			common.Logger.Fatal(err)
			os.Exit(1)
		}
		common.Logger.Info("Tidy tag completed")
	},
}

var cleanTagCmd = &cobra.Command{
	Use:   "clean",
	Short: "Clean tags",
	Long:  "按给定的正则表达式清理 tag",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			fmt.Println("请提供表达式")
			os.Exit(1)
		}
		config.InitConfig()
		common.Init()
		ctx, cancel := context.WithTimeout(context.TODO(), 30*time.Second)
		defer cancel()
		common.Logger.Info("Start migrating")
		dao.InitDB(ctx)
		defer func() {
			if err := dao.Client.Disconnect(ctx); err != nil {
				common.Logger.Fatal(err)
				os.Exit(1)
			}
		}()
		if err := dao.CleanTag(ctx, args[0]); err != nil {
			common.Logger.Fatal(err)
			os.Exit(1)
		}
		common.Logger.Info("Clean tag completed")
	},
}

var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Migrate database from mongodb to sql(pgsql, mysql and sqlite supported)",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
		defer cancel()
		config.InitConfig()
		if config.Cfg.Mirate.DSN == "" || config.Cfg.Mirate.Target == "" {
			fmt.Println("请在配置文件中设置 migrate.dsn 和 migrate.target")
			os.Exit(1)
		}
		return migrate.Run(ctx, nil)
	},
}

func init() {
	rootCmd.AddCommand(dbCmd)
	dbCmd.AddCommand(artistCmd)
	artistCmd.AddCommand(tidyArtistCmd)
	dbCmd.AddCommand(tagCmd)
	tagCmd.AddCommand(tidyTagCmd)
	tagCmd.AddCommand(cleanTagCmd)
	dbCmd.AddCommand(migrateCmd)
}
