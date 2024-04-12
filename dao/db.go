package dao

import (
	"ManyACG-Bot/config"
	"ManyACG-Bot/model/entity"
	"fmt"
	"os"

	. "ManyACG-Bot/logger"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var db *gorm.DB

func init() {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		config.Cfg.Database.User,
		config.Cfg.Database.Password,
		config.Cfg.Database.Host,
		config.Cfg.Database.Port,
		config.Cfg.Database.DBName,
	)
	var err error
	db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
		TranslateError: true,
		Logger:         logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		Logger.Fatalf("Error when connecting to database: %s", err)
		os.Exit(1)
	}
	db.AutoMigrate(&entity.Artwork{}, &entity.Picture{}, &entity.Tag{})
}
