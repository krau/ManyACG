package dao

import (
	"ManyACG-Bot/config"
	"ManyACG-Bot/dao/collections"
	. "ManyACG-Bot/logger"
	"context"
	"fmt"
	"os"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

var Client *mongo.Client
var DB *mongo.Database

func InitDB(ctx context.Context) {
	Logger.Info("Initializing database...")
	uri := config.Cfg.Database.URI
	if uri == "" {
		uri = fmt.Sprintf(
			"mongodb://%s:%s@%s:%d",
			config.Cfg.Database.User,
			config.Cfg.Database.Password,
			config.Cfg.Database.Host,
			config.Cfg.Database.Port,
		)
	}
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		Logger.Fatal(err)
		os.Exit(1)
	}
	if err = client.Ping(ctx, readpref.Primary()); err != nil {
		Logger.Fatal(err)
		os.Exit(1)
	}
	Client = client
	DB = Client.Database(config.Cfg.Database.Database)
	if DB == nil {
		Logger.Fatal("Failed to get database")
		os.Exit(1)
	}

	DB.CreateCollection(ctx, collections.Artworks)
	artworkCollection = DB.Collection(collections.Artworks)
	artworkCollection.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "source_url", Value: 1}},
			Options: options.Index().SetName("source_url"),
		},
	})

	DB.CreateCollection(ctx, collections.Tags)
	tagCollection = DB.Collection(collections.Tags)
	tagCollection.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "name", Value: 1}},
			Options: options.Index().SetName("name"),
		},
	})

	DB.CreateCollection(ctx, collections.Pictures)
	pictureCollection = DB.Collection(collections.Pictures)
	pictureCollection.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "original", Value: 1}},
			Options: options.Index().SetName("original"),
		},
	})

	DB.CreateCollection(ctx, collections.Artists)
	artistCollection = DB.Collection(collections.Artists)
	artistCollection.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "name", Value: 1}},
			Options: options.Index().SetName("name"),
		},
	})

	DB.CreateCollection(ctx, collections.Admins)
	adminCollection = DB.Collection(collections.Admins)
	adminCollection.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "user_id", Value: 1}},
			Options: options.Index().SetName("user_id").SetUnique(true),
		},
	})
	for _, admin := range config.Cfg.Telegram.Admins {
		if err := CreateAdminIfNotExist(ctx, admin); err != nil {
			Logger.Warnf("Failed to create admin %d: %s", admin, err)
		}
	}

	Logger.Info("Database initialized")
}
