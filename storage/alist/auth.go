package alist

import (
	"encoding/json"
	"time"

	. "ManyACG/logger"

	"github.com/imroc/req/v3"
)

type loginRequset struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type loginResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		Token string `json:"token"`
	} `json:"data"`
}

func getJwtToken() (string, error) {
	resp, err := reqClient.R().SetBodyJsonMarshal(loginReq).Post("/api/auth/login")
	if err != nil {
		return "", err
	}
	var loginResp loginResponse
	if err := json.Unmarshal(resp.Bytes(), &loginResp); err != nil {
		return "", err
	}
	return loginResp.Data.Token, nil
}

func refreshJwtToken(client *req.Client) {
	for {
		time.Sleep(time.Hour * 24)
		token, err := getJwtToken()
		if err != nil {
			Logger.Errorf("Failed to refresh jwt token: %v", err)
			continue
		}
		client.SetCommonBearerAuthToken(token)
		Logger.Info("Refreshed Alist jwt token")
	}
}
