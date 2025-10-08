package tagging

import (
	"context"
	"errors"
	"io"
	"sync"
	"time"

	"github.com/krau/ManyACG/internal/infra/config/runtimecfg"
	"github.com/krau/ManyACG/internal/infra/tagging/konatagger"
	"github.com/krau/ManyACG/internal/infra/tagging/mocktagger"
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
	ErrNotEnabled = errors.New("tagging engine is not enabled")
)

type noopTagger struct{}

func (n *noopTagger) Predict(ctx context.Context, file io.Reader) (map[string]float64, error) {
	return nil, ErrNotEnabled
}

func Default() Tagger {
	defaultOnce.Do(func() {
		cfg := runtimecfg.Get().Tagging
		if !cfg.Enable {
			defaultTagger = &noopTagger{}
			return
		}
		switch cfg.Engine {
		case "konatagger":
			tagger, err := konatagger.New(cfg.Konatagger.Host, cfg.Konatagger.Token, time.Duration(cfg.Konatagger.Timeout)*time.Second)
			if err != nil {
				log.Fatal("failed to create konatagger client", "err", err)
			}
			defaultTagger = tagger
		case "mock":
			log.Warn("using mock tagger, this is only for testing purposes")
			defaultTagger = mocktagger.New()
		default:
			log.Fatal("unknown tagging engine: %s", cfg.Engine)
		}
	})
	return defaultTagger
}

func Enabled() bool {
	return runtimecfg.Get().Tagging.Enable
}