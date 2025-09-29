package model

import (
	"github.com/krau/ManyACG/internal/shared"
	"github.com/krau/ManyACG/pkg/objectuuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type Admin struct {
	ID          objectuuid.ObjectUUID                  `gorm:"primaryKey;type:uuid" json:"id"`
	TelegramID  int64                                  `gorm:"index" json:"telegram_id"`
	Permissions datatypes.JSONSlice[shared.Permission] `json:"permissions"`
}

func (a *Admin) BeforeCreate(tx *gorm.DB) (err error) {
	if a.ID == objectuuid.Nil {
		a.ID = objectuuid.New()
	}
	return nil
}
