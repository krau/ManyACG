package shared

type TelegramInfo struct {
	// PhotoFileID    string `json:"photo_file_id"`
	// DocumentFileID string `json:"document_file_id"`
	// MessageID      int    `json:"message_id"`
	// MediaGroupID   string `json:"media_group_id"`

	// {bot_id:{"photo": file_id, "document": file_id}}
	FileIDs  map[int64]map[TelegramMediaType]string `json:"file_ids"` // (file_id is bot specific)
	Messages map[int64]TelegramMessage              `json:"messages"` // key: chat id
}

type TelegramMessage struct {
	MessageID    int    `json:"message_id"`
	MediaGroupID string `json:"media_group_id,omitempty"`
}

func (t TelegramInfo) MessageID(chatID int64) int {
	if t.Messages == nil {
		return 0
	}
	if msg, ok := t.Messages[chatID]; ok {
		return msg.MessageID
	}
	return 0
}

func (t TelegramInfo) FileID(botID int64, mediaType TelegramMediaType) string {
	if t.FileIDs == nil {
		return ""
	}
	if botFiles, ok := t.FileIDs[botID]; ok {
		if fileID, ok := botFiles[mediaType]; ok {
			return fileID
		}
	}
	return ""
}

func (t TelegramInfo) PhotoFileID(botID int64) string {
	return t.FileID(botID, TelegramMediaTypePhoto)
}

func (t TelegramInfo) DocumentFileID(botID int64) string {
	return t.FileID(botID, TelegramMediaTypeDocument)
}

func (t TelegramInfo) VideoFileID(botID int64) string {
	return t.FileID(botID, TelegramMediaTypeVideo)
}

func (t TelegramInfo) IsZero() bool {
	if len(t.FileIDs) != 0 {
		return false
	}
	if len(t.Messages) != 0 {
		return false
	}
	return true
}

func (t *TelegramInfo) SetMessage(chatID int64, messageID int, mediaGroupID string) {
	if t.Messages == nil {
		t.Messages = make(map[int64]TelegramMessage)
	}
	t.Messages[chatID] = TelegramMessage{
		MessageID:    messageID,
		MediaGroupID: mediaGroupID,
	}
}

func (t *TelegramInfo) SetFileID(botID int64, mediaType TelegramMediaType, fileID string) {
	if t.FileIDs == nil {
		t.FileIDs = make(map[int64]map[TelegramMediaType]string)
	}
	if _, ok := t.FileIDs[botID]; !ok {
		t.FileIDs[botID] = make(map[TelegramMediaType]string)
	}
	t.FileIDs[botID][mediaType] = fileID
}

func (t *TelegramInfo) ClearFileIDs() {
	t.FileIDs = make(map[int64]map[TelegramMediaType]string)
}

func (t *TelegramInfo) MergeFrom(other *TelegramInfo) {
	if other == nil {
		return
	}
	if t.FileIDs == nil {
		t.FileIDs = make(map[int64]map[TelegramMediaType]string)
	}
	for botID, botFiles := range other.FileIDs {
		if _, ok := t.FileIDs[botID]; !ok {
			t.FileIDs[botID] = make(map[TelegramMediaType]string)
		}
		for mediaType, fileID := range botFiles {
			t.FileIDs[botID][mediaType] = fileID
		}
	}
	if t.Messages == nil {
		t.Messages = make(map[int64]TelegramMessage)
	}
	for chatID, msg := range other.Messages {
		t.Messages[chatID] = msg
	}
}
