package types

type ArtworkStatus string

const (
	ArtworkStatusCached  ArtworkStatus = "cached"
	ArtworkStatusPosting ArtworkStatus = "posting"
	ArtworkStatusPosted  ArtworkStatus = "posted"
)
