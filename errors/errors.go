package errors

import (
	"errors"
)

var (
	ErrNilBot           = errors.New("bot is nil")
	ErrArtworkDeleted   = errors.New("artwork deleted")
	ErrNoPhotoInMessage = errors.New("message has no photo")
	ErrFileTooLarge     = errors.New("file too large (>20MB)")
	ErrNoAvailableFile  = errors.New("no available file")

	ErrNilArtwork          = errors.New("artwork is nil")
	ErrNotFoundArtworks    = errors.New("artworks not found")
	ErrArtworkAlreadyExist = errors.New("artwork already exists")
	ErrSourceNotSupported  = errors.New("source not supported")
	ErrIndexOOB            = errors.New("index out of bounds")
	ErrFailedToGetArtwork  = errors.New("failed to get artwork")

	ErrStorageUnkown = errors.New("unknown storage")

	ErrNotFoundArtwork = errors.New("artwork not found")
	ErrNotFoundUser    = errors.New("user not found")

	ErrLikeExists = errors.New("like exists")

	ErrSettingsNil = errors.New("settings is nil")
	ErrSettingsKey = errors.New("settings key not found")

	ErrChatIDNotSet = errors.New("chat id not set")
)
