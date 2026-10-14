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
	"os"
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

const (
	artworkFetchPath    = "/internal/artwork/fetch"
	artworkDownloadPath = "/internal/artwork/download"
)

func New(cfg runtimecfg.RemoteOperationConfig, token string) (*Client, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	if len(cfg.Endpoints) == 0 {
		return nil, errors.New("remote endpoints must contain at least one node")
	}
	if strings.TrimSpace(token) == "" {
		return nil, errors.New("remote token is required")
	}
	endpoints := make([]string, len(cfg.Endpoints))
	for i, endpoint := range cfg.Endpoints {
		endpoints[i] = strings.TrimRight(endpoint, "/")
	}
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.Proxy = nil
	return &Client{
		endpoints: endpoints,
		token:     token,
		http: &http.Client{
			Transport: transport,
			Timeout:   time.Duration(cfg.Timeout) * time.Second,
			CheckRedirect: func(_ *http.Request, _ []*http.Request) error {
				return http.ErrUseLastResponse
			},
		},
	}, nil
}

func (c *Client) newRequest(ctx context.Context, path string, payload []byte) (*http.Request, error) {
	endpoint := c.endpoints[(c.next.Add(1)-1)%uint64(len(c.endpoints))] + path
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(payload))
	if err != nil {
		return nil, errors.New("creating reader request failed")
	}
	req.GetBody = nil
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Content-Type", "application/json")
	return req, nil
}

func (c *Client) GetArtworkInfo(ctx context.Context, sourceURL string) (*dto.FetchedArtwork, error) {
	body, err := json.Marshal(struct {
		URL string `json:"url"`
	}{URL: sourceURL})
	if err != nil {
		return nil, errors.New("encoding reader artwork request failed")
	}
	req, err := c.newRequest(ctx, artworkFetchPath, body)
	if err != nil {
		return nil, err
	}
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

func (c *Client) FetchToFile(ctx context.Context, url, path string) error {
	payload, err := json.Marshal(struct {
		URL string `json:"url"`
	}{URL: url})
	if err != nil {
		return errors.New("encoding reader download request failed")
	}
	req, err := c.newRequest(ctx, artworkDownloadPath, payload)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/octet-stream")
	response, err := c.http.Do(req)
	if err != nil {
		if transferErr := transferError(ctx, err); transferErr != nil {
			return transferErr
		}
		return errors.New("reader download request failed")
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("reader download request returned HTTP %d", response.StatusCode)
	}
	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("create reader download file: %w", err)
	}
	if _, err := io.Copy(file, response.Body); err != nil {
		file.Close()
		if transferErr := transferError(ctx, err); transferErr != nil {
			return transferErr
		}
		return fmt.Errorf("reader download response could not be read: %w", err)
	}
	if err := file.Close(); err != nil {
		return fmt.Errorf("close reader download file: %w", err)
	}
	return nil
}

func transferError(ctx context.Context, err error) error {
	if ctxErr := ctx.Err(); ctxErr != nil {
		return fmt.Errorf("reader request canceled: %w", ctxErr)
	}
	var netErr net.Error
	if errors.Is(err, context.DeadlineExceeded) || (errors.As(err, &netErr) && netErr.Timeout()) {
		return fmt.Errorf("reader request timed out: %w", context.DeadlineExceeded)
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
