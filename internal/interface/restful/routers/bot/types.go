package bot

type SendArtworkInfoRequest struct {
	SourceURL     string `json:"source_url" binding:"required"`
	ChatID        int64  `json:"chat_id" binding:"required"`
	AppendCaption string `json:"append_caption"`
}
