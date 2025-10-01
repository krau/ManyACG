package tagging

import (
	"context"
	"errors"
	"io"
	"sort"
	"sync"
	"time"

	"github.com/krau/ManyACG/internal/infra/config/runtimecfg"
	"github.com/krau/ManyACG/internal/infra/tagging/konatagger"
	"github.com/krau/ManyACG/pkg/log"
)

type PredictResult struct {
	PredictedTags []string           `json:"predicted_tags"`
	Scores        map[string]float64 `json:"scores"`
}

type Tagger interface {
	Predict(ctx context.Context, file io.Reader) (map[string]float64, error)
}

var (
	defaultTagger Tagger
	defaultOnce   sync.Once
)

type noopTagger struct{}

func (n *noopTagger) Predict(ctx context.Context, file io.Reader) (map[string]float64, error) {
	return nil, errors.New("tagging is disabled")
}

func Default() Tagger {
	defaultOnce.Do(func() {
		cfg := runtimecfg.Get().Tagging
		if !cfg.Enable {
			defaultTagger = &noopTagger{}
			return
		}
		tagger, err := konatagger.New(cfg.Konatagger.Host, cfg.Konatagger.Token, time.Duration(cfg.Konatagger.Timeout)*time.Second)
		if err != nil {
			log.Fatal("failed to create konatagger client", "err", err)
		}
		defaultTagger = tagger
	})
	return defaultTagger
}

func Predict(ctx context.Context, file io.Reader) (*PredictResult, error) {
	scores, err := Default().Predict(ctx, file)
	if err != nil {
		return nil, err
	}
	keys := make([]string, 0, len(scores))
	for k := range scores {
		keys = append(keys, k)
	}
	sort.SliceStable(keys, func(i, j int) bool {
		return scores[keys[i]] > scores[keys[j]]
	})
	return &PredictResult{
		PredictedTags: keys,
		Scores:        scores,
	}, nil
}
