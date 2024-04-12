package errors

import (
	"errors"
)

var (
	ErrNilBot     = errors.New("bot is nil")
	ErrNilArtwork = errors.New("artwork is nil")
)
