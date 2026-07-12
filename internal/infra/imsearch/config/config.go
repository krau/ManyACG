package config

import "runtime"

const (
	QuantizerFile = "quantizer.bin"
	InvlistsFile  = "invlists.bin"
)

type SearchOptions struct {
	Distance   uint32
	Count      int
	K          int
	NProbe     int
	Threads    int
	MinMatches int     // min accepted descriptor matches per image (0=default)
	MinScore   float32 // min Wilson score×100 (0=default)
}

func DefaultSearchOptions() SearchOptions {
	return SearchOptions{
		Distance:   64,
		Count:      10,
		K:          3,
		NProbe:     16,
		Threads:    runtime.NumCPU(),
		MinMatches: 8,
		MinScore:   25,
	}
}
