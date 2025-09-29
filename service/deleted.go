package service

import (
	"context"

	"github.com/krau/ManyACG/internal/infra/database"
	"github.com/krau/ManyACG/types"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func DeleteDeletedByURL(ctx context.Context, sourceURL string) error {
	// _, err := dao.DeleteDeletedByURL(ctx, sourceURL)
	// if err != nil {
	// 	return err
	// }
	// return nil
	return database.Default().CancelDeletedByURL(ctx, sourceURL)
}

func CheckDeletedByURL(ctx context.Context, sourceURL string) bool {
	// return dao.CheckDeletedByURL(ctx, sourceURL)
	return database.Default().CheckDeletedByURL(ctx, sourceURL)
}

func GetDeletedByURL(ctx context.Context, sourceURL string) (*types.DeletedModel, error) {
	// return dao.GetDeletedByURL(ctx, sourceURL)
	deleted, err := database.Default().GetDeletedByURL(ctx, sourceURL)
	if err != nil {
		return nil, err
	}
	return &types.DeletedModel{
		ID:        primitive.ObjectID(deleted.ID.ToObjectID()),
		ArtworkID: primitive.ObjectID(deleted.ArtworkID.ToObjectID()),
		SourceURL: deleted.SourceURL,
		DeletedAt: primitive.NewDateTimeFromTime(deleted.DeletedAt),
	}, nil
}
