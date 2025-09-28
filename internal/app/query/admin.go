package query

import (
	"context"

	"github.com/krau/ManyACG/internal/common/decorator"
	"github.com/krau/ManyACG/internal/shared"
	"github.com/krau/ManyACG/pkg/objectuuid"
)

type AdminQuery struct {
	TelegramID int64
}

func NewAdminQuery(telegramID int64) *AdminQuery {
	return &AdminQuery{TelegramID: telegramID}
}

type AdminQueryResult struct {
	ID          objectuuid.ObjectUUID
	TelegramID  int64
	Permissions []shared.Permission
}

type AdminQueryHandler decorator.QueryHandler[AdminQuery, bool]

type AdminQueryRepo interface {
	FindByTelegramID(ctx context.Context, telegramID int64) (*AdminQueryResult, error)
	FindByID(ctx context.Context, id objectuuid.ObjectUUID) (*AdminQueryResult, error)
	IsAdmin(ctx context.Context, tgUserID int64) (bool, error)
}

type adminQueryHandler struct {
	queryRepo AdminQueryRepo
}

func NewAdminQueryHandler(queryRepo AdminQueryRepo) *adminQueryHandler {
	return &adminQueryHandler{queryRepo: queryRepo}
}

func (h *adminQueryHandler) Handle(ctx context.Context, q *AdminQuery) (bool, error) {
	return h.queryRepo.IsAdmin(ctx, q.TelegramID)
}
