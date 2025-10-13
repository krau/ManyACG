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
	GetID() string
	GetSourceURL() string
	GetTitle() string
	GetR18() bool
	GetArtist() ArtistLike
	GetDescription() string
	GetTags() []string
	GetPictures() []PictureLike
	GetType() SourceType
}

type ArtistLike interface {
	GetName() string
	GetUserName() string
	GetUID() string
}

type UgoiraArtworkLike interface {
	ArtworkLike
	GetUgoiraMetas() []UgoiraMetaLike
}

type UgoiraMetaLike interface {
	GetIndex() uint
	GetUgoiraMetaData() UgoiraMetaData
	GetOriginalStorage() StorageDetail
	GetTelegramInfo() TelegramInfo
}
