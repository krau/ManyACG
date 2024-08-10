package utils

import (
	"errors"
)

var (
	ErrNoPhotoInMessage = errors.New("message has no photo")
	ErrFileTooLarge     = errors.New("file too large (>20MB)")
	ErrNoAvailableFile  = errors.New("no available file")
	ErrNilBot           = errors.New("nil bot")
)
