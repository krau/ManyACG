package mocktagger

import (
	"context"
	"io"

	"github.com/duke-git/lancet/v2/random"
)

type mockTagger struct {
	randTags []string
}

func (c *mockTagger) Predict(ctx context.Context, file io.Reader) (map[string]float64, error) {
	scores := make(map[string]float64)
	for range 5 {
		tag := random.RandFromGivenSlice(c.randTags)
		scores[tag] = random.RandFloat(0.5, 1.0, 2)
	}
	return scores, nil
}

func New() *mockTagger {
	return &mockTagger{
		randTags: random.RandStringSlice(random.Letters, 10, 8),
	}
}
