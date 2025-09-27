package telegram

import (
	"errors"
	"fmt"
)

var (
	ErrReadFile                 = errors.New("failed to read file")
	ErrFailedSendDocument       = errors.New("failed to send document")
	ErrFailedMarshalFileMessage = errors.New("failed to marshal file message")
)

type fileMessage struct {
	ChatID     int64  `json:"chat_id"`
	MessaageID int    `json:"message_id"`
	FileID     string `json:"file_id"`
}

func (f *fileMessage) String() string {
	return fmt.Sprintf("chat_id: %d, message_id: %d, file_id: %s", f.ChatID, f.MessaageID, f.FileID)
}
