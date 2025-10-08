package errs

import (
	"errors"

	"github.com/krau/ManyACG/internal/infra/search"
	"github.com/krau/ManyACG/internal/infra/tagging"
	"gorm.io/gorm"
)

var (
	ErrRecordNotFound         = gorm.ErrRecordNotFound
	ErrArtworkAlreadyExist    = errors.New("artwork already exists")
	ErrArtworkDeleted         = errors.New("artwork has been deleted")
	ErrAliasAlreadyUsed       = errors.New("alias already used")
	ErrSearchEngineNotEnabled = search.ErrNotEnabled
	ErrTaggingNotEnabled      = tagging.ErrNotEnabled
)
