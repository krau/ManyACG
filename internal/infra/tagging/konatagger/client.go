package konatagger

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/imroc/req/v3"
	"github.com/krau/ManyACG/config"
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

var TaggerClient *taggerClient

func initTaggerClient() {
	if config.Cfg.Tagger.Host == "" || config.Cfg.Tagger.Token == "" {
		Logger.Fatalf("Tagger configuration is incomplete")
		os.Exit(1)
	}
	client := req.C().
		SetCommonBearerAuthToken(config.Cfg.Tagger.Token).
		SetBaseURL(config.Cfg.Tagger.Host).
		SetTimeout(time.Duration(config.Cfg.Tagger.Timeout) * time.Second).
		SetUserAgent("ManyACG/" + Version)
	TaggerClient = &taggerClient{
		Client:  client,
		host:    config.Cfg.Tagger.Host,
		token:   config.Cfg.Tagger.Token,
		timeout: time.Duration(config.Cfg.Tagger.Timeout) * time.Second,
	}
	if status, err := TaggerClient.Health(); err != nil {
		Logger.Fatalf("Tagger health check failed: %s", err)
		os.Exit(1)
	} else {
		Logger.Infof("Tagger health check: %s", status)
	}
}
