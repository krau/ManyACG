package shared

type ArtworkLike interface {
	GetID() string
	GetSourceURL() string
	GetTitle() string
	GetR18() bool
	GetArtist() ArtistLike
	GetDescription() string
	GetTags() []string
	MediasCount() int
	GetType() SourceType
	FirstMedia() MediaLike
	GetPictures() []PictureLike
	GetUgoiraMetas() []UgoiraMetaLike
	GetVideos() []VideoLike
}

type ArtistLike interface {
	GetName() string
	GetUserName() string
	GetUID() string
}

type PictureLike interface {
	IsHide() bool
	GetIndex() uint
	GetTelegramInfo() TelegramInfo
	GetOriginal() string
	GetThumbnail() string
	GetSize() (width, height uint)
	GetStorageInfo() StorageInfo
}

type VideoLike interface {
	GetIndex() uint
	GetURL() string
	GetWidth() uint
	GetHeight() uint
	GetDuration() uint
	GetPoster() string
	GetMimeType() string
	GetOriginalStorage() StorageDetail
	GetTelegramInfo() TelegramInfo
}

type UgoiraMetaLike interface {
	GetIndex() uint
	GetUgoiraMetaData() UgoiraMetaData
	GetOriginalStorage() StorageDetail
	GetTelegramInfo() TelegramInfo
}

type MediaLike interface {
	GetIndex() uint
	GetTelegramInfo() TelegramInfo
}
