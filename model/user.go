package model

import (
	"ManyACG/types"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type LikeModel struct {
	ID        primitive.ObjectID `bson:"_id,omitempty"`
	ArtworkID primitive.ObjectID `bson:"artwork_id"`
	UserID    primitive.ObjectID `bson:"user_id"`
}

type FavoriteModel struct {
	ID        primitive.ObjectID `bson:"_id,omitempty"`
	ArtworkID primitive.ObjectID `bson:"artwork_id"`
	UserID    primitive.ObjectID `bson:"user_id"`
}

type UserModel struct {
	ID         primitive.ObjectID `bson:"_id,omitempty"`
	Username   string             `bson:"username"`
	Password   string             `bson:"password"`
	Email      string             `bson:"email"`
	TelegramID int64              `bson:"telegram_id"`
	Blocked    bool               `bson:"blocked"`
	UpdatedAt  primitive.DateTime `bson:"updated_at"`
	DeletedAt  primitive.DateTime `bson:"deleted_at,omitempty"`

	// Settings
	Settings *UserSettings `bson:"settings" json:"settings"`
}

type UserSettings struct {
	Language string `bson:"language" json:"language"`
	Theme    string `bson:"theme" json:"theme"`
	R18      bool   `bson:"r18" json:"r18"`
}

type UnauthUserModel struct {
	ID         primitive.ObjectID `bson:"_id,omitempty"`
	Code       string             `bson:"code"` // 注册验证码
	Username   string             `bson:"username"`
	TelegramID int64              `bson:"telegram_id"`
	Email      string             `bson:"email"`
	AuthMethod types.AuthMethod   `bson:"auth_method"`
}
