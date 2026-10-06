package database

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"slices"
	"testing"

	"github.com/krau/ManyACG/internal/model/converter"
	"github.com/krau/ManyACG/internal/model/dto"
	"github.com/krau/ManyACG/internal/model/entity"
	_ "github.com/ncruces/go-sqlite3/embed"
	"github.com/ncruces/go-sqlite3/gormlite"
	"github.com/unvgo/ouid"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestGetArtworkSearchDocuments(t *testing.T) {
	gdb, err := gorm.Open(gormlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatal(err)
	}
	sqlDB, err := gdb.DB()
	if err != nil {
		t.Fatal(err)
	}
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = sqlDB.Close() })
	if err := gdb.AutoMigrate(
		&entity.Artist{}, &entity.Tag{}, &entity.TagAlias{}, &entity.Artwork{},
		&entity.Picture{}, &entity.Video{}, &entity.UgoiraMeta{},
	); err != nil {
		t.Fatal(err)
	}
	id := func(n int) ouid.OUID {
		t.Helper()
		value, err := ouid.FromObjectIDHex(fmt.Sprintf("%024x", n))
		if err != nil {
			t.Fatal(err)
		}
		return value
	}
	artist := &entity.Artist{
		ID: id(10), Name: "作者", Type: "pixiv", UID: "123", Username: "unused-username",
	}
	tag := &entity.Tag{
		ID: id(20), Name: "original",
		Alias: []entity.TagAlias{
			{ID: id(21), TagID: id(20), Alias: "別名"},
			{ID: id(22), TagID: id(20), Alias: "original"},
		},
	}
	secondTag := &entity.Tag{
		ID: id(30), Name: "second",
		Alias: []entity.TagAlias{{ID: id(31), TagID: id(30), Alias: "translated"}},
	}
	artworks := []*entity.Artwork{
		{
			ID: id(1), Title: "First", Description: "Full description", R18: true,
			Artist: artist, ArtistID: artist.ID, Tags: []*entity.Tag{tag, secondTag},
			SourceType: "pixiv", SourceURL: "https://example.test/1", LikeCount: 9,
			Pictures:    []*entity.Picture{{ID: id(40), Original: "unused-picture"}},
			Videos:      []*entity.Video{{ID: id(41), URL: "unused-video"}},
			UgoiraMetas: []*entity.UgoiraMeta{{ID: id(42)}},
		},
		{
			ID: id(2), Title: "Second", Description: "Safe artwork",
			Artist: artist, ArtistID: artist.ID, Tags: []*entity.Tag{tag},
			SourceType: "pixiv", SourceURL: "https://example.test/2",
		},
		{
			ID: id(3), ArtistID: artist.ID, Title: "Without tags", SourceURL: "https://example.test/3",
		},
		{
			ID: id(4), ArtistID: artist.ID, Title: "Removed", SourceURL: "https://example.test/4",
		},
	}
	for _, artwork := range artworks {
		if err := gdb.Create(artwork).Error; err != nil {
			t.Fatal(err)
		}
	}
	if err := gdb.Delete(artworks[3]).Error; err != nil {
		t.Fatal(err)
	}
	repository := &DB{db: gdb}
	ids := []ouid.OUID{id(1), id(2), id(3), id(4), id(99)}
	fullArtworks, err := repository.GetArtworksByIDs(context.Background(), ids)
	if err != nil {
		t.Fatal(err)
	}
	want := make(map[string]*dto.ArtworkSearchDocument)
	for _, artwork := range fullArtworks {
		doc := converter.EntityArtworkToSearchDocument(artwork)
		slices.Sort(doc.Tags)
		want[doc.ID] = doc
	}
	if len(want) != 3 {
		t.Fatalf("full artwork fixture returned %d artworks, want 3", len(want))
	}
	if first := want[id(1).Hex()]; first.Artist != artist.Name || !first.R18 ||
		!reflect.DeepEqual(first.Tags, []string{"original", "second", "translated", "別名"}) {
		t.Fatalf("incomplete full artwork fixture: %+v", first)
	}

	queries := make(map[string]int)
	if err := gdb.Callback().Query().After("gorm:after_query").Register("test:search_document_reads", func(tx *gorm.DB) {
		queries[tx.Statement.Table]++
		loaded, ok := tx.Statement.Dest.(*[]*entity.Artwork)
		if !ok {
			return
		}
		for _, artwork := range *loaded {
			if artwork.SourceType != "" || artwork.SourceURL != "" || artwork.LikeCount != 0 ||
				!artwork.CreatedAt.IsZero() || !artwork.UpdatedAt.IsZero() {
				t.Errorf("loaded non-index artwork fields: %+v", artwork)
			}
			if len(artwork.Pictures) != 0 || len(artwork.Videos) != 0 || len(artwork.UgoiraMetas) != 0 {
				t.Error("loaded media for an index document")
			}
			if artist := artwork.Artist; artist != nil &&
				(artist.UID != "" || artist.Username != "" || artist.Type != "" || len(artist.Artworks) != 0) {
				t.Errorf("loaded non-index artist fields: %+v", artist)
			}
		}
	}); err != nil {
		t.Fatal(err)
	}

	docs, err := repository.GetArtworkSearchDocuments(context.Background(), ids)
	if err != nil {
		t.Fatal(err)
	}
	got := make(map[string]*dto.ArtworkSearchDocument)
	for _, doc := range docs {
		slices.Sort(doc.Tags)
		got[doc.ID] = doc
	}
	if len(docs) != len(want) || !reflect.DeepEqual(got, want) {
		t.Fatalf("search documents = %+v, want %+v", got, want)
	}
	wantQueries := map[string]int{
		"artworks": 1, "artists": 1, "artwork_tags": 1, "tags": 1, "tag_aliases": 1,
	}
	if !reflect.DeepEqual(queries, wantQueries) {
		t.Fatalf("batch query counts = %v, want %v", queries, wantQueries)
	}

	clear(queries)
	docs, err = repository.GetArtworkSearchDocuments(context.Background(), nil)
	if err != nil || len(docs) != 0 || len(queries) != 0 {
		t.Fatalf("empty IDs returned docs=%v err=%v queries=%v", docs, err, queries)
	}
	docs, err = repository.GetArtworkSearchDocuments(context.Background(), []ouid.OUID{id(4), id(99)})
	if err != nil || len(docs) != 0 {
		t.Fatalf("removed/missing IDs returned docs=%v err=%v", docs, err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := repository.GetArtworkSearchDocuments(ctx, []ouid.OUID{id(1)}); !errors.Is(err, context.Canceled) {
		t.Fatalf("canceled read returned %v, want context.Canceled", err)
	}
}
