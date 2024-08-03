package errors

import (
	"errors"
)

var (
	ErrNilArtwork          = errors.New("artwork is nil")
	ErrArtworkAlreadyExist = errors.New("artwork already exists")
	ErrSourceNotSupported  = errors.New("source not supported")
	ErrArtworkDeleted      = errors.New("artwork deleted")
	ErrIndexOOB            = errors.New("index out of bounds")
)
