package restful

type SendArtworkInfoRequest struct {
	SourceURL     string `json:"source_url" binding:"required"`
	ChatID        int64  `json:"chat_id"`
	AppendCaption string `json:"append_caption"`
}
