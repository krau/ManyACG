package model

type Artwork struct {
	Title       string
	Description string
	SourceType  string
	SourceURL   string
	Author      string
	Tags        []string
	R18         bool
	Pictures    []*Picture
}

type Picture struct {
	DirectURL string
	Width     uint
	Height    uint
	Hash      string
	BlurScore float64
	Format    string
}
