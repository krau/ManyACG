package dao

import (
	"ManyACG/model"
	"ManyACG/types"
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func QueryArtworksByTexts(ctx context.Context, texts [][]string, r18 types.R18Type, limit int) ([]*model.ArtworkModel, error) {
	if len(texts) == 0 {
		return GetArtworksByR18(ctx, r18, limit)
	}

	query := &artworkTextsQuery{
		Texts: texts,
		R18:   r18,
		Limit: limit,
	}
	pipeline := query.buildPipeline(ctx)
	cursor, err := artworkCollection.Aggregate(ctx, pipeline)

	if err != nil {
		return nil, err
	}
	var artworks []*model.ArtworkModel

	err = cursor.All(ctx, &artworks)
	if err != nil {
		return nil, err
	}
	if len(artworks) == 0 {
		return nil, mongo.ErrNoDocuments
	}
	return artworks, nil
}

func QueryArtworksByTextsPage(ctx context.Context, texts [][]string, r18 types.R18Type, page, pageSize int64) ([]*model.ArtworkModel, error) {
	if len(texts) == 0 {
		return GetArtworksByR18(ctx, r18, int(pageSize))
	}
	query := &artworkTextsQuery{
		Texts:    texts,
		R18:      r18,
		Page:     page,
		PageSize: pageSize,
	}
	pipeline := query.buildPipeline(ctx)
	cursor, err := artworkCollection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	var artworks []*model.ArtworkModel
	err = cursor.All(ctx, &artworks)
	if err != nil {
		return nil, err
	}
	if len(artworks) == 0 {
		return nil, mongo.ErrNoDocuments
	}
	return artworks, nil
}

type artworkTextsQuery struct {
	Texts    [][]string
	R18      types.R18Type
	Limit    int
	Page     int64
	PageSize int64
}

func (q *artworkTextsQuery) buildPipeline(ctx context.Context) []bson.M {
	var andConditions []bson.M
	for _, textGroup := range q.Texts {
		var orConditions []bson.M
		tagIDs := getTagIDs(ctx, textGroup)
		if len(tagIDs) > 0 {
			orConditions = append(orConditions, bson.M{"tags": bson.M{"$in": tagIDs}})
		}

		artistIDs := getArtistIDs(ctx, textGroup)
		if len(artistIDs) > 0 {
			orConditions = append(orConditions, bson.M{"artist_id": bson.M{"$in": artistIDs}})
		}

		for _, text := range textGroup {
			orConditions = append(orConditions, bson.M{"title": bson.M{"$regex": text, "$options": "i"}})
			orConditions = append(orConditions, bson.M{"description": bson.M{"$regex": text, "$options": "i"}})
		}

		if len(orConditions) > 0 {
			andConditions = append(andConditions, bson.M{"$or": orConditions})
		}
	}

	match := bson.M{"$and": andConditions}
	if q.R18 != types.R18TypeAll {
		match["r18"] = q.R18 == types.R18TypeOnly
	}

	// 如果没有分页返回 limit 条数据
	if q.Page <= 0 && q.PageSize <= 0 {
		return []bson.M{
			{"$match": match},
			{"$sample": bson.M{"size": q.Limit}},
		}
	}

	return []bson.M{
		{"$match": match},
		{"$skip": (q.Page - 1) * q.PageSize},
		{"$limit": q.PageSize},
	}
}

func getTagIDs(ctx context.Context, tags []string) []primitive.ObjectID {
	var tagIDs []primitive.ObjectID
	for _, tag := range tags {
		tagModels, err := QueryTagsByName(ctx, tag)
		if err == nil {
			for _, tagModel := range tagModels {
				tagIDs = append(tagIDs, tagModel.ID)
			}
		}
	}
	return tagIDs
}

func getArtistIDs(ctx context.Context, artists []string) []primitive.ObjectID {
	var artistIDs []primitive.ObjectID
	for _, artist := range artists {
		artistModels, err := QueryArtistsByName(ctx, artist)
		if err == nil {
			for _, artistModel := range artistModels {
				artistIDs = append(artistIDs, artistModel.ID)
			}
		}
		artistModelsByUsername, err := QueryArtistsByUserName(ctx, artist)
		if err == nil {
			for _, artistModel := range artistModelsByUsername {
				artistIDs = append(artistIDs, artistModel.ID)
			}
		}
	}
	return artistIDs
}
