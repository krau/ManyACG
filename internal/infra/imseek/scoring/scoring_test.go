package scoring

import (
	"math"
	"testing"
)

func TestWilsonEmpty(t *testing.T) {
	if Wilson(nil) != 0 {
		t.Fatal("empty should score 0")
	}
}

func TestWilsonMonotonic(t *testing.T) {
	// More matches of the same similarity should score higher (tighter lower
	// bound).
	few := Wilson([]float32{0.9, 0.9})
	many := Wilson([]float32{0.9, 0.9, 0.9, 0.9, 0.9, 0.9, 0.9, 0.9})
	if many <= few {
		t.Fatalf("more matches should score higher: few=%f many=%f", few, many)
	}
}

func TestWilsonHigherSimilarityScoresHigher(t *testing.T) {
	low := Wilson([]float32{0.5, 0.5, 0.5, 0.5})
	high := Wilson([]float32{0.95, 0.95, 0.95, 0.95})
	if high <= low {
		t.Fatalf("higher similarity should score higher: low=%f high=%f", low, high)
	}
}

func TestWilsonBounds(t *testing.T) {
	// Score should stay within [0,1] for valid similarity inputs.
	s := Wilson([]float32{1, 1, 1, 1, 1})
	if s < 0 || s > 1.0001 {
		t.Fatalf("score %f out of range", s)
	}
	if math.IsNaN(float64(s)) {
		t.Fatal("score is NaN")
	}
}
