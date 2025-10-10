package shared

type UgoiraMetaData struct {
	PosterOriginal string        `json:"poster_original"`
	PosterThumb    string        `json:"poster_thumb"`
	OriginalZip    string        `json:"original_zip"`
	ThumbZip       string        `json:"thumb_zip"`
	MimeType       string        `json:"mime_type"`
	Width          int           `json:"width"`
	Height         int           `json:"height"`
	Frames         []UgoiraFrame `json:"frames"`
}

type UgoiraFrame struct {
	File  string `json:"file"`
	Delay int    `json:"delay"`
}
