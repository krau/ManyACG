package cmd

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/krau/ManyACG/cmd/migrate"
	"github.com/krau/ManyACG/dao"
	"github.com/krau/ManyACG/internal/infra/config"
	"github.com/ncruces/go-sqlite3/gormlite"
	"github.com/spf13/cobra"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
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

var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Migrate database from mongodb to sql(pgsql, mysql and sqlite supported)",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
		defer cancel()
		config.Get()
		if config.Get().Mirate.DSN == "" || config.Get().Mirate.Target == "" {
			fmt.Println("请在配置文件中设置 migrate.dsn 和 migrate.target")
			os.Exit(1)
		}
		log.Println("Start migrating")
		defer log.Println("Migrate completed, the log is in migrate_sql.log")
		dao.InitDB(ctx)
		defer func() {
			if err := dao.Client.Disconnect(ctx); err != nil {
				fmt.Printf("Error when disconnecting mongodb: %s\n", err)
				os.Exit(1)
			}
		}()

		dbLogFile, err := os.Create("migrate_sql.log")
		if err != nil {
			return fmt.Errorf("failed to open migrate_sql.log: %w", err)
		}
		defer dbLogFile.Close()
		newLogger := logger.New(log.New(dbLogFile, "\r\n", log.LstdFlags), logger.Config{
			SlowThreshold: 10 * time.Second,
			LogLevel:      logger.Warn,
			Colorful:      false,
		})
		var db *gorm.DB
		switch config.Get().Mirate.Target {
		case "sqlite":
			log.Println("Using sqlite")
			gormDB, err := gorm.Open(gormlite.Open(config.Get().Mirate.DSN), &gorm.Config{
				Logger: newLogger,
			})
			if err != nil {
				return fmt.Errorf("failed to connect to sqlite: %w", err)
			}
			db = gormDB
		case "mysql":
			log.Println("Using mysql")
			gormDB, err := gorm.Open(mysql.Open(config.Get().Mirate.DSN), &gorm.Config{
				Logger: newLogger,
			})
			if err != nil {
				return fmt.Errorf("failed to connect to mysql: %w", err)
			}
			db = gormDB
		case "pgsql":
			log.Println("Using pgsql")
			gormDB, err := gorm.Open(postgres.Open(config.Get().Mirate.DSN), &gorm.Config{
				Logger: newLogger,
			})
			if err != nil {
				return fmt.Errorf("failed to connect to pgsql: %w", err)
			}
			db = gormDB
		default:
			return fmt.Errorf("不支持的目标数据库: %s", config.Get().Mirate.Target)
		}
		db = db.WithContext(ctx)
		return migrate.Run(ctx, &migrate.Option{
			MongoClient: dao.Client,
			GormDB:      db,
			Cfg:         config.Get(),
		})
	},
}

func init() {
	rootCmd.AddCommand(dbCmd)
	dbCmd.AddCommand(migrateCmd)
}
