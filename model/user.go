package model

import "go.mongodb.org/mongo-driver/bson/primitive"

type LikeModel struct {
	ID        primitive.ObjectID `bson:"_id,omitempty"`
	ArtworkID primitive.ObjectID `bson:"artwork_id"`
	UserID    primitive.ObjectID `bson:"user_id"`
	CreatedAt primitive.DateTime `bson:"created_at"`
}

type FavoriteModel struct {
	ID        primitive.ObjectID `bson:"_id,omitempty"`
	ArtworkID primitive.ObjectID `bson:"artwork_id"`
	UserID    primitive.ObjectID `bson:"user_id"`
	CreatedAt primitive.DateTime `bson:"created_at"`
	DeletedAt primitive.DateTime `bson:"deleted_at"`
}

type UserModel struct {
	ID         primitive.ObjectID `bson:"_id,omitempty"`
	Username   string             `bson:"username"`
	Password   string             `bson:"password"`
	Email      string             `bson:"email"`
	TelegramID int                `bson:"telegram_id"`
	Blocked    bool               `bson:"blocked"`
	CreatedAt  primitive.DateTime `bson:"created_at"`
	UpdatedAt  primitive.DateTime `bson:"updated_at"`
	DeletedAt  primitive.DateTime `bson:"deleted_at"`

	// Settings
	Settings *UserSettings `bson:"settings"`
}

type UserSettings struct {
	Language string `bson:"language"`
	Theme    string `bson:"theme"`
	R18      bool   `bson:"r18"`
}
