package user

type UnauthUserResponse struct {
	ID         string `json:"id"`
	Username   string `json:"username"`
	TelegramID int64  `json:"telegram_id"`
	// TODO:
}
