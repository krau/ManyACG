package entity

import (
	"slices"

	"github.com/krau/ManyACG/internal/shared"
	"github.com/unvgo/ouid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type ApiKey struct {
	ID          ouid.OUID                              `gorm:"primaryKey;type:uuid" json:"id"`
	Key         string                                 `gorm:"type:text;not null;uniqueIndex" json:"key"`
	Quota       int                                    `gorm:"not null;default:0" json:"quota"`
	Used        int                                    `gorm:"not null;default:0" json:"used"`
	Permissions datatypes.JSONSlice[shared.Permission] `gorm:"type:json" json:"permissions"`
	Description string                                 `gorm:"type:text" json:"description"`
}

func (a *ApiKey) BeforeCreate(tx *gorm.DB) (err error) {
	if a.ID.IsZero() {
		a.ID = ouid.New()
	}
	return nil
}

func (a *ApiKey) HasPermission(p shared.Permission) bool {
	if slices.Contains(a.Permissions, shared.PermissionSudo) {
		return true
	}
	return slices.Contains(a.Permissions, p)
}

func (a *ApiKey) CanUse() bool {
	return a.Quota == 0 || a.Used < a.Quota
}
