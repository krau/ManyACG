package entity

import (
	"slices"

	"github.com/krau/ManyACG/internal/shared"
	"github.com/unvgo/ouid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type Admin struct {
	Permissions datatypes.JSONSlice[shared.Permission] `json:"permissions"`
	TelegramID  int64                                  `gorm:"index" json:"telegram_id"`
	ID          ouid.OUID                              `gorm:"primaryKey;type:uuid" json:"id"`
}

func (a *Admin) BeforeCreate(tx *gorm.DB) (err error) {
	if a.ID.IsZero() {
		a.ID = ouid.New()
	}
	return nil
}

func (a *Admin) HasPermission(perm shared.Permission) bool {
	return slices.Contains(a.Permissions, perm)
}
