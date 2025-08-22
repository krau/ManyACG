package types

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ArtworkModel struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	Title       string             `bson:"title" json:"title"`
	Description string             `bson:"description" json:"description"`
	R18         bool               `bson:"r18" json:"r18"`
	CreatedAt   primitive.DateTime `bson:"created_at" json:"created_at"`
	SourceType  SourceType         `bson:"source_type" json:"source_type"`
	SourceURL   string             `bson:"source_url" json:"source_url"`
	LikeCount   uint               `bson:"like_count" json:"like_count"`

	ArtistID primitive.ObjectID   `bson:"artist_id" json:"artist_id"`
	Tags     []primitive.ObjectID `bson:"tags" json:"tags"`
	Pictures []primitive.ObjectID `bson:"pictures" json:"pictures"`
}

type ArtistModel struct {
	ID       primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	Name     string             `bson:"name" json:"name"`
	Type     SourceType         `bson:"type" json:"type"`
	UID      string             `bson:"uid" json:"uid"`
	Username string             `bson:"username" json:"username"`
}

type TagModel struct {
	// example:
	/*
		{
			id: objectId,
			name: "碧蓝档案",
			alias: ["碧蓝档案","蔚蓝档案","blue archive"]
		}
	*/
	ID    primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	Name  string             `bson:"name" json:"name"`
	Alias []string           `bson:"alias" json:"alias"`
}

type PictureModel struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	ArtworkID primitive.ObjectID `bson:"artwork_id" json:"artwork_id"`
	Index     uint               `bson:"index" json:"index"`
	Thumbnail string             `bson:"thumbnail" json:"thumbnail"`
	Original  string             `bson:"original" json:"original"`
	Width     uint               `bson:"width" json:"width"`
	Height    uint               `bson:"height" json:"height"`
	Hash      string             `bson:"hash" json:"hash"`             // phash
	BlurScore float64            `bson:"blur_score" json:"blur_score"` // deprecated
	ThumbHash string             `bson:"thumb_hash" json:"thumb_hash"` // https://github.com/evanw/thumbhash

	TelegramInfo *TelegramInfo `bson:"telegram_info" json:"telegram_info"`
	StorageInfo  *StorageInfo  `bson:"storage_info" json:"storage_info"`
}

type DeletedModel struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	ArtworkID primitive.ObjectID `bson:"artwork_id" json:"artwork_id"`
	SourceURL string             `bson:"source_url" json:"source_url"`
	DeletedAt primitive.DateTime `bson:"deleted_at" json:"deleted_at"`
}

type CallbackDataModel struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	Data      string             `bson:"data" json:"data"`
	CreatedAt primitive.DateTime `bson:"created_at" json:"created_at"`
}

type CachedArtworksModel struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	SourceURL string             `bson:"source_url" json:"source_url"`
	CreatedAt primitive.DateTime `bson:"created_at" json:"created_at"`
	Artwork   *Artwork           `bson:"artwork" json:"artwork"`
	Status    ArtworkStatus      `bson:"status" json:"status"`
}

func (picture *PictureModel) ToPicture() *Picture {
	return &Picture{
		ID:        picture.ID.Hex(),
		ArtworkID: picture.ArtworkID.Hex(),
		Index:     picture.Index,
		Thumbnail: picture.Thumbnail,
		Original:  picture.Original,
		Width:     picture.Width,
		Height:    picture.Height,
		Hash:      picture.Hash,
		// BlurScore:    picture.BlurScore,
		TelegramInfo: picture.TelegramInfo,
		StorageInfo:  picture.StorageInfo,
	}
}

func (artist *ArtistModel) ToArtist() *Artist {
	return &Artist{
		ID:       artist.ID.Hex(),
		Name:     artist.Name,
		Type:     artist.Type,
		UID:      artist.UID,
		Username: artist.Username,
	}
}
