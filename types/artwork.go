package types

type Artwork struct {
	Title       string
	Description string
	R18         bool
	Source      ArtworkSource
	Artist      Artist
	Tags        []ArtworkTag
	Pictures    []*Picture
}

type ArtworkSource struct {
	Type SourceType
	URL  string
}

type Artist struct {
	Name     string
	Type     SourceType
	UID      int
	Username string
}

type ArtworkTag struct {
	Name string
}

type Picture struct {
	Index     uint   // 图片在作品中的顺序
	Thumbnail string // 缩略图 URL
	Original  string // 原图 URL

	Width     uint
	Height    uint
	Hash      string
	BlurScore float64

	Format       string
	TelegramInfo *TelegramInfo
}

type TelegramInfo struct {
	PhotoFileID    string
	DocumentFileID string
	MessageID      int
	MediaGroupID   string
}
