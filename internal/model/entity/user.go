package entity

import (
	"time"

	"github.com/krau/ManyACG/pkg/objectuuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type User struct {
	ID         objectuuid.ObjectUUID `gorm:"primaryKey;type:uuid" json:"id"`
	Username   string                `gorm:"type:text;uniqueIndex" json:"username"`
	Password   string                `gorm:"type:text;not null" json:"password"`
	Email      *string               `gorm:"type:text;uniqueIndex" json:"email"`
	TelegramID *int64                `gorm:"type:bigint;uniqueIndex" json:"telegram_id"`
	Blocked    bool                  `gorm:"not null;default:false;index" json:"blocked"`
	UpdatedAt  time.Time             `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt  gorm.DeletedAt        `gorm:"index" json:"deleted_at"`

	Favorites []*Artwork `gorm:"many2many:user_favorites;constraint:OnDelete:CASCADE" json:"favorites"`

	Settings datatypes.JSONType[*UserSettings] `gorm:"type:json" json:"settings"`
}

type UserSettings struct {
	Language string `json:"language"`
	Theme    string `json:"theme"`
	R18      bool   `json:"r18"`
}

func (u *User) BeforeCreate(tx *gorm.DB) (err error) {
	if u.ID == objectuuid.Nil {
		u.ID = objectuuid.New()
	}
	return nil
}
