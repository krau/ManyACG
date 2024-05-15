package dao

import (
	"ManyACG/config"
	"ManyACG/dao/collections"
	. "ManyACG/logger"
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
	createCollection(ctx)
	createIndex(ctx)

	Logger.Info("Database initialized")
}

func createCollection(ctx context.Context) {
	DB.CreateCollection(ctx, collections.Artworks)
	artworkCollection = DB.Collection(collections.Artworks)
	DB.CreateCollection(ctx, collections.Tags)
	tagCollection = DB.Collection(collections.Tags)
	DB.CreateCollection(ctx, collections.Pictures)
	pictureCollection = DB.Collection(collections.Pictures)
	DB.CreateCollection(ctx, collections.Artists)
	artistCollection = DB.Collection(collections.Artists)
	DB.CreateCollection(ctx, collections.Admins)
	adminCollection = DB.Collection(collections.Admins)
	DB.CreateCollection(ctx, collections.Deleted)
	deletedCollection = DB.Collection(collections.Deleted)
	DB.CreateCollection(ctx, collections.CallbackData)
	callbackDataCollection = DB.Collection(collections.CallbackData)
	DB.CreateCollection(ctx, collections.CachedArtworks)
	cachedArtworkCollection = DB.Collection(collections.CachedArtworks)
}

func createIndex(ctx context.Context) {
	artworkCollection.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "source_url", Value: 1}},
			Options: options.Index().SetName("source_url").SetUnique(true),
		},
	})

	tagCollection.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "name", Value: 1}},
			Options: options.Index().SetName("name").SetUnique(true),
		},
	})

	pictureCollection.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "original", Value: 1}},
			Options: options.Index().SetName("original").SetUnique(true),
		},
		{
			Keys:    bson.D{{Key: "telegram_info.message_id", Value: 1}},
			Options: options.Index().SetName("message_id").SetUnique(true),
		},
		{
			Keys:    bson.D{{Key: "hash", Value: 1}},
			Options: options.Index().SetName("hash"),
		},
	})

	artistCollection.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "name", Value: 1}},
			Options: options.Index().SetName("name"),
		},
	})

	adminCollection.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "user_id", Value: 1}},
			Options: options.Index().SetName("user_id").SetUnique(true),
		},
	})
	for _, admin := range config.Cfg.Telegram.Admins {
		_, err := CreateSuperAdminByUserID(ctx, admin, 0)
		if err != nil {
			Logger.Error(err)
		}
	}

	deletedCollection.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "source_url", Value: 1}},
			Options: options.Index().SetName("source_url").SetUnique(true),
		},
	})

	callbackDataCollection.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "created_at", Value: 1}},
			Options: options.Index().SetExpireAfterSeconds(86400).SetName("created_at"),
		},
	})

	cachedArtworkCollection.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "source_url", Value: 1}},
			Options: options.Index().SetName("source_url").SetUnique(true),
		},
		{
			Keys:    bson.D{{Key: "created_at", Value: 1}},
			Options: options.Index().SetExpireAfterSeconds(86400).SetName("created_at"),
		},
	})
}
