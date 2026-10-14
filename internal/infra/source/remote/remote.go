package remote

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync/atomic"
	"time"

	"github.com/krau/ManyACG/internal/infra/config/runtimecfg"
	"github.com/krau/ManyACG/internal/model/dto"
)

type Client struct {
	endpoints []string
	token     string
	http      *http.Client
	next      atomic.Uint64
}

func New(cfg runtimecfg.SourceRemoteConfig) (*Client, error) {
	if err := cfg.Valid("", false); err != nil {
		return nil, err
	}
	if len(cfg.Endpoints) == 0 {
		return nil, errors.New("source.remote.endpoints must contain at least one reader")
	}
	endpoints := make([]string, len(cfg.Endpoints))
	for i, endpoint := range cfg.Endpoints {
		u, err := url.Parse(endpoint)
		if err != nil {
			return nil, fmt.Errorf("parse source.remote.endpoints[%d]: %w", i, err)
		}
		endpoints[i] = strings.TrimRight(u.String(), "/") + "/internal/artwork/fetch"
	}
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.Proxy = nil
	return &Client{
		endpoints: endpoints,
		token:     cfg.Token,
		http: &http.Client{
			Transport: transport,
			Timeout:   time.Duration(cfg.Timeout) * time.Second,
			CheckRedirect: func(_ *http.Request, _ []*http.Request) error {
				return http.ErrUseLastResponse
			},
		},
	}, nil
}

func (c *Client) GetArtworkInfo(ctx context.Context, sourceURL string) (*dto.FetchedArtwork, error) {
	body, err := json.Marshal(struct {
		URL string `json:"url"`
	}{URL: sourceURL})
	if err != nil {
		return nil, errors.New("encoding reader artwork request failed")
	}
	endpoint := c.endpoints[(c.next.Add(1)-1)%uint64(len(c.endpoints))]
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, errors.New("creating reader artwork request failed")
	}
	req.GetBody = nil
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	response, err := c.http.Do(req)
	if err != nil {
		if transferErr := transferError(ctx, err); transferErr != nil {
			return nil, transferErr
		}
		return nil, errors.New("reader artwork request failed")
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("reader artwork request returned HTTP %d", response.StatusCode)
	}
	var artwork *dto.FetchedArtwork
	decoder := json.NewDecoder(response.Body)
	if err := decoder.Decode(&artwork); err != nil {
		if transferErr := transferError(ctx, err); transferErr != nil {
			return nil, transferErr
		}
		return nil, errors.New("reader returned invalid artwork JSON")
	}
	var extra json.RawMessage
	if err := decoder.Decode(&extra); err != io.EOF {
		if transferErr := transferError(ctx, err); transferErr != nil {
			return nil, transferErr
		}
		return nil, errors.New("reader returned trailing artwork response data")
	}
	if err := validateArtwork(artwork); err != nil {
		return nil, err
	}
	return artwork, nil
}

func transferError(ctx context.Context, err error) error {
	if ctxErr := ctx.Err(); ctxErr != nil {
		return fmt.Errorf("reader artwork request canceled: %w", ctxErr)
	}
	var netErr net.Error
	if errors.Is(err, context.DeadlineExceeded) || (errors.As(err, &netErr) && netErr.Timeout()) {
		return fmt.Errorf("reader artwork request timed out: %w", context.DeadlineExceeded)
	}
	return nil
}

func validateArtwork(artwork *dto.FetchedArtwork) error {
	if artwork == nil {
		return errors.New("reader returned a null artwork")
	}
	if strings.TrimSpace(artwork.SourceURL) == "" {
		return errors.New("reader artwork is missing source_url")
	}
	if artwork.Artist == nil {
		return errors.New("reader artwork is missing artist")
	}
	for i, picture := range artwork.Pictures {
		if picture == nil {
			return fmt.Errorf("reader artwork pictures[%d] is null", i)
		}
	}
	for i, meta := range artwork.UgoiraMetas {
		if meta == nil {
			return fmt.Errorf("reader artwork ugoira_metas[%d] is null", i)
		}
	}
	for i, video := range artwork.Videos {
		if video == nil {
			return fmt.Errorf("reader artwork videos[%d] is null", i)
		}
	}
	return nil
}

func (c *Client) Close() error {
	c.http.CloseIdleConnections()
	return nil
}
