package shared

// type PictureInfo struct {
// 	Index     uint
// 	Thumbnail string
// 	Original  string
// 	Width     uint
// 	Height    uint
// 	Phash     string // phash
// 	ThumbHash string // thumbhash

// 	TelegramInfo *TelegramInfo
// 	StorageInfo  *StorageInfo
// }

var ZeroTelegramInfo = TelegramInfo{}

type TelegramInfo struct {
	PhotoFileID    string `json:"photo_file_id"`
	DocumentFileID string `json:"document_file_id"`
	MessageID      int    `json:"message_id"`
	MediaGroupID   string `json:"media_group_id"`
}

var ZeroStorageInfo = StorageInfo{}

type StorageInfo struct {
	Original *StorageDetail `json:"original"`
	Regular  *StorageDetail `json:"regular"`
	Thumb    *StorageDetail `json:"thumb"`
}

type StorageDetail struct {
	Type StorageType `json:"type"`
	Path string      `json:"path"`
}
