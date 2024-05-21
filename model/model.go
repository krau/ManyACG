package model

import (
	"ManyACG/types"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ArtworkModel struct {
	ID          primitive.ObjectID `bson:"_id,omitempty"`
	Title       string             `bson:"title"`
	Description string             `bson:"description"`
	R18         bool               `bson:"r18"`
	CreatedAt   primitive.DateTime `bson:"created_at"`
	SourceType  types.SourceType   `bson:"source_type"`
	SourceURL   string             `bson:"source_url"`

	ArtistID primitive.ObjectID   `bson:"artist_id"`
	Tags     []primitive.ObjectID `bson:"tags"`
	Pictures []primitive.ObjectID `bson:"pictures"`
}

type ArtistModel struct {
	ID       primitive.ObjectID `bson:"_id,omitempty"`
	Name     string             `bson:"name"`
	Type     types.SourceType   `bson:"type"`
	UID      int                `bson:"uid"`
	Username string             `bson:"username"`
}

type TagModel struct {
	ID   primitive.ObjectID `bson:"_id,omitempty"`
	Name string             `bson:"name"`
	// Alias []string           `bson:"alias"`
}

type PictureModel struct {
	ID        primitive.ObjectID `bson:"_id,omitempty"`
	ArtworkID primitive.ObjectID `bson:"artwork_id"`
	Index     uint               `bson:"index"`
	Thumbnail string             `bson:"thumbnail"`
	Original  string             `bson:"original"`
	Width     uint               `bson:"width"`
	Height    uint               `bson:"height"`
	Hash      string             `bson:"hash"`
	BlurScore float64            `bson:"blur_score"`

	TelegramInfo *TelegramInfo `bson:"telegram_info"`
	StorageInfo  *StorageInfo  `bson:"storage_info"`
}

type TelegramInfo struct {
	PhotoFileID    string `bson:"photo_file_id"`
	DocumentFileID string `bson:"document_file_id"`
	MessageID      int    `bson:"message_id"`
	MediaGroupID   string `bson:"media_group_id"`
}

type StorageInfo struct {
	Type types.StorageType `bson:"type"`
	Path string            `bson:"path"`
}

type DeletedModel struct {
	ID        primitive.ObjectID `bson:"_id,omitempty"`
	ArtworkID primitive.ObjectID `bson:"artwork_id"`
	SourceURL string             `bson:"source_url"`
	DeletedAt primitive.DateTime `bson:"deleted_at"`
}

type CallbackDataModel struct {
	ID        primitive.ObjectID `bson:"_id,omitempty"`
	Data      string             `bson:"data"`
	CreatedAt primitive.DateTime `bson:"created_at"`
}

type CachedArtworksModel struct {
	ID        primitive.ObjectID  `bson:"_id,omitempty"`
	SourceURL string              `bson:"source_url"`
	CreatedAt primitive.DateTime  `bson:"created_at"`
	Artwork   *types.Artwork      `bson:"artwork"`
	Status    types.ArtworkStatus `bson:"status"`
}
