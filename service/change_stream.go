package service

import (
	"context"
	"time"

	"github.com/goccy/go-json"
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

func syncArtworkToSearchEngine(ctx context.Context) {
	changeStream, err := GetArtworkChangeStream(ctx)
	if err != nil {
		common.Logger.Fatalf("get artwork change stream error: %s", err)
	}
	manager := &artworkSyncManager{
		ctx:          ctx,
		changeStream: changeStream,
	}
	manager.Start()
}

type artworkSyncManager struct {
	ctx          context.Context
	changeStream *mongo.ChangeStream
}

func (m *artworkSyncManager) Start() {
	defer m.Close()
	for {
		select {
		case <-m.ctx.Done():
			common.Logger.Info("artwork sync manager context done, exiting...")
			return
		default:
			ok := m.changeStream.Next(m.ctx)
			if !ok {
				if err := m.changeStream.Err(); err != nil {
					common.Logger.Errorf("change stream error: %s", err)
				} else {
					// [TODO] resume the change stream
					common.Logger.Info("change stream closed")
				}
				return
			}
			var event bson.M
			if err := m.changeStream.Decode(&event); err != nil {
				common.Logger.Errorf("decode change stream error: %s", err)
				continue
			}
			m.ProcessArtworkChangeEvent(event)
		}

	}
}

func (m *artworkSyncManager) Close() {
	if err := m.changeStream.Close(m.ctx); err != nil {
		common.Logger.Errorf("close change stream error: %s", err)
	} else {
		common.Logger.Info("change stream closed successfully")
	}
}

func (m *artworkSyncManager) ProcessArtworkChangeEvent(event bson.M) {
	defer func() {
		if r := recover(); r != nil {
			common.Logger.Fatalf("panic when processing artwork change event: %s", r)
		}
	}()
	operationType := event["operationType"].(string)
	switch operationType {
	case "update":
		m.ProcessArtworkUpdateEvent(event)
	case "delete":
		m.ProcessArtworkDeleteEvent(event)
	case "replace":
		m.ProcessArtworkReplaceEvent(event)
	case "insert":
		// do nothing
	default:
		common.Logger.Debugf("unknown operation type: %s", operationType)
	}
}

func (m *artworkSyncManager) DecodeArtworkFromEvent(event bson.M) (*types.ArtworkModel, error) {
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

func (m *artworkSyncManager) ProcessArtworkUpdateEvent(event bson.M) {
	artwork, err := m.DecodeArtworkFromEvent(event)
	if err != nil {
		common.Logger.Errorf("decode artwork from event error: %s", err)
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	searchDoc, err := adapter.ConvertToSearchDoc(ctx, artwork)
	if err != nil {
		common.Logger.Errorf("convert to search doc error: %s", err)
		return
	}
	artworkJSON, err := json.Marshal(searchDoc)
	if err != nil {
		common.Logger.Errorf("marshal search doc error: %s", err)
		return
	}
	primKey := "id"
	task, err := common.MeilisearchClient.Index(config.Cfg.Search.MeiliSearch.Index).UpdateDocumentsWithContext(ctx, artworkJSON, &primKey)
	if err != nil {
		common.Logger.Errorf("update artwork to meilisearch error: %s", err)
		return
	}
	common.Logger.Debugf("commited update artwork task to meilisearch: %d", task.TaskUID)
}

func (m *artworkSyncManager) ProcessArtworkDeleteEvent(event bson.M) {
	docID := event["documentKey"].(bson.M)["_id"].(primitive.ObjectID).Hex()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	task, err := common.MeilisearchClient.Index(config.Cfg.Search.MeiliSearch.Index).DeleteDocumentWithContext(ctx, docID)
	if err != nil {
		common.Logger.Errorf("delete artwork from meilisearch error: %s", err)
		return
	}
	common.Logger.Debugf("commited delete artwork task to meilisearch: %d", task.TaskUID)
}

func (m *artworkSyncManager) ProcessArtworkReplaceEvent(event bson.M) {
	// just update the artwork?
	artwork, err := m.DecodeArtworkFromEvent(event)
	if err != nil {
		common.Logger.Errorf("decode artwork from event error: %s", err)
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	searchDoc, err := adapter.ConvertToSearchDoc(ctx, artwork)
	if err != nil {
		common.Logger.Errorf("convert to search doc error: %s", err)
		return
	}
	artworkJSON, err := json.Marshal(searchDoc)
	if err != nil {
		common.Logger.Errorf("marshal search doc error: %s", err)
		return
	}
	primKey := "id"
	task, err := common.MeilisearchClient.Index(config.Cfg.Search.MeiliSearch.Index).UpdateDocumentsWithContext(ctx, artworkJSON, &primKey)
	if err != nil {
		common.Logger.Errorf("update artwork to meilisearch error: %s", err)
		return
	}
	common.Logger.Debugf("commited update artwork task to meilisearch: %d", task.TaskUID)
}
