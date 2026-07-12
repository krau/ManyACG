package scoring

import "math"

const z = 2.326

func Wilson(scores []float32) float32 {
	n := len(scores)
	if n == 0 {
		return 0
	}
	count := float64(n)
	var sum float64
	for _, s := range scores {
		sum += float64(s)
	}
	mean := sum / count

	var variance float64
	for _, s := range scores {
		d := mean - float64(s)
		variance += d * d
	}
	variance /= count

	z2 := z * z
	numerator := mean + z2/(2*count) - (z/(2*count))*math.Sqrt(4*count*variance+z2)
	denominator := 1 + z2/count
	return float32(numerator / denominator)
}
