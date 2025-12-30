package entity

import (
	"time"

	"github.com/unvgo/ouid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// 只是为了兼容...
//
// V1 里或许永远用不上
type User struct {
	UpdatedAt  time.Time `gorm:"autoUpdateTime" json:"updated_at"`
	Email      *string   `gorm:"type:text;uniqueIndex" json:"email"`
	TelegramID *int64    `gorm:"type:bigint;uniqueIndex" json:"telegram_id"`

	Settings  datatypes.JSONType[*UserSettings] `gorm:"type:json" json:"settings"`
	DeletedAt gorm.DeletedAt                    `gorm:"index" json:"deleted_at"`

	Username string `gorm:"type:text;uniqueIndex" json:"username"`
	Password string `gorm:"type:text;not null" json:"password"`

	Favorites []*Artwork `gorm:"many2many:user_favorites;constraint:OnDelete:CASCADE" json:"favorites"`

	ID      ouid.OUID `gorm:"primaryKey;type:uuid" json:"id"`
	Blocked bool      `gorm:"not null;default:false;index" json:"blocked"`
}

type UserSettings struct {
	Language string `json:"language"`
	Theme    string `json:"theme"`
	R18      bool   `json:"r18"`
}

func (u *User) BeforeCreate(tx *gorm.DB) (err error) {
	if u.ID.IsZero() {
		u.ID = ouid.New()
	}
	return nil
}
