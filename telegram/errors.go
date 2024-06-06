package telegram

import (
	"errors"
)

var (
	ErrNoPhotoInMessage = errors.New("message has no photo")
	ErrFileTooLarge     = errors.New("file too large (>20MB)")
)
