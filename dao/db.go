package dao

import (
	"ManyACG-Bot/config"
	"ManyACG-Bot/dao/collections"
	. "ManyACG-Bot/logger"
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

var Client *mongo.Client
var DB *mongo.Database

func InitDB(ctx context.Context) {
	uri := fmt.Sprintf(
		"mongodb://%s:%s@%s:%d",
		config.Cfg.Database.User,
		config.Cfg.Database.Password,
		config.Cfg.Database.Host,
		config.Cfg.Database.Port,
	)
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		Logger.Panic(err)
	}
	if err = client.Ping(ctx, readpref.Primary()); err != nil {
		Logger.Panic(err)
	}
	Client = client
	DB = Client.Database(config.Cfg.Database.Database)
	DB.CreateCollection(ctx, collections.Artworks)
}
