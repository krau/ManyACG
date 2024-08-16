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
	ErrArtworkAlreadyExist = errors.New("artwork already exists")
	ErrSourceNotSupported  = errors.New("source not supported")
	ErrIndexOOB            = errors.New("index out of bounds")

	ErrStorageNotSupported = errors.New("storage not supported")

	ErrNotFoundArtwork = errors.New("artwork not found")
	ErrNotFoundUser    = errors.New("user not found")

	ErrLikeExists = errors.New("like exists")
)
