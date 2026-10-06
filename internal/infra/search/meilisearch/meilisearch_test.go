package meilisearch

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"slices"
	"testing"

	"github.com/meilisearch/meilisearch-go"
	"github.com/unvgo/ouid"
)

func TestGetMissingArtworkIDsHandlesDocumentPagination(t *testing.T) {
	ids := make([]ouid.OUID, 4)
	for i := range ids {
		var err error
		ids[i], err = ouid.FromObjectIDHex(fmt.Sprintf("%024x", i+1))
		if err != nil {
			t.Fatal(err)
		}
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req meilisearch.DocumentsQuery
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Error(err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		if req.Offset == 0 {
			fmt.Fprintf(w, `{"results":[{"id":%q}],"offset":0,"limit":1,"total":2}`, ids[0].Hex())
		} else {
			fmt.Fprintf(w, `{"results":[{"id":%q}],"offset":1,"limit":1,"total":2}`, ids[2].Hex())
		}
	}))
	defer server.Close()
	searcher := &SearcherMeilisearch{client: meilisearch.New(server.URL).Index("artworks")}
	missing, err := searcher.GetMissingArtworkIDs(context.Background(), ids)
	if err != nil {
		t.Fatal(err)
	}
	if want := []ouid.OUID{ids[1], ids[3]}; !slices.Equal(missing, want) {
		t.Fatalf("missing IDs = %v, want %v", missing, want)
	}
}

func TestGetMissingArtworkIDsRejectsIncompleteOrFailedQuery(t *testing.T) {
	id := ouid.New()
	for _, tc := range []struct {
		name   string
		status int
		body   string
	}{
		{"incomplete page", http.StatusOK, `{"results":[],"offset":0,"limit":1,"total":1}`},
		{"invalid document", http.StatusOK, `{"results":[{}],"offset":0,"limit":1,"total":1}`},
		{"unavailable index", http.StatusNotFound, `{"message":"index missing","code":"index_not_found","type":"invalid_request","link":""}`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tc.status)
				fmt.Fprint(w, tc.body)
			}))
			defer server.Close()
			searcher := &SearcherMeilisearch{client: meilisearch.New(server.URL).Index("artworks")}
			if missing, err := searcher.GetMissingArtworkIDs(context.Background(), []ouid.OUID{id}); err == nil || len(missing) != 0 {
				t.Fatalf("lookup = (%v, %v), want error without missing IDs", missing, err)
			}
		})
	}
}
