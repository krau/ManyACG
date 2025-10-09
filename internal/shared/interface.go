package shared

type PictureLike interface {
	IsHide() bool
	GetIndex() uint
	GetTelegramInfo() TelegramInfo
	GetOriginal() string
	GetThumbnail() string
	GetSize() (width, height uint)
	GetStorageInfo() StorageInfo
}

type ArtworkLike interface {
	GetSourceURL() string
	GetTitle() string
	GetR18() bool
	GetArtistName() string
	GetDescription() string
	GetTags() []string
	GetPictures() []PictureLike
}
