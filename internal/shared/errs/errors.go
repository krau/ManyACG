package errs

import (
	"errors"

	"gorm.io/gorm"
)

var (
	ErrRecordNotFound      = gorm.ErrRecordNotFound
	ErrArtworkAlreadyExist = errors.New("artwork already exists")
	ErrArtworkDeleted      = errors.New("artwork has been deleted")
	ErrAliasAlreadyUsed    = errors.New("alias already used")
)
