package dao

import "go.mongodb.org/mongo-driver/mongo"

var (
	userCollection     *mongo.Collection
	likeCollection     *mongo.Collection
	favoriteCollection *mongo.Collection
)
