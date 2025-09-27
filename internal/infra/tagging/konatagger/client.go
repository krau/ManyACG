package konatagger

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/imroc/req/v3"
)

type taggerClient struct {
	Client  *req.Client
	host    string
	token   string
	timeout time.Duration
}

func (c *taggerClient) Health() (string, error) {
	var health struct {
		Status string `json:"status"`
	}
	resp, err := c.Client.R().Get("/health")
	if err != nil {
		return "", err
	}
	if err := json.Unmarshal(resp.Bytes(), &health); err != nil {
		return "", err
	}
	return health.Status, nil
}

type taggerPredictResponse struct {
	PredictedTags []string           `json:"predicted_tags"`
	Scores        map[string]float64 `json:"scores"`
}

func (c *taggerClient) Predict(ctx context.Context, file []byte) (*taggerPredictResponse, error) {
	resp, err := c.Client.R().SetContext(ctx).SetFileBytes("file", "image", file).Post("/predict")
	if err != nil {
		return nil, err
	}
	if resp.IsErrorState() {
		return nil, fmt.Errorf("tagger predict failed: %s", resp.Status)
	}
	var predict taggerPredictResponse
	if err := json.Unmarshal(resp.Bytes(), &predict); err != nil {
		return nil, err
	}
	return &predict, nil
}
func NewKonatagger(host, token string, timeout time.Duration) (*taggerClient, error) {
	client := req.C().
		SetCommonBearerAuthToken(token).
		SetBaseURL(host).
		SetTimeout(timeout * time.Second).
		SetUserAgent("ManyACG")
	tagerC := &taggerClient{
		Client:  client,
		host:    host,
		token:   token,
		timeout: timeout * time.Second,
	}
	if _, err := tagerC.Health(); err != nil {
		return nil, fmt.Errorf("tagger health check failed: %w", err)
	}
	return tagerC, nil
}
