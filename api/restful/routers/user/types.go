package user

import "ManyACG/model"

type UnauthUserResponse struct {
	ID         string `json:"id"`
	Username   string `json:"username"`
	TelegramID int64  `json:"telegram_id"`
	// TODO:
}

type UserResponseData struct {
	Username   string `json:"username"`
	Email      string `json:"email"`
	TelegramID int64  `json:"telegram_id"`

	Settings *model.UserSettings `json:"settings"`
}

type UserSettingsRequest struct {
	Language string `json:"language"`
	Theme    string `json:"theme"`
	R18      bool   `json:"r18"`
}
