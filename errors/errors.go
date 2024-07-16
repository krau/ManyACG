package errors

import (
	"errors"
)

var (
	ErrNilBot              = errors.New("bot is nil")
	ErrNilArtwork          = errors.New("artwork is nil")
	ErrArtworkAlreadyExist = errors.New("artwork already exists")
	ErrSourceNotSupported  = errors.New("source not supported")
	ErrArtworkDeleted      = errors.New("artwork deleted")
	ErrIndexOOB            = errors.New("index out of bounds")
)
