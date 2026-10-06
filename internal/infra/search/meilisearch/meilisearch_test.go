package meilisearch

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/krau/ManyACG/internal/model/dto"
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

func TestAddDocumentsAndWaitDoesNotCompleteBeforeTaskSucceeds(t *testing.T) {
	observed := make(chan string, 3)
	releaseSuccess := make(chan struct{})
	var submissions, polls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/indexes/artworks/documents":
			submissions.Add(1)
			w.WriteHeader(http.StatusAccepted)
			fmt.Fprint(w, `{"taskUid":42,"status":"enqueued"}`)
		case r.Method == http.MethodGet && r.URL.Path == "/tasks/42":
			status := "enqueued"
			switch polls.Add(1) {
			case 1:
			case 2:
				status = "processing"
			default:
				status = "succeeded"
			}
			observed <- status
			if status == "succeeded" {
				select {
				case <-releaseSuccess:
				case <-r.Context().Done():
					return
				}
			}
			fmt.Fprintf(w, `{"uid":42,"status":%q}`, status)
		default:
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	searcher := &SearcherMeilisearch{client: meilisearch.New(server.URL).Index("artworks")}
	result := make(chan error, 1)
	go func() {
		result <- searcher.AddDocumentsAndWait(ctx, []*dto.ArtworkSearchDocument{{ID: "artwork"}})
	}()
	for _, want := range []string{"enqueued", "processing", "succeeded"} {
		select {
		case status := <-observed:
			if status != want {
				t.Fatalf("observed status = %q, want %q", status, want)
			}
		case err := <-result:
			t.Fatalf("completed before task success: %v", err)
		case <-ctx.Done():
			t.Fatal(ctx.Err())
		}
	}
	select {
	case err := <-result:
		t.Fatalf("completed before success response was released: %v", err)
	default:
	}
	close(releaseSuccess)
	select {
	case err := <-result:
		if err != nil {
			t.Fatal(err)
		}
	case <-ctx.Done():
		t.Fatal(ctx.Err())
	}
	if got := submissions.Load(); got != 1 {
		t.Fatalf("submitted %d batches, want 1", got)
	}
}

func TestAddDocumentsAndWaitRejectsFailedAndCanceledTasks(t *testing.T) {
	for _, tc := range []struct {
		name    string
		status  string
		message string
		code    string
	}{
		{"failed", "failed", "Document is missing its primary key", "missing_document_id"},
		{"canceled with reason", "canceled", "Task canceled by operator", "task_canceled"},
		{"canceled without reason", "canceled", "", ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var submissions atomic.Int32
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				switch {
				case r.Method == http.MethodPost && r.URL.Path == "/indexes/artworks/documents":
					submissions.Add(1)
					w.WriteHeader(http.StatusAccepted)
					fmt.Fprint(w, `{"taskUid":42,"status":"enqueued"}`)
				case r.Method == http.MethodGet && r.URL.Path == "/tasks/42":
					fmt.Fprintf(w, `{"uid":42,"status":%q,"error":{"message":%q,"code":%q}}`, tc.status, tc.message, tc.code)
				default:
					t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
					w.WriteHeader(http.StatusNotFound)
				}
			}))
			defer server.Close()
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			searcher := &SearcherMeilisearch{client: meilisearch.New(server.URL).Index("artworks")}
			err := searcher.AddDocumentsAndWait(ctx, []*dto.ArtworkSearchDocument{{ID: "artwork"}})
			if err == nil {
				t.Fatalf("%s task reported success", tc.status)
			}
			for _, want := range []string{"42", tc.status, tc.message, tc.code} {
				if !strings.Contains(err.Error(), want) {
					t.Errorf("error %q does not contain %q", err, want)
				}
			}
			if got := submissions.Load(); got != 1 {
				t.Fatalf("submitted %d batches, want 1", got)
			}
		})
	}
}

func TestAddDocumentsAndWaitPropagatesCancellation(t *testing.T) {
	for _, inFlight := range []bool{false, true} {
		t.Run(fmt.Sprintf("in-flight poll %t", inFlight), func(t *testing.T) {
			polled := make(chan struct{})
			var pollReported atomic.Bool
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				switch {
				case r.Method == http.MethodPost && r.URL.Path == "/indexes/artworks/documents":
					w.WriteHeader(http.StatusAccepted)
					fmt.Fprint(w, `{"taskUid":42,"status":"enqueued"}`)
				case r.Method == http.MethodGet && r.URL.Path == "/tasks/42":
					if !inFlight {
						fmt.Fprint(w, `{"uid":42,"status":"enqueued"}`)
						w.(http.Flusher).Flush()
					}
					if !pollReported.Swap(true) {
						close(polled)
					}
					if inFlight {
						<-r.Context().Done()
					}
				default:
					t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
					w.WriteHeader(http.StatusNotFound)
				}
			}))
			defer server.Close()
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			searcher := &SearcherMeilisearch{client: meilisearch.New(server.URL).Index("artworks")}
			result := make(chan error, 1)
			go func() {
				result <- searcher.AddDocumentsAndWait(ctx, []*dto.ArtworkSearchDocument{{ID: "artwork"}})
			}()
			select {
			case <-polled:
			case err := <-result:
				t.Fatalf("completed before cancellation: %v", err)
			case <-ctx.Done():
				t.Fatal(ctx.Err())
			}
			cancel()
			select {
			case err := <-result:
				if !errors.Is(err, context.Canceled) {
					t.Fatalf("error = %v, want context cancellation", err)
				}
			case <-time.After(5 * time.Second):
				t.Fatal("task wait did not stop after cancellation")
			}
		})
	}
}

func TestAddDocumentsAndWaitPropagatesPollingError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method == http.MethodPost {
			w.WriteHeader(http.StatusAccepted)
			fmt.Fprint(w, `{"taskUid":42,"status":"enqueued"}`)
			return
		}
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, `{"message":"Task 42 not found","code":"task_not_found","type":"invalid_request","link":""}`)
	}))
	defer server.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	searcher := &SearcherMeilisearch{client: meilisearch.New(server.URL).Index("artworks")}
	err := searcher.AddDocumentsAndWait(ctx, []*dto.ArtworkSearchDocument{{ID: "artwork"}})
	var sdkErr *meilisearch.Error
	if !errors.As(err, &sdkErr) || sdkErr.MeilisearchApiError.Code != "task_not_found" {
		t.Fatalf("error = %v, want task_not_found polling error", err)
	}
	if !strings.Contains(err.Error(), "42") {
		t.Fatalf("error %q does not identify task 42", err)
	}
}

func TestAddDocumentsAndWaitSkipsEmptyBatches(t *testing.T) {
	var requests atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests.Add(1)
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()
	searcher := &SearcherMeilisearch{client: meilisearch.New(server.URL).Index("artworks")}
	for _, docs := range [][]*dto.ArtworkSearchDocument{nil, {}} {
		if err := searcher.AddDocumentsAndWait(context.Background(), docs); err != nil {
			t.Fatal(err)
		}
	}
	if got := requests.Load(); got != 0 {
		t.Fatalf("empty batches made %d requests, want 0", got)
	}
}
