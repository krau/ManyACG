package database

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"testing"
	"time"

	"github.com/krau/ManyACG/internal/model/entity"
	"github.com/krau/ManyACG/internal/model/query"
	"github.com/krau/ManyACG/internal/shared"
	"github.com/ncruces/go-sqlite3/gormlite"
	"github.com/unvgo/ouid"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func newArtworkQueryTestDB(tb testing.TB) *DB {
	tb.Helper()
	gdb, err := gorm.Open(gormlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		tb.Fatal(err)
	}
	sqlDB, err := gdb.DB()
	if err != nil {
		tb.Fatal(err)
	}
	sqlDB.SetMaxOpenConns(1)
	tb.Cleanup(func() { _ = sqlDB.Close() })
	if err := gdb.AutoMigrate(
		&entity.Artist{}, &entity.Tag{}, &entity.TagAlias{}, &entity.Artwork{},
		&entity.Picture{}, &entity.Video{}, &entity.UgoiraMeta{},
	); err != nil {
		tb.Fatal(err)
	}
	return &DB{db: gdb}
}

func artworkQueryTestID(tb testing.TB, n int) ouid.OUID {
	tb.Helper()
	id, err := ouid.FromObjectIDHex(fmt.Sprintf("%024x", n))
	if err != nil {
		tb.Fatal(err)
	}
	return id
}

func artworkQueryTestTime(n int) time.Time {
	return time.Date(2025, time.January, 1, 0, 0, n, 0, time.UTC)
}

func artworkQueryTestInsert(tb testing.TB, repository *DB, value any) {
	tb.Helper()
	if err := repository.db.Create(value).Error; err != nil {
		tb.Fatal(err)
	}
}

func artworkKeywordQuery(groups [][]string) query.ArtworksDB {
	return query.ArtworksDB{
		R18: shared.R18TypeAll, Keywords: groups}
}

func seedArtworkKeywordFixtures(tb testing.TB) *DB {
	tb.Helper()
	repository := newArtworkQueryTestDB(tb)
	id := func(n int) ouid.OUID { return artworkQueryTestID(tb, n) }
	neutral := &entity.Artist{ID: id(100), Name: "Neutral creator", Type: "pixiv", UID: "neutral", Username: "neutral"}
	beta := &entity.Artist{ID: id(101), Name: "artistBeta", Type: "pixiv", UID: "beta", Username: "beta"}
	red := &entity.Tag{ID: id(200), Name: "red-tag"}
	blue := &entity.Tag{ID: id(201), Name: "blue-tag"}
	green := &entity.Tag{ID: id(202), Name: "green-tag"}
	gamma := &entity.Tag{
		ID: id(203), Name: "tagGamma",
		Alias: []entity.TagAlias{
			{ID: id(300), TagID: id(203), Alias: "aliasAlpha"},
			{ID: id(301), TagID: id(203), Alias: "aliasEpsilon"},
		},
	}
	duplicate := &entity.Tag{
		ID: id(204), Name: "alpha gamma",
		Alias: []entity.TagAlias{
			{ID: id(302), TagID: id(204), Alias: "alpha gamma"},
			{ID: id(303), TagID: id(204), Alias: "alpha gamma repeated"},
		},
	}
	otherGamma := &entity.Tag{
		ID: id(205), Name: "tagGammaOther",
		Alias: []entity.TagAlias{{ID: id(304), TagID: id(205), Alias: "aliasEpsilonOther"}},
	}
	artworks := []*entity.Artwork{
		{ID: id(1), Title: "prefix-alpha-suffix", Description: "descGamma", Artist: neutral, ArtistID: neutral.ID, Tags: []*entity.Tag{red, blue}, Pictures: []*entity.Picture{{ID: id(400)}}},
		{ID: id(2), Title: "beta frame", Description: "delta trail", Artist: neutral, ArtistID: neutral.ID, Tags: []*entity.Tag{blue, green}, R18: true, Videos: []*entity.Video{{ID: id(401)}}},
		{ID: id(3), Title: "alpha alone", Description: "quiet details", Artist: neutral, ArtistID: neutral.ID, Tags: []*entity.Tag{red}},
		{ID: id(4), Title: "gamma only", Description: "quiet details", Artist: neutral, ArtistID: neutral.ID, Tags: []*entity.Tag{green}, UgoiraMetas: []*entity.UgoiraMeta{{ID: id(402)}}},
		{ID: id(5), Title: "quiet frame", Description: "delta journal", Artist: beta, ArtistID: beta.ID, Tags: []*entity.Tag{gamma, red, green}, Pictures: []*entity.Picture{{ID: id(403)}}, UgoiraMetas: []*entity.UgoiraMeta{{ID: id(404)}}},
		{ID: id(6), Title: "alpha beta duplicate", Description: "gamma delta duplicate", Artist: beta, ArtistID: beta.ID, Tags: []*entity.Tag{duplicate, red, blue, green}, R18: true, Pictures: []*entity.Picture{{ID: id(405)}, {ID: id(406)}}, Videos: []*entity.Video{{ID: id(407)}, {ID: id(408)}}, UgoiraMetas: []*entity.UgoiraMeta{{ID: id(409)}, {ID: id(410)}}},
		{ID: id(7), Title: "quiet frame", Description: "quiet details", Artist: beta, ArtistID: beta.ID, Tags: []*entity.Tag{otherGamma, blue}, Videos: []*entity.Video{{ID: id(411)}}},
		{ID: id(8), Title: "unrelated", Description: "quiet details", Artist: neutral, ArtistID: neutral.ID, R18: true, Pictures: []*entity.Picture{{ID: id(412)}}, Videos: []*entity.Video{{ID: id(413)}}, UgoiraMetas: []*entity.UgoiraMeta{{ID: id(414)}}},
	}
	for i, artwork := range artworks {
		artwork.CreatedAt = artworkQueryTestTime(i + 1)
		artwork.UpdatedAt = artwork.CreatedAt
		artwork.SourceType = "pixiv"
		artwork.SourceURL = fmt.Sprintf("https://example.test/keyword/%d", i+1)
		for _, picture := range artwork.Pictures {
			picture.CreatedAt, picture.UpdatedAt = artwork.CreatedAt, artwork.CreatedAt
		}
		for _, video := range artwork.Videos {
			video.CreatedAt, video.UpdatedAt = artwork.CreatedAt, artwork.CreatedAt
		}
		for _, ugoira := range artwork.UgoiraMetas {
			ugoira.CreatedAt, ugoira.UpdatedAt = artwork.CreatedAt, artwork.CreatedAt
		}
		artworkQueryTestInsert(tb, repository, artwork)
	}
	return repository
}

func assertArtworkKeywordResult(t *testing.T, repository *DB, que query.ArtworksDB, want []int, wantCount int64) {
	t.Helper()
	artworks, err := repository.QueryArtworks(context.Background(), que)
	if err != nil {
		t.Fatal(err)
	}
	gotIDs := make([]string, len(artworks))
	for i, artwork := range artworks {
		gotIDs[i] = artwork.ID.Hex()
	}
	wantIDs := make([]string, len(want))
	for i, n := range want {
		wantIDs[i] = artworkQueryTestID(t, n).Hex()
	}
	if !slices.Equal(gotIDs, wantIDs) {
		t.Fatalf("artwork IDs = %v, want %v", gotIDs, wantIDs)
	}
	count, err := repository.CountArtworks(context.Background(), que)
	if err != nil {
		t.Fatal(err)
	}
	if count != wantCount {
		t.Fatalf("CountArtworks = %d, want %d", count, wantCount)
	}
}

func TestArtworksKeywordGroups(t *testing.T) {
	repository := seedArtworkKeywordFixtures(t)
	tests := []struct {
		name   string
		groups [][]string
		want   []int
	}{
		{"two_OR_groups", [][]string{{"alpha", "beta"}, {"gamma", "delta"}}, []int{7, 6, 5, 2, 1}},
		{"three_OR_groups", [][]string{{"alpha", "beta"}, {"gamma", "delta"}, {"red-tag", "aliasEpsilon"}}, []int{7, 6, 5, 1}},
		{"overlapping_OR_and_association_hits", [][]string{{"alpha", "alpha", "beta"}, {"gamma", "gamma", "delta"}}, []int{7, 6, 5, 2, 1}},
		{"artist_AND_tag", [][]string{{"artistBeta", "missing-author"}, {"tagGamma", "missing-tag"}}, []int{7, 5}},
		{"alias_AND_description", [][]string{{"aliasAlpha", "missing-alias"}, {"journal", "missing-description"}}, []int{5}},
		{"three_association_groups", [][]string{{"artistBeta", "missing-author"}, {"tagGamma", "missing-tag"}, {"aliasEpsilon", "missing-alias"}}, []int{7, 5}},
		{"title_field", [][]string{{"prefix-alpha-suffix"}}, []int{1}},
		{"description_field", [][]string{{"journal"}}, []int{5}},
		{"artist_field", [][]string{{"artistBeta"}}, []int{7, 6, 5}},
		{"tag_field", [][]string{{"tagGamma"}}, []int{7, 5}},
		{"alias_field", [][]string{{"aliasAlpha"}}, []int{5}},
		{"middle_substring", [][]string{{"pha"}, {"alone", "journal", "duplicate", "suffix"}}, []int{6, 5, 3, 1}},
		{"empty_terms_preserve_OR_AND", [][]string{{""}, {"", "alpha"}, {}, {"gamma", ""}}, []int{6, 5, 1}},
		{"empty_OR_does_not_match_everything", [][]string{{"", "missing-keyword"}, {"alpha", "beta"}}, nil},
		{"all_empty_groups_are_ignored", [][]string{nil, {}, {"", ""}}, []int{8, 7, 6, 5, 4, 3, 2, 1}},
		{"no_groups", nil, []int{8, 7, 6, 5, 4, 3, 2, 1}},
		{"no_match_in_one_group", [][]string{{"alpha", "beta"}, {"missing-one", "missing-two"}}, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertArtworkKeywordResult(t, repository, artworkKeywordQuery(tt.groups), tt.want, int64(len(tt.want)))
		})
	}
}

func TestArtworksKeywordPaginationAndCount(t *testing.T) {
	repository := seedArtworkKeywordFixtures(t)
	groups := [][]string{{"alpha", "beta"}, {"gamma", "delta"}}
	for _, tt := range []struct {
		name   string
		limit  int
		offset int
		want   []int
	}{
		{"first_page", 2, 0, []int{7, 6}},
		{"middle_page", 2, 2, []int{5, 2}},
		{"last_page", 2, 4, []int{1}},
		{"past_end", 2, 5, nil},
		{"offset_without_limit", 0, 2, []int{5, 2, 1}},
		{"unbounded", 0, 0, []int{7, 6, 5, 2, 1}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			que := artworkKeywordQuery(groups)
			que.Limit, que.Offset = tt.limit, tt.offset
			assertArtworkKeywordResult(t, repository, que, tt.want, 5)
		})
	}
}

func TestArtworksKeywordBaseFilters(t *testing.T) {
	repository := seedArtworkKeywordFixtures(t)
	id := func(n int) ouid.OUID { return artworkQueryTestID(t, n) }
	groups := [][]string{{"alpha", "beta"}, {"gamma", "delta"}}
	tests := []struct {
		name  string
		apply func(*query.ArtworksDB)
		want  []int
	}{
		{"safe", func(q *query.ArtworksDB) { q.R18 = shared.R18TypeNone }, []int{7, 5, 1}},
		{"R18", func(q *query.ArtworksDB) { q.R18 = shared.R18TypeR18 }, []int{6, 2}},
		{"artist", func(q *query.ArtworksDB) { q.ArtistID = id(101) }, []int{7, 6, 5}},
		{"tag_OR_AND", func(q *query.ArtworksDB) { q.Tags = [][]ouid.OUID{{id(200), id(201)}, {id(202)}} }, []int{6, 5, 2}},
		{"empty_tag_groups", func(q *query.ArtworksDB) { q.Tags = [][]ouid.OUID{nil, {}, {id(200)}} }, []int{6, 5, 1}},
		{"tag_mismatch", func(q *query.ArtworksDB) { q.Tags = [][]ouid.OUID{{id(200), id(201)}, {id(999)}} }, nil},
		{"picture", func(q *query.ArtworksDB) { q.HasPicture = true }, []int{6, 5, 1}},
		{"video", func(q *query.ArtworksDB) { q.HasVideo = true }, []int{7, 6, 2}},
		{"ugoira", func(q *query.ArtworksDB) { q.HasUgoira = true }, []int{6, 5}},
		{"picture_AND_ugoira", func(q *query.ArtworksDB) { q.HasPicture, q.HasUgoira = true, true }, []int{6, 5}},
		{"all_media", func(q *query.ArtworksDB) { q.HasPicture, q.HasVideo, q.HasUgoira = true, true, true }, []int{6}},
		{"all_filters", func(q *query.ArtworksDB) {
			q.R18 = shared.R18TypeR18
			q.ArtistID = id(101)
			q.Tags = [][]ouid.OUID{{id(200), id(201)}, {id(202)}}
			q.HasPicture, q.HasVideo, q.HasUgoira = true, true, true
		}, []int{6}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			que := artworkKeywordQuery(groups)
			tt.apply(&que)
			assertArtworkKeywordResult(t, repository, que, tt.want, int64(len(tt.want)))
		})
	}
}

func TestArtworksKeywordsLiteralSubstrings(t *testing.T) {
	repository := newArtworkQueryTestDB(t)
	id := func(n int) ouid.OUID { return artworkQueryTestID(t, n) }
	fields := []string{"title", "description", "artist", "tag", "alias"}
	for variant, text := range []string{`prefix50%_!\雪中间suffix !% \_ \\ end`, "prefix50ABQplainmiddleSuffix"} {
		for fieldIndex, field := range fields {
			n := variant*len(fields) + fieldIndex + 1
			artist := &entity.Artist{ID: id(100 + n), Name: "plain creator", Type: "pixiv", UID: fmt.Sprint(n), Username: "plain"}
			tag := &entity.Tag{ID: id(200 + n), Name: fmt.Sprintf("plain-tag-%d", n)}
			artwork := &entity.Artwork{
				ID: id(n), Title: "plain title", Description: "plain details", Artist: artist, ArtistID: artist.ID,
				Tags: []*entity.Tag{tag}, SourceType: "pixiv", SourceURL: fmt.Sprintf("https://example.test/literal/%d", n),
				CreatedAt: artworkQueryTestTime(n), UpdatedAt: artworkQueryTestTime(n),
			}
			switch field {
			case "title":
				artwork.Title = text
			case "description":
				artwork.Description = text
			case "artist":
				artist.Name = text
			case "tag":
				tag.Name = text
			case "alias":
				tag.Alias = []entity.TagAlias{{ID: id(300 + n), TagID: tag.ID, Alias: text}}
			}
			artworkQueryTestInsert(t, repository, artwork)
		}
	}
	want := []int{5, 4, 3, 2, 1}
	for _, tt := range []struct {
		name string
		term string
	}{
		{"percent", "%"},
		{"underscore", "_"},
		{"escape_character", "!"},
		{"backslash", `\`},
		{"combined_metacharacters", `50%_!\`},
		{"escape_before_percent", "!%"},
		{"backslash_before_underscore", `\_`},
		{"two_backslashes", `\\`},
		{"unicode", "雪中间"},
		{"unicode_middle_substring", "间suf"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			groups := [][]string{{tt.term, "absent-first"}, {"prefix", "absent-second"}}
			assertArtworkKeywordResult(t, repository, artworkKeywordQuery(groups), want, 5)
		})
	}
	t.Run("three_literal_OR_groups", func(t *testing.T) {
		groups := [][]string{{"雪", "absent-first"}, {"中间", "absent-second"}, {`50%_!\`, "absent-third"}}
		assertArtworkKeywordResult(t, repository, artworkKeywordQuery(groups), want, 5)
	})
}

func assertArtworkRandomMatches(t *testing.T, repository *DB, que query.ArtworksDB, eligible []int, wantLength int) {
	t.Helper()
	que.Random = true
	artworks, err := repository.QueryArtworks(context.Background(), que)
	if err != nil {
		t.Fatal(err)
	}
	if len(artworks) != wantLength {
		t.Fatalf("random result length = %d, want %d", len(artworks), wantLength)
	}
	allowed := make(map[string]bool, len(eligible))
	for _, n := range eligible {
		allowed[artworkQueryTestID(t, n).Hex()] = true
	}
	seen := make(map[string]bool, len(artworks))
	for _, artwork := range artworks {
		id := artwork.ID.Hex()
		if !allowed[id] {
			t.Fatalf("random artwork %s does not satisfy all filters", id)
		}
		if seen[id] {
			t.Fatalf("random result repeats artwork %s", id)
		}
		seen[id] = true
	}
	count, err := repository.CountArtworks(context.Background(), que)
	if err != nil {
		t.Fatal(err)
	}
	if count != int64(len(eligible)) {
		t.Fatalf("random query count = %d, want %d", count, len(eligible))
	}
}

func TestArtworksKeywordRandom(t *testing.T) {
	repository := seedArtworkKeywordFixtures(t)
	for _, limit := range []int{1, 3, 20} {
		t.Run(fmt.Sprintf("two_groups_limit_%d", limit), func(t *testing.T) {
			que := artworkKeywordQuery([][]string{{"alpha", "beta"}, {"gamma", "delta"}})
			que.Limit = limit
			assertArtworkRandomMatches(t, repository, que, []int{1, 2, 5, 6, 7}, min(limit, 5))
		})
	}
	t.Run("three_groups", func(t *testing.T) {
		que := artworkKeywordQuery([][]string{{"alpha", "beta"}, {"gamma", "delta"}, {"red-tag", "aliasEpsilon"}})
		que.Limit = 2
		assertArtworkRandomMatches(t, repository, que, []int{1, 5, 6, 7}, 2)
	})
	t.Run("association_groups", func(t *testing.T) {
		que := artworkKeywordQuery([][]string{{"artistBeta", "missing-author"}, {"tagGamma", "missing-tag"}, {"aliasEpsilon", "missing-alias"}})
		que.Limit = 1
		assertArtworkRandomMatches(t, repository, que, []int{5, 7}, 1)
	})
	t.Run("all_filters", func(t *testing.T) {
		que := artworkKeywordQuery([][]string{{"alpha", "beta"}, {"gamma", "delta"}})
		que.Limit = 4
		que.R18 = shared.R18TypeR18
		que.ArtistID = artworkQueryTestID(t, 101)
		que.Tags = [][]ouid.OUID{{artworkQueryTestID(t, 200), artworkQueryTestID(t, 201)}, {artworkQueryTestID(t, 202)}}
		que.HasPicture, que.HasVideo, que.HasUgoira = true, true, true
		assertArtworkRandomMatches(t, repository, que, []int{6}, 1)
	})
	t.Run("empty_results", func(t *testing.T) {
		que := artworkKeywordQuery([][]string{{"alpha", "beta"}, {"missing-one", "missing-two"}})
		que.Random, que.Limit = true, 3
		artworks, err := repository.QueryArtworks(context.Background(), que)
		if !errors.Is(err, gorm.ErrRecordNotFound) || len(artworks) != 0 {
			t.Fatalf("empty random result = %v, error = %v; want no artworks and ErrRecordNotFound", artworks, err)
		}
		count, err := repository.CountArtworks(context.Background(), que)
		if err != nil || count != 0 {
			t.Fatalf("empty random CountArtworks = %d, error = %v; want 0, nil", count, err)
		}
	})
}

func seedArtworkKeywordBenchmark(tb testing.TB, size int) *DB {
	tb.Helper()
	repository := newArtworkQueryTestDB(tb)
	id := func(n int) ouid.OUID { return artworkQueryTestID(tb, n) }
	artists := make([]*entity.Artist, 24)
	tags := make([]*entity.Tag, 24)
	for i := range artists {
		name := fmt.Sprintf("common-artist creator-%d", i)
		if i%2 == 0 {
			name += " artist-a"
		}
		if i == 0 {
			name += " rare-artist"
		}
		artists[i] = &entity.Artist{ID: id(10000 + i), Name: name, Type: "pixiv", UID: fmt.Sprint(i), Username: "creator"}
		tagName := fmt.Sprintf("common-tag topic-%d", i)
		alias := "common-alias"
		if i%3 == 0 {
			tagName += " tag-d"
		}
		if i == 0 {
			tagName += " rare-tag"
			alias += " rare-alias"
		}
		tags[i] = &entity.Tag{ID: id(11000 + i), Name: tagName, Alias: []entity.TagAlias{{ID: id(12000 + i), TagID: id(11000 + i), Alias: alias}}}
	}
	artworkQueryTestInsert(tb, repository, &artists)
	artworkQueryTestInsert(tb, repository, &tags)
	artworks := make([]*entity.Artwork, size)
	links := make([]struct {
		ArtworkID ouid.OUID
		TagID     ouid.OUID
	}, size)
	for i := range artworks {
		title := fmt.Sprintf("common-title frame-%d", i)
		description := fmt.Sprintf("common-description details-%d", i)
		if i%2 == 0 {
			title += " title-a"
		}
		if i%3 != 0 {
			description += " description-c"
		}
		if i%97 == 0 {
			title += " rare-title"
		}
		if i%89 == 0 {
			description += " rare-description"
		}
		if i%101 == 0 {
			title += " rare-third"
		}
		artworks[i] = &entity.Artwork{
			ID: id(i + 1), Title: title, Description: description, ArtistID: artists[i%len(artists)].ID,
			SourceType: "pixiv", SourceURL: fmt.Sprintf("https://example.test/benchmark/%d", i),
			CreatedAt: artworkQueryTestTime(i + 1), UpdatedAt: artworkQueryTestTime(i + 1),
		}
		links[i].ArtworkID, links[i].TagID = artworks[i].ID, tags[i%23].ID
	}
	if err := repository.db.CreateInBatches(&artworks, 128).Error; err != nil {
		tb.Fatal(err)
	}
	if err := repository.db.Table("artwork_tags").CreateInBatches(&links, 128).Error; err != nil {
		tb.Fatal(err)
	}
	return repository
}

func TestArtworksKeywordRandomLargeMatchSet(t *testing.T) {
	const size = 1536
	repository := seedArtworkKeywordBenchmark(t, size)
	eligible := make([]int, 0, size)
	for i := range size {
		if i%3 != 0 || (i%23)%3 == 0 {
			eligible = append(eligible, i+1)
		}
	}
	for _, limit := range []int{7, size} {
		t.Run(fmt.Sprintf("limit_%d", limit), func(t *testing.T) {
			que := artworkKeywordQuery([][]string{{"common-title", "common-artist"}, {"description-c", "tag-d"}})
			que.Limit = limit
			assertArtworkRandomMatches(t, repository, que, eligible, min(limit, len(eligible)))
		})
	}
}

func BenchmarkArtworksKeywordGroups(b *testing.B) {
	repository := seedArtworkKeywordBenchmark(b, 1536)
	ctx := context.Background()
	cases := []struct {
		name   string
		groups [][]string
	}{
		{"SparseThreeGroups", [][]string{{"rare-title", "rare-artist"}, {"rare-description", "rare-tag"}, {"rare-alias", "rare-third"}}},
		{"MixedTwoGroups", [][]string{{"title-a", "artist-a"}, {"description-c", "tag-d"}}},
		{"HighFrequencyThreeGroups", [][]string{{"common-title", "common-artist"}, {"common-description", "common-tag"}, {"common-alias", "common-title"}}},
		{"NoMatchThreeGroups", [][]string{{"common-title", "common-artist"}, {"absent-description", "absent-tag"}, {"common-alias", "common-tag"}}},
	}
	for _, tc := range cases {
		b.Run(tc.name, func(b *testing.B) {
			for _, operation := range []string{"List", "Count", "Random"} {
				b.Run(operation, func(b *testing.B) {
					que := artworkKeywordQuery(tc.groups)
					que.Limit = 20
					que.Random = operation == "Random"
					b.ReportAllocs()
					b.ResetTimer()
					for b.Loop() {
						if operation == "Count" {
							if _, err := repository.CountArtworks(ctx, que); err != nil {
								b.Fatal(err)
							}
							continue
						}
						_, err := repository.QueryArtworks(ctx, que)
						if que.Random && tc.name == "NoMatchThreeGroups" {
							if !errors.Is(err, gorm.ErrRecordNotFound) {
								b.Fatalf("empty random query error = %v, want ErrRecordNotFound", err)
							}
						} else if err != nil {
							b.Fatal(err)
						}
					}
				})
			}
		})
	}
}
