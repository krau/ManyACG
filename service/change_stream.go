package service

import (
	"context"

	"github.com/bytedance/sonic"
	"github.com/krau/ManyACG/adapter"
	"github.com/krau/ManyACG/common"
	"github.com/krau/ManyACG/config"
	"github.com/krau/ManyACG/dao"
	"github.com/krau/ManyACG/types"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func GetArtworkChangeStream(ctx context.Context) (*mongo.ChangeStream, error) {
	collection := dao.GetCollection("Artworks")
	opts := options.ChangeStream().SetFullDocument(options.UpdateLookup)
	pipeline := mongo.Pipeline{}
	changeStream, err := collection.Watch(ctx, pipeline, opts)
	if err != nil {
		return nil, err
	}
	return changeStream, nil
}

func syncArtworkChangeStream() {
	ctx := context.Background()
	changeStream, err := GetArtworkChangeStream(ctx)
	if err != nil {
		common.Logger.Errorf("get artwork change stream error: %s", err)
		return
	}
	defer changeStream.Close(ctx)
	for changeStream.Next(ctx) {
		var event bson.M
		if err := changeStream.Decode(&event); err != nil {
			common.Logger.Errorf("decode change stream error: %s", err)
			continue
		}
		processArtworkChangeEvent(event)
	}
}

func processArtworkChangeEvent(event bson.M) {
	defer func() {
		if r := recover(); r != nil {
			common.Logger.Errorf("panic when processing artwork change event: %s", r)
		}
	}()
	operationType := event["operationType"].(string)
	switch operationType {
	case "insert", "update":
		processArtworkUpdateEvent(event)
	case "delete":
		processArtworkDeleteEvent(event)
	case "replace":
		processArtworkReplaceEvent(event)
	default:
		common.Logger.Debugf("unknown operation type: %s", operationType)
	}
}

func decodeArtworkFromEvent(event bson.M) (*types.ArtworkModel, error) {
	doc := event["fullDocument"].(bson.M)
	var artwork types.ArtworkModel
	docBytes, err := bson.Marshal(doc)
	if err != nil {
		return nil, err
	}
	if err := bson.Unmarshal(docBytes, &artwork); err != nil {
		return nil, err
	}
	return &artwork, nil
}

func processArtworkUpdateEvent(event bson.M) {
	artwork, err := decodeArtworkFromEvent(event)
	if err != nil {
		common.Logger.Errorf("decode artwork from event error: %s", err)
		return
	}
	searchDoc, err := adapter.ConvertToSearchDoc(context.Background(), artwork)
	if err != nil {
		common.Logger.Errorf("convert to search doc error: %s", err)
		return
	}
	artworkJSON, err := sonic.Marshal(searchDoc)
	if err != nil {
		common.Logger.Errorf("marshal search doc error: %s", err)
		return
	}
	task, err := common.MeilisearchClient.Index(config.Cfg.Search.MeiliSearch.Index).UpdateDocuments(artworkJSON)
	if err != nil {
		common.Logger.Errorf("add artwork to meilisearch error: %s", err)
		return
	}
	common.Logger.Debugf("commited add artwork task to meilisearch: %d", task.TaskUID)
}

func processArtworkDeleteEvent(event bson.M) {
	docID := event["documentKey"].(bson.M)["_id"].(primitive.ObjectID).Hex()
	task, err := common.MeilisearchClient.Index(config.Cfg.Search.MeiliSearch.Index).DeleteDocument(docID)
	if err != nil {
		common.Logger.Errorf("delete artwork from meilisearch error: %s", err)
		return
	}
	common.Logger.Debugf("commited delete artwork task to meilisearch: %d", task.TaskUID)
}

func processArtworkReplaceEvent(event bson.M) {
	// just update the artwork?
	artwork, err := decodeArtworkFromEvent(event)
	if err != nil {
		common.Logger.Errorf("decode artwork from event error: %s", err)
		return
	}
	searchDoc, err := adapter.ConvertToSearchDoc(context.Background(), artwork)
	if err != nil {
		common.Logger.Errorf("convert to search doc error: %s", err)
		return
	}
	artworkJSON, err := sonic.Marshal(searchDoc)
	if err != nil {
		common.Logger.Errorf("marshal search doc error: %s", err)
		return
	}
	task, err := common.MeilisearchClient.Index(config.Cfg.Search.MeiliSearch.Index).UpdateDocuments(artworkJSON)
	if err != nil {
		common.Logger.Errorf("update artwork to meilisearch error: %s", err)
		return
	}
	common.Logger.Debugf("commited update artwork task to meilisearch: %d", task.TaskUID)
}
