package alist

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/krau/ManyACG/common"
	"github.com/krau/ManyACG/config"

	"github.com/imroc/req/v3"
)

func getJwtToken() (string, error) {
	resp, err := reqClient.R().SetBodyJsonMarshal(loginReq).Post("/api/auth/login")
	if err != nil {
		return "", err
	}
	var loginResp loginResponse
	if err := json.Unmarshal(resp.Bytes(), &loginResp); err != nil {
		return "", err
	}
	if loginResp.Code != http.StatusOK {
		return "", fmt.Errorf("%w: %s", ErrAlistLoginFailed, loginResp.Message)
	}
	return loginResp.Data.Token, nil
}

func refreshJwtToken(client *req.Client) {
	for {
		time.Sleep(time.Duration(config.Cfg.Storage.Alist.TokenExpire) * time.Second)
		token, err := getJwtToken()
		if err != nil {
			common.Logger.Errorf("Failed to refresh jwt token: %v", err)
			continue
		}
		client.SetCommonHeader("Authorization", token)
		common.Logger.Info("Refreshed Alist jwt token")
	}
}
