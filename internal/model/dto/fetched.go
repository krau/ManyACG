package dto

import (
	"github.com/krau/ManyACG/internal/shared"
)

// FetchedArtwork 只会传递给 service 层, 其他层获取到的应该是 CachedArtwork 或 Artwork 实体.
// var _ shared.ArtworkLike = (*FetchedArtwork)(nil)
// var _ shared.PictureLike = (*FetchedPicture)(nil)
// var _ shared.UgoiraMetaLike = (*FetchedUgoiraMeta)(nil)
// var _ shared.VideoLike = (*FetchedVideo)(nil)
// var _ shared.ArtistLike = (*FetchedArtist)(nil)

type FetchedArtwork struct {
	Title       string            `json:"title"`
	Description string            `json:"description"`
	R18         bool              `json:"r18"`
	SourceType  shared.SourceType `json:"source_type"`
	SourceURL   string            `json:"source_url"`

	Artist      *FetchedArtist       `json:"artist"`
	Tags        []string             `json:"tags"`
	Pictures    []*FetchedPicture    `json:"pictures"`
	UgoiraMetas []*FetchedUgoiraMeta `json:"ugoira_metas,omitempty"`
	Videos      []*FetchedVideo      `json:"videos,omitempty"`
}

// // GetUgoiraMetas implements [shared.ArtworkLike].
// func (f *FetchedArtwork) GetUgoiraMetas() []shared.UgoiraMetaLike {
// 	var metas []shared.UgoiraMetaLike
// 	for _, m := range f.UgoiraMetas {
// 		metas = append(metas, m)
// 	}
// 	return metas
// }

// // GetVideos implements [shared.ArtworkLike].
// func (f *FetchedArtwork) GetVideos() []shared.VideoLike {
// 	var videos []shared.VideoLike
// 	for _, v := range f.Videos {
// 		videos = append(videos, v)
// 	}
// 	return videos
// }

// // MediasCount implements [shared.ArtworkLike].
// func (f *FetchedArtwork) MediasCount() int {
// 	count := len(f.Pictures)
// 	count += len(f.Videos)
// 	count += len(f.UgoiraMetas)
// 	return count
// }

// // GetType implements shared.ArtworkLike.
// func (f *FetchedArtwork) GetType() shared.SourceType {
// 	return f.SourceType
// }

type FetchedArtist struct {
	Name     string            `json:"name"`
	Type     shared.SourceType `json:"type"`
	UID      string            `json:"uid"`
	Username string            `json:"username"`
}

// // GetName implements shared.ArtistLike.
// func (f *FetchedArtist) GetName() string {
// 	return f.Name
// }

// // GetUID implements shared.ArtistLike.
// func (f *FetchedArtist) GetUID() string {
// 	return f.UID
// }

// // GetUserName implements shared.ArtistLike.
// func (f *FetchedArtist) GetUserName() string {
// 	return f.Username
// }

type FetchedPicture struct {
	Index     uint   `json:"index"`
	Thumbnail string `json:"thumbnail"`
	Original  string `json:"original"`

	Width  uint `json:"width"`
	Height uint `json:"height"`
}

// // IsHide implements shared.PictureLike.
// func (f *FetchedPicture) IsHide() bool {
// 	return false
// }

// func (f *FetchedArtwork) GetID() string {
// 	return ""
// }

// // GetIndex implements shared.PictureLike.
// func (f *FetchedPicture) GetIndex() uint {
// 	return f.Index
// }

// // GetArtist implements shared.ArtworkLike.
// func (f *FetchedArtwork) GetArtist() shared.ArtistLike {
// 	return f.Artist
// }

// // GetDescription implements shared.ArtworkLike.
// func (f *FetchedArtwork) GetDescription() string {
// 	return f.Description
// }

// // GetPictures implements shared.ArtworkLike.
// func (f *FetchedArtwork) GetPictures() []shared.PictureLike {
// 	var pictures []shared.PictureLike
// 	for _, pic := range f.Pictures {
// 		pictures = append(pictures, pic)
// 	}
// 	return pictures
// }

// // GetR18 implements shared.ArtworkLike.
// func (f *FetchedArtwork) GetR18() bool {
// 	return f.R18
// }

// // GetSourceURL implements shared.ArtworkLike.
// func (f *FetchedArtwork) GetSourceURL() string {
// 	return f.SourceURL
// }

// // GetTags implements shared.ArtworkLike.
// func (f *FetchedArtwork) GetTags() []string {
// 	return f.Tags
// }

// // GetTitle implements shared.ArtworkLike.
// func (f *FetchedArtwork) GetTitle() string {
// 	return f.Title
// }

// // GetOriginal implements shared.PictureLike.
// func (f *FetchedPicture) GetOriginal() string {
// 	return f.Original
// }

// // GetSize implements shared.PictureLike.
// func (f *FetchedPicture) GetSize() (width uint, height uint) {
// 	return f.Width, f.Height
// }

// // GetStorageInfo implements shared.PictureLike.
// func (f *FetchedPicture) GetStorageInfo() shared.StorageInfo {
// 	return shared.StorageInfo{}
// }

// // GetTelegramInfo implements shared.PictureLike.
// func (f *FetchedPicture) GetTelegramInfo() shared.TelegramInfo {
// 	return shared.TelegramInfo{}
// }

// // GetThumbnail implements shared.PictureLike.
// func (f *FetchedPicture) GetThumbnail() string {
// 	return f.Thumbnail
// }

type FetchedUgoiraMeta struct {
	Index uint                  `json:"index"`
	Data  shared.UgoiraMetaData `json:"data"`
}

// // GetIndex implements [shared.UgoiraMetaLike].
// func (f *FetchedUgoiraMeta) GetIndex() uint {
// 	return f.Index
// }

// // GetOriginalStorage implements [shared.UgoiraMetaLike].
// func (f *FetchedUgoiraMeta) GetOriginalStorage() shared.StorageDetail {
// 	return shared.ZeroStorageDetail
// }

// // GetTelegramInfo implements [shared.UgoiraMetaLike].
// func (f *FetchedUgoiraMeta) GetTelegramInfo() shared.TelegramInfo {
// 	return shared.TelegramInfo{}
// }

// // GetUgoiraMetaData implements [shared.UgoiraMetaLike].
// func (f *FetchedUgoiraMeta) GetUgoiraMetaData() shared.UgoiraMetaData {
// 	return f.Data
// }

type FetchedVideo struct {
	Index    uint   `json:"index"`
	URL      string `json:"url"`
	Width    uint   `json:"width"`
	Height   uint   `json:"height"`
	Duration uint   `json:"duration"` // in milliseconds
	Poster   string `json:"poster"`
	MimeType string `json:"mime_type"`
}

// // GetDuration implements [shared.VideoLike].
// func (f *FetchedVideo) GetDuration() uint {
// 	return f.Duration
// }

// // GetHeight implements [shared.VideoLike].
// func (f *FetchedVideo) GetHeight() uint {
// 	return f.Height
// }

// // GetIndex implements [shared.VideoLike].
// func (f *FetchedVideo) GetIndex() uint {
// 	return f.Index
// }

// // GetMimeType implements [shared.VideoLike].
// func (f *FetchedVideo) GetMimeType() string {
// 	return f.MimeType
// }

// // GetPoster implements [shared.VideoLike].
// func (f *FetchedVideo) GetPoster() string {
// 	return f.Poster
// }

// // GetStorageInfo implements [shared.VideoLike].
// func (f *FetchedVideo) GetStorageInfo() shared.StorageInfo {
// 	return shared.StorageInfo{}
// }

// // GetTelegramInfo implements [shared.VideoLike].
// func (f *FetchedVideo) GetTelegramInfo() shared.TelegramInfo {
// 	return shared.TelegramInfo{}
// }

// // GetURL implements [shared.VideoLike].
// func (f *FetchedVideo) GetURL() string {
// 	return f.URL
// }

// // GetWidth implements [shared.VideoLike].
// func (f *FetchedVideo) GetWidth() uint {
// 	return f.Width
// }
