package shared

//go:generate go-enum --values --names --nocase

// SourceType
/*
ENUM(
pixiv
twitter
bilibili
danbooru
kemono
yandere
nhentai
)
*/
type SourceType string
