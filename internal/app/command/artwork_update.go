package command

import (
	"context"

	"github.com/krau/ManyACG/internal/common/decorator"
	"github.com/krau/ManyACG/internal/domain/repo"
	"github.com/krau/ManyACG/internal/shared"
)

type ArtworkUpdate struct {
	ID string
	shared.ArtworkInfo
}

type UpdateArtworkHandler decorator.CommandHandler[ArtworkUpdate]

type updateArtworkHandler struct {
	txRepo repo.TransactionRepo
}

// Handle implements UpdateArtworkHandler.
func (u *updateArtworkHandler) Handle(ctx context.Context, cmd ArtworkUpdate) error {
	panic("unimplemented")
}

func NewUpdateArtworkHandler(txRepo repo.TransactionRepo) UpdateArtworkHandler {
	return &updateArtworkHandler{txRepo: txRepo}
}
