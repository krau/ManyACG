package types

type SourceType string

const (
	SourceTypePixiv    SourceType = "pixiv"
	SourceTypeTwitter  SourceType = "twitter"
	SourceTypeBilibili SourceType = "bilibili"
	SourceTypeDanbooru SourceType = "danbooru"
	SourceTypeKemono   SourceType = "kemono"
)

var SourceTypes []SourceType = []SourceType{
	SourceTypePixiv,
	SourceTypeTwitter,
	SourceTypeBilibili,
	SourceTypeDanbooru,
	SourceTypeKemono,
}
