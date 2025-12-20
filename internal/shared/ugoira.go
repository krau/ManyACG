package shared

type UgoiraMetaData struct {
	// PosterOriginal is the original poster image URL.
	PosterOriginal string        `json:"poster_original"`
	PosterThumb    string        `json:"poster_thumb"`
	// OriginalZip is the original ugoira zip URL.
	//
	// The zip should be pixiv's ugoira zip style
	OriginalZip    string        `json:"original_zip"`
	ThumbZip       string        `json:"thumb_zip"`
	MimeType       string        `json:"mime_type"`
	Width          int           `json:"width"`
	Height         int           `json:"height"`
	Frames         []UgoiraFrame `json:"frames"`
}

var ZeroUgoiraMetaData = UgoiraMetaData{}

type UgoiraFrame struct {
	File  string `json:"file"`
	Delay int    `json:"delay"`
}
