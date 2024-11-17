package dao

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/krau/ManyACG/common"
	"github.com/krau/ManyACG/config"
	"github.com/krau/ManyACG/dao/collections"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

var Client *mongo.Client
var DB *mongo.Database

func InitDB(ctx context.Context) {
	common.Logger.Info("Initializing database...")
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
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri),
		options.Client().SetReadPreference(readpref.Nearest(readpref.WithMaxStaleness(time.Duration(config.Cfg.Database.MaxStaleness)*time.Second))))
	if err != nil {
		common.Logger.Fatal(err)
		os.Exit(1)
	}
	if err = client.Ping(ctx, nil); err != nil {
		common.Logger.Fatal(err)
		os.Exit(1)
	}
	Client = client
	DB = Client.Database(config.Cfg.Database.Database)
	if DB == nil {
		common.Logger.Fatal("Failed to get database")
		os.Exit(1)
	}
	createCollection(ctx)
	createIndex(ctx)

	common.Logger.Info("Database initialized")
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
	DB.CreateCollection(ctx, collections.EtcData)
	etcDataCollection = DB.Collection(collections.EtcData)
	DB.CreateCollection(ctx, collections.Users)
	userCollection = DB.Collection(collections.Users)
	DB.CreateCollection(ctx, collections.Likes)
	likeCollection = DB.Collection(collections.Likes)
	DB.CreateCollection(ctx, collections.Favorites)
	favoriteCollection = DB.Collection(collections.Favorites)
	DB.CreateCollection(ctx, collections.UnauthUser)
	unauthUserCollection = DB.Collection(collections.UnauthUser)
}

func createIndex(ctx context.Context) {

	// 作品
	artworkCollection.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "source_url", Value: 1}},
			Options: options.Index().SetName("source_url").SetUnique(true),
		},
	})

	// 缓存的作品
	cachedArtworkCollection.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "source_url", Value: 1}},
			Options: options.Index().SetName("source_url"),
		},
		{
			Keys:    bson.D{{Key: "created_at", Value: 1}},
			Options: options.Index().SetExpireAfterSeconds(86400 * 30).SetName("created_at"),
		},
	})

	// 标签
	tagCollection.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "name", Value: 1}},
			Options: options.Index().SetName("name").SetUnique(true),
		},
	})

	// 图片
	pictureCollection.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "original", Value: 1}},
			Options: options.Index().SetName("original").SetUnique(true),
		},
		{
			Keys:    bson.D{{Key: "telegram_info.message_id", Value: 1}},
			Options: options.Index().SetName("message_id"),
		},
		{
			Keys:    bson.D{{Key: "hash", Value: 1}},
			Options: options.Index().SetName("hash"),
		},
	})

	// 画师
	artistCollection.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "name", Value: 1}},
			Options: options.Index().SetName("name"),
		},
		{
			Keys:    bson.D{{Key: "uid", Value: 1}},
			Options: options.Index().SetName("uid"),
		},
		{
			Keys:    bson.D{{Key: "username", Value: 1}, {Key: "type", Value: 1}},
			Options: options.Index().SetName("username_type").SetUnique(true),
		},
	})

	// 管理员 (deprecated)
	adminCollection.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "user_id", Value: 1}},
			Options: options.Index().SetName("user_id").SetUnique(true),
		},
	})

	for _, admin := range config.Cfg.Telegram.Admins {
		_, err := CreateSuperAdminByUserID(ctx, admin, 0)
		if err != nil {
			common.Logger.Error(err)
		}
	}

	// 已删除的作品
	deletedCollection.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "source_url", Value: 1}},
			Options: options.Index().SetName("source_url").SetUnique(true),
		},
	})

	// 回调数据
	callbackDataCollection.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "created_at", Value: 1}},
			Options: options.Index().SetExpireAfterSeconds(86400 * 3).SetName("created_at"),
		},
	})

	// 用户
	userCollection.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "username", Value: 1}},
			Options: options.Index().SetName("username").SetUnique(true),
		},
		// TODO:
		// {
		// 	Keys:    bson.D{{Key: "email", Value: 1}},
		// 	Options: options.Index().SetName("email").SetUnique(true),
		// },
		// {
		// 	Keys:    bson.D{{Key: "telegram_id", Value: 1}},
		// 	Options: options.Index().SetName("telegram_id").SetUnique(true),
		// },
	})

	// 用户喜欢的作品 (24小时过期)
	likeCollection.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "artwork_id", Value: 1}},
			Options: options.Index().SetName("artwork_id"),
		},
		{
			Keys:    bson.D{{Key: "user_id", Value: 1}},
			Options: options.Index().SetName("user_id"),
		},
		{
			Keys:    bson.D{{Key: "created_at", Value: 1}},
			Options: options.Index().SetExpireAfterSeconds(86400).SetName("created_at"),
		},
	})

	// 用户收藏的作品
	favoriteCollection.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "artwork_id", Value: 1}},
			Options: options.Index().SetName("artwork_id"),
		},
		{
			Keys:    bson.D{{Key: "user_id", Value: 1}},
			Options: options.Index().SetName("user_id"),
		},
		{
			Keys:    bson.D{{Key: "artwork_id", Value: 1}, {Key: "user_id", Value: 1}},
			Options: options.Index().SetName("artwork_id_user_id").SetUnique(true),
		},
	})

	unauthUserCollection.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "created_at", Value: 1}},
			Options: options.Index().SetExpireAfterSeconds(600).SetName("created_at"),
		},
		{
			Keys:    bson.D{{Key: "username", Value: 1}},
			Options: options.Index().SetName("username").SetUnique(true),
		},
		// {
		// 	Keys:    bson.D{{Key: "telegram_id", Value: 1}},
		// 	Options: options.Index().SetName("telegram_id").SetUnique(true),
		// },
	})
}
