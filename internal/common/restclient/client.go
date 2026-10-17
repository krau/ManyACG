// Package restclient 通过 REST 接口查询作品 ID。
package restclient

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/unvgo/ouid"
)

// ErrNoArtworks 表示接口按当前条件查不到作品。
var ErrNoArtworks = errors.New("no artworks found")

const (
	requestTimeout   = 10 * time.Second
	maxResponseBytes = 8 << 20
)

type Client struct {
	base string
	http *http.Client
}

// New 的 base 留空时禁用客户端。
func New(base string) *Client {
	return &Client{
		base: strings.TrimRight(strings.TrimSpace(base), "/"),
		http: &http.Client{Timeout: requestTimeout},
	}
}

func (c *Client) Enabled() bool {
	return c != nil && c.base != ""
}

type ArtworkQuery struct {
	R18           int
	Limit         int
	Page          int
	Keyword       string
	ArtistID      string
	Tag           string
	Hybrid        bool
	SemanticRatio float64
}

func (c *Client) RandomArtworks(ctx context.Context, q ArtworkQuery) ([]ouid.OUID, error) {
	values := q.values()
	if q.Limit > 0 {
		values.Set("limit", strconv.Itoa(q.Limit))
	}
	return c.get(ctx, "/artwork/random", values)
}

func (c *Client) ListArtworks(ctx context.Context, q ArtworkQuery) ([]ouid.OUID, error) {
	values := q.values()
	if q.Limit > 0 {
		values.Set("page_size", strconv.Itoa(q.Limit))
	}
	if q.Page > 0 {
		values.Set("page", strconv.Itoa(q.Page))
	}
	if q.Hybrid {
		values.Set("hybrid", "true")
		values.Set("semantic_ratio", strconv.FormatFloat(q.SemanticRatio, 'f', -1, 64))
	}
	return c.get(ctx, "/artwork/list", values)
}

func (q ArtworkQuery) values() url.Values {
	values := url.Values{}
	values.Set("r18", strconv.Itoa(q.R18))
	if q.Keyword != "" {
		values.Set("keyword", q.Keyword)
	}
	if q.ArtistID != "" {
		values.Set("artist_id", q.ArtistID)
	}
	if q.Tag != "" {
		values.Set("tag", q.Tag)
	}
	return values
}

type envelope struct {
	Message string `json:"message"`
	Data    []struct {
		ID string `json:"id"`
	} `json:"data"`
}

func (c *Client) get(ctx context.Context, path string, values url.Values) ([]ouid.OUID, error) {
	if !c.Enabled() {
		return nil, errors.New("rest client is not configured")
	}
	target := c.base + path
	if encoded := values.Encode(); encoded != "" {
		target += "?" + encoded
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, target, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("query %s failed: %w", path, err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, maxResponseBytes))
	if err != nil {
		return nil, fmt.Errorf("read %s response failed: %w", path, err)
	}
	if resp.StatusCode == http.StatusNotFound {
		return nil, ErrNoArtworks
	}
	var parsed envelope
	if err := json.Unmarshal(body, &parsed); err != nil {
		return nil, fmt.Errorf("decode %s response failed: status %d: %w", path, resp.StatusCode, err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("query %s failed: status %d: %s", path, resp.StatusCode, parsed.Message)
	}
	if len(parsed.Data) == 0 {
		return nil, ErrNoArtworks
	}
	ids := make([]ouid.OUID, 0, len(parsed.Data))
	for _, item := range parsed.Data {
		id, err := ouid.FromObjectIDHex(item.ID)
		if err != nil {
			return nil, fmt.Errorf("decode %s artwork id %q: %w", path, item.ID, err)
		}
		ids = append(ids, id)
	}
	return ids, nil
}
